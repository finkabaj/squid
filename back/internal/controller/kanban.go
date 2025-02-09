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
		r.With(middleware.ValidateJWT, middleware.ValidateJson[types.CreateProject]()).Post("/project", c.createProject)
		r.With(middleware.ValidateJWT).Get("/project/{project_id}", c.getProject)
		r.With(middleware.ValidateJWT, middleware.ValidateJson[types.UpdateProject]()).Patch("/project/{project_id}", c.updateProject)
		r.With(middleware.ValidateJWT).Delete("/project/{project_id}", c.deleteProject)
		r.With(middleware.ValidateJWT).Get("/projects", c.getProjects)

		r.With(middleware.ValidateJWT, middleware.ValidateJson[types.CreateKanbanColumn]()).Post("/column", c.createColumn)
		r.With(middleware.ValidateJWT).Get("/column/{column_id}", c.getColumn)
		r.With(middleware.ValidateJWT, middleware.ValidateJson[types.UpdateKanbanColumn]()).Patch("/column/{column_id}", c.updateColumn)
		r.With(middleware.ValidateJWT).Delete("/column/{column_id}", c.deleteColumn)
		r.With(middleware.ValidateJWT).Get("/columns/{project_id}", c.getColumns)

		r.With(middleware.ValidateJWT, middleware.ValidateJson[types.CreateKanbanColumnLabel]()).Post("/column/label", c.createColumnLabel)
		r.With(middleware.ValidateJWT).Delete("/column/label/{label_id}", c.deleteColumnLabel)
		r.With(middleware.ValidateJWT, middleware.ValidateJson[types.UpdateKanbanColumnLabel]()).Patch("/column/label/{label_id}", c.updateColumnLabel)
		r.With(middleware.ValidateJWT).Get("/column/labels/{project_id}", c.getColumnLabels)

		r.With(middleware.ValidateJWT, middleware.ValidateJson[types.CreateKanbanRow]()).Post("/row", c.createRow)
		r.With(middleware.ValidateJWT, middleware.ValidateJson[types.UpdateKanbanRow]()).Patch("/row/{row_id}", c.updateRows)
		r.With(middleware.ValidateJWT).Delete("/row/{row_id}", c.deleteRow)
		r.With(middleware.ValidateJWT).Get("/rows/{column_id}", c.getRows)
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

func (c *KanbanController) createProject(w http.ResponseWriter, r *http.Request) {
	projectData, ok := middleware.JsonFromContext(r.Context()).(types.CreateProject)
	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get project info from context")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	newProject, err := service.CreateProject(&user.ID, &projectData)
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

	if err = utils.MarshalBody(w, http.StatusCreated, map[string]interface{}{
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
		map[string]interface{}{
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

	if err = utils.MarshalBody(w, http.StatusOK, map[string]interface{}{
		"updated_column": column,
		"columns":        columns,
	}); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal column"))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanColumnUpdatedEvent, "kanban column updated", map[string]interface{}{
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

	if err := utils.MarshalBody(w, http.StatusCreated, map[string]interface{}{"new_row": row, "rows": rows}); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal row")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(
		project.ID,
		websocket.KanbanRowCreatedEvent,
		"kanban column label updated",
		map[string]interface{}{
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

	if err := utils.MarshalBody(w, http.StatusOK, map[string]interface{}{"updated_row": row, "rows": rows}); err != nil {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to marshal label")))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(
		project.ID,
		websocket.KanbanColumnLabelUpdatedEvent,
		"kanban column label updated",
		map[string]interface{}{
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
