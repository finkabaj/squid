package controller

import (
	"github.com/finkabaj/squid/back/internal/service"
	"github.com/finkabaj/squid/back/internal/websocket"
	"github.com/pkg/errors"

	"net/http"

	"github.com/finkabaj/squid/back/internal/middleware"
	"github.com/finkabaj/squid/back/internal/types"
	"github.com/finkabaj/squid/back/internal/utils"
	"github.com/go-chi/chi/v5"
)

var kanbanControllerInitialized = false

type KanbanController struct {
	WSServer *websocket.Server
}

func NewKanbanController(wsServer *websocket.Server) *KanbanController {
	return &KanbanController{
		WSServer: wsServer,
	}
}

func (c *KanbanController) RegisterKanbanRoutes(r *chi.Mux) {
	if kanbanControllerInitialized {
		return
	}

	r.Route("/kanban", func(r chi.Router) {
		r.Use(middleware.ValidateJWT)

		r.With(middleware.ValidateJson[types.CreateProject]()).Post("/project", c.createProject)
		r.Get("/project/{project_id}", c.getProject)
		r.With(middleware.ValidateJson[types.UpdateProject]()).Patch("/project/{project_id}", c.updateProject)
		r.Delete("/project/{project_id}", c.deleteProject)
		r.Get("/projects", c.getProjects)
		r.Get("/project/users/{project_id}", c.getProjectUsers)

		r.With(middleware.ValidateJson[types.CreateKanbanColumn]()).Post("/column", c.createColumn)
		r.Get("/column/{column_id}", c.getColumn)
		r.With(middleware.ValidateJson[types.UpdateKanbanColumn]()).Patch("/column/{column_id}", c.updateColumn)
		r.Delete("/column/{column_id}", c.deleteColumn)
		r.Get("/columns/{project_id}", c.getColumns)

		r.With(middleware.ValidateJson[types.CreateKanbanColumnLabel]()).Post("/column/label", c.createColumnLabel)
		r.Delete("/column/label/{label_id}", c.deleteColumnLabel)
		r.With(middleware.ValidateJson[types.UpdateKanbanColumnLabel]()).Patch("/column/label/{label_id}", c.updateColumnLabel)
		r.Get("/column/labels/{project_id}", c.getColumnLabels)

		r.With(middleware.ValidateJson[types.CreateKanbanRow]()).Post("/row", c.createRow)
		r.With(middleware.ValidateJson[types.UpdateKanbanRow]()).Patch("/row/{row_id}", c.updateRows)
		r.Delete("/row/{row_id}", c.deleteRow)
		r.Get("/rows/{column_id}", c.getRows)

		r.With(middleware.ValidateJson[types.CreateKanbanRowLabel]()).Post("/row/label", c.createRowLabel)
		r.Delete("/row/label/{label_id}", c.deleteRowLabel)
		r.With(middleware.ValidateJson[types.UpdateKanbanRowLabel]()).Patch("/row/label/{label_id}", c.updateRowLabel)
		r.Get("/row/labels/{project_id}", c.getRowLabels)

		r.Post("/checklist/{row_id}", c.createChecklist)
		r.Delete("/checklist/{checklist_id}", c.deleteChecklist)

		r.With(middleware.ValidateJson[types.CreatePoint]()).Post("/checklist/point", c.createPoint)
		r.With(middleware.ValidateJson[types.UpdatePoint]()).Patch("/checklist/point/{point_id}", c.updatePoint)
		r.Patch("/checklist/point/status/{point_id}", c.updatePointStatus)
		r.Delete("/checklist/point/{point_id}", c.deletePoint)
		r.Get("/checklist/points/{checklist_id}", c.getPoints)

		r.Patch("/comment/can_comment/{comment_section_id}", c.updateCanComment)

		r.With(middleware.ValidateJson[types.CreateComment]()).Post("/comment", c.createComment)
		r.Delete("/comment/{comment_id}", c.deleteComment)
		r.Get("/comments/{comment_section_id}", c.getComments)
	})

	kanbanControllerInitialized = true
}

func (c *KanbanController) getProjects(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	projects, err := service.GetProjects(&user.ID)

	if err != nil {
		utils.HandleError(w, err)
	}

	if err = utils.MarshalBody(w, http.StatusOK, projects); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal projects"))
		return
	}
}

func (c *KanbanController) getProjectUsers(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	projectID := chi.URLParam(r, "project_id")
	if projectID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("project id is required")))
		return
	}

	projectUsers, err := service.GetProjectUsers(&user.ID, &projectID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, projectUsers); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal project users"))
		return
	}
}

func (c *KanbanController) createProject(w http.ResponseWriter, r *http.Request) {
	projectData, ok := middleware.JsonFromContext(r.Context()).(types.CreateProject)
	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get project info from context")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	newProject, err := service.CreateProject(&user, &projectData)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusCreated, newProject); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal project"))
		return
	}

	projectUsers := append(newProject.AdminIDs, newProject.MembersIDs...)
	projectUsers = append(projectUsers, newProject.CreatorID)

	c.WSServer.BroadcastToProject(newProject.ID, websocket.ProjectCreatedEvent, "project created", newProject, projectUsers)
}

func (c *KanbanController) getProject(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	if projectID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("project id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	project, err := service.GetProject(&user.ID, &projectID)

	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, project); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal project"))
	}
}

func (c *KanbanController) updateProject(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	if projectID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("project id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	updateProject, ok := middleware.JsonFromContext(r.Context()).(types.UpdateProject)

	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("error getting updateProject from contex")))
		return
	}

	project, err := service.UpdateProject(&projectID, &user, &updateProject)

	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, project); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal project"))
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.ProjectUpdatedEvent, "project updated", project, projectUsers)
}

func (c *KanbanController) deleteProject(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	if projectID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("project id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	project, err := service.DeleteProject(&user, &projectID)

	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, utils.OkResponse{Message: "project deleted succesfully"}); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal OkResponse"))
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.ProjectDeletedEvent, "project deleted", nil, projectUsers)
}

func (c *KanbanController) createColumn(w http.ResponseWriter, r *http.Request) {
	columnData, ok := middleware.JsonFromContext(r.Context()).(types.CreateKanbanColumn)
	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get column info from context")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	newColumn, columns, project, err := service.CreateColumn(&user, &columnData)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusCreated, map[string]any{
		"new_column": newColumn,
		"columns":    columns,
	}); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal column"))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(
		newColumn.ProjectID,
		websocket.KanbanColumnCreatedEvent,
		"kanban column created",
		map[string]any{
			"new_column": newColumn,
			"columns":    columns,
		},
		projectUsers)
}

func (c *KanbanController) getColumn(w http.ResponseWriter, r *http.Request) {
	columnID := chi.URLParam(r, "column_id")
	if columnID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("column id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	column, err := service.GetColumn(&columnID, &user.ID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, column); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal column"))
		return
	}
}

func (c *KanbanController) updateColumn(w http.ResponseWriter, r *http.Request) {
	columnID := chi.URLParam(r, "column_id")
	if columnID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("column id is required")))
		return
	}

	columnData, ok := middleware.JsonFromContext(r.Context()).(types.UpdateKanbanColumn)
	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get column info from context")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	column, columns, project, err := service.UpdateColumn(&columnID, &user, &columnData)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, map[string]any{
		"updated_column": column,
		"columns":        columns,
	}); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal column"))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanColumnUpdatedEvent, "kanban column updated", map[string]any{
		"updated_column": column,
		"columns":        columns,
	}, projectUsers)
}

func (c *KanbanController) deleteColumn(w http.ResponseWriter, r *http.Request) {
	columnID := chi.URLParam(r, "column_id")
	if columnID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("column id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	columns, project, err := service.DeleteColumn(&columnID, &user)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, columns); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal okResponse"))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanColumnDeletedEvent, "kanban column deleted", columns, projectUsers)
}

func (c *KanbanController) getColumns(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	if projectID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("project id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	columns, err := service.GetColumns(&projectID, &user.ID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, columns); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal colums"))
		return
	}
}

func (c *KanbanController) createColumnLabel(w http.ResponseWriter, r *http.Request) {
	createColumnLabel, ok := middleware.JsonFromContext(r.Context()).(types.CreateKanbanColumnLabel)
	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get createColumnLabel from context")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	label, project, err := service.CreateColumnLabel(&user.ID, &createColumnLabel)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusCreated, label); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal label")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanColumnLabelCreatedEvent, "kanban column label created", label, projectUsers)
}

func (c *KanbanController) deleteColumnLabel(w http.ResponseWriter, r *http.Request) {
	labelID := chi.URLParam(r, "label_id")
	if labelID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("column label id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	project, err := service.DeleteColumnLabel(&user.ID, &labelID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, utils.OkResponse{Message: "column label deleted"}); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal okResponse")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanColumnLabelDeletedEvent, "kanban column label deleted", nil, projectUsers)
}

func (c *KanbanController) updateColumnLabel(w http.ResponseWriter, r *http.Request) {
	updateLabel, ok := middleware.JsonFromContext(r.Context()).(types.UpdateKanbanColumnLabel)
	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get updateColumnLabel from context")))
		return
	}

	labelID := chi.URLParam(r, "label_id")
	if labelID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("column label id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	label, project, err := service.UpdateColumnLabel(&user.ID, &labelID, &updateLabel)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, label); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal label")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanColumnLabelUpdatedEvent, "kanban column label updated", label, projectUsers)
}

func (c *KanbanController) getColumnLabels(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	if projectID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("project id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	labels, err := service.GetColumnLabels(&user.ID, &projectID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, labels); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal labels")))
		return
	}
}

func (c *KanbanController) createRow(w http.ResponseWriter, r *http.Request) {
	createRow, ok := middleware.JsonFromContext(r.Context()).(types.CreateKanbanRow)
	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get createRow from context")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	row, rows, project, err := service.CreateRow(&user.ID, &createRow)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusCreated, map[string]any{"new_row": row, "rows": rows}); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal row")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(
		project.ID,
		websocket.KanbanRowCreatedEvent,
		"kanban column label updated",
		map[string]any{
			"new_row": row,
			"rows":    rows,
		},
		projectUsers,
	)
}

func (c *KanbanController) updateRows(w http.ResponseWriter, r *http.Request) {
	updateRow, ok := middleware.JsonFromContext(r.Context()).(types.UpdateKanbanRow)
	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get updateRow from context")))
		return
	}

	rowID := chi.URLParam(r, "row_id")
	if rowID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("row id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	row, rows, project, err := service.UpdateRow(&user.ID, &rowID, &updateRow)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, map[string]any{"updated_row": row, "rows": rows}); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal label")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(
		project.ID,
		websocket.KanbanColumnLabelUpdatedEvent,
		"kanban column label updated",
		map[string]any{
			"updated_row": row,
			"rows":        rows,
		},
		projectUsers,
	)
}

func (c *KanbanController) deleteRow(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	rowID := chi.URLParam(r, "row_id")
	if rowID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("row id is required")))
		return
	}

	rows, project, err := service.DeleteRow(&user.ID, &rowID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, rows); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal okResponse"))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanRowDeletedEvent, "kanban row deleted", rows, projectUsers)
}

func (c *KanbanController) getRows(w http.ResponseWriter, r *http.Request) {
	columnID := chi.URLParam(r, "column_id")
	if columnID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("column id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	rows, err := service.GetRows(&user.ID, &columnID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, rows); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal rows")))
		return
	}
}

func (c *KanbanController) createRowLabel(w http.ResponseWriter, r *http.Request) {
	createRowLabel, ok := middleware.JsonFromContext(r.Context()).(types.CreateKanbanRowLabel)
	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get createColumnLabel from context")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	label, project, err := service.CreateRowLabel(&user.ID, &createRowLabel)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusCreated, label); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal label")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanRowLabelCreatedEvent, "kanban row label created", label, projectUsers)
}

func (c *KanbanController) deleteRowLabel(w http.ResponseWriter, r *http.Request) {
	labelID := chi.URLParam(r, "label_id")
	if labelID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("row label id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	project, err := service.DeleteRowLabel(&user.ID, &labelID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, utils.OkResponse{Message: "row label deleted"}); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal okResponse")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanRowLabelDeletedEvent, "kanban row label deleted", nil, projectUsers)
}

func (c *KanbanController) updateRowLabel(w http.ResponseWriter, r *http.Request) {
	updateLabel, ok := middleware.JsonFromContext(r.Context()).(types.UpdateKanbanRowLabel)
	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get updateColumnLabel from context")))
		return
	}

	labelID := chi.URLParam(r, "label_id")
	if labelID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("row label id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	label, project, err := service.UpdateRowLabel(&user.ID, &labelID, &updateLabel)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, label); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal label")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanRowLabelUpdatedEvent, "kanban row label updated", label, projectUsers)
}

func (c *KanbanController) getRowLabels(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	if projectID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("project id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	labels, err := service.GetRowLabels(&user.ID, &projectID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, labels); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal labels")))
		return
	}
}

func (c *KanbanController) createChecklist(w http.ResponseWriter, r *http.Request) {
	rowID := chi.URLParam(r, "row_id")
	if rowID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("row id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	checklist, project, err := service.CreateChecklist(&user.ID, &rowID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, checklist); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal checklist")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanChecklistCreatedEvent, "checklist created", checklist, projectUsers)
}

func (c *KanbanController) deleteChecklist(w http.ResponseWriter, r *http.Request) {
	checklistID := chi.URLParam(r, "checklist_id")
	if checklistID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("checklist id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	project, err := service.DeleteChecklist(&user.ID, &checklistID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, utils.OkResponse{Message: "checklist deleted"}); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal okResponse")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanChecklistDeletedEvent, "checklist deleted", nil, projectUsers)
}

func (c *KanbanController) createPoint(w http.ResponseWriter, r *http.Request) {
	createPoint, ok := middleware.JsonFromContext(r.Context()).(types.CreatePoint)
	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get createPoint from context")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	point, project, err := service.CreatePoint(&user.ID, &createPoint)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, point); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal point")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanPointCreatedEvent, "point created", point, projectUsers)
}

func (c *KanbanController) updatePoint(w http.ResponseWriter, r *http.Request) {
	updatePoint, ok := middleware.JsonFromContext(r.Context()).(types.UpdatePoint)
	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get updatePoint from context")))
		return
	}

	pointID := chi.URLParam(r, "point_id")
	if pointID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("point id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	point, project, err := service.UpdatePoint(&user.ID, &pointID, &updatePoint)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, point); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal point")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanPointUpdatedEvent, "point updated", point, projectUsers)
}

func (c *KanbanController) updatePointStatus(w http.ResponseWriter, r *http.Request) {
	pointID := chi.URLParam(r, "point_id")
	if pointID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("point id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	point, project, err := service.UpdatePointStatus(&user.ID, &pointID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, point); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal point")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanPointUpdatedEvent, "point updated", point, projectUsers)
}

func (c *KanbanController) deletePoint(w http.ResponseWriter, r *http.Request) {
	pointID := chi.URLParam(r, "point_id")
	if pointID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("point id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	project, err := service.DeletePoint(&user.ID, &pointID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, utils.OkResponse{Message: "point deleted"}); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal okResponse")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanPointDeletedEvent, "point deleted", nil, projectUsers)
}

func (c *KanbanController) getPoints(w http.ResponseWriter, r *http.Request) {
	checklistID := chi.URLParam(r, "checklist_id")
	if checklistID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("checklist id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	points, err := service.GetPoints(&user.ID, &checklistID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, points); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal points")))
		return
	}
}

func (c *KanbanController) updateCanComment(w http.ResponseWriter, r *http.Request) {
	commentSectionID := chi.URLParam(r, "comment_section_id")
	if commentSectionID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("comment section id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	canComment, project, err := service.UpdateCanComment(&user.ID, &commentSectionID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, canComment); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal project")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanCanCommentEvent, "can comment updated", canComment, projectUsers)
}

func (c *KanbanController) createComment(w http.ResponseWriter, r *http.Request) {
	createComment, ok := middleware.JsonFromContext(r.Context()).(types.CreateComment)
	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get createComment from context")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	comment, project, err := service.CreateComment(&user.ID, &createComment)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, comment); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal comment")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanCommentCreatedEvent, "comment created", comment, projectUsers)
}

func (c *KanbanController) deleteComment(w http.ResponseWriter, r *http.Request) {
	commentID := chi.URLParam(r, "comment_id")
	if commentID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("comment id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	project, err := service.DeleteComment(&user.ID, &commentID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, utils.OkResponse{Message: "comment deleted"}); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal project")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanCommendDeletedEvent, "comment deleted", nil, projectUsers)
}

func (c *KanbanController) getComments(w http.ResponseWriter, r *http.Request) {
	commentSectionID := chi.URLParam(r, "comment_section_id")
	if commentSectionID == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("comment section id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	comments, err := service.GetComments(&user.ID, &commentSectionID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := utils.MarshalBody(w, http.StatusOK, comments); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal comments")))
		return
	}
}
