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
		r.With(middleware.ValidateJWT).Get("/project/{id}", c.getProject)
		r.With(middleware.ValidateJWT, middleware.ValidateJson[types.UpdateProject]()).Patch("/project/{id}", c.updateProject)
		r.With(middleware.ValidateJWT).Delete("/project/{id}", c.deleteProject)
		r.With(middleware.ValidateJWT).Get("/projects", c.getProjects)

		r.With(middleware.ValidateJWT, middleware.ValidateJson[types.CreateKanbanColumn]()).Post("/column", c.createColumn)
		r.With(middleware.ValidateJWT).Get("/column/{id}", c.getColumn)
		r.With(middleware.ValidateJWT, middleware.ValidateJson[types.UpdateKanbanColumn]()).Patch("/column/{id}", c.updateColumn)
		r.With(middleware.ValidateJWT).Delete("/column/{id}", c.deleteColumn)
		r.With(middleware.ValidateJWT).Get("/columns/{id}", c.getColumns)
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
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("project id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	project, err := service.GetProject(&user.ID, &id)

	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, project); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal project"))
	}
}

func (c *KanbanController) updateProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("project id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	updateProject, ok := middleware.JsonFromContext(r.Context()).(types.UpdateProject)

	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("error getting updateProject from contex")))
		return
	}

	project, err := service.UpdateProject(&id, &user, &updateProject)

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
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("project id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	project, err := service.DeleteProject(&user, &id)

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

	newColumn, project, err := service.CreateColumn(&user, &columnData)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusCreated, newColumn); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal column"))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(newColumn.ProjectID, websocket.KanbanColumnCreatedEvent, "kanban column created", newColumn, projectUsers)
}

func (c *KanbanController) getColumn(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("column id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	column, err := service.GetColumn(&id, &user.ID)
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
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("column id is required")))
		return
	}

	columnData, ok := middleware.JsonFromContext(r.Context()).(types.UpdateKanbanColumn)
	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get column info from context")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	column, project, err := service.UpdateColumn(&id, &user, &columnData)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, column); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal column"))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanColumnUpdatedEvent, "kanban column updated", column, projectUsers)
}

func (c *KanbanController) deleteColumn(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("column id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	project, err := service.DeleteColumn(&id, &user)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, utils.OkResponse{Message: "kanban column deleted succesfully"}); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal okResponse"))
		return
	}

	projectUsers := append(project.AdminIDs, project.MembersIDs...)
	projectUsers = append(projectUsers, project.CreatorID)

	c.WSServer.BroadcastToProject(project.ID, websocket.KanbanColumnDeletedEvent, "kanban column deleted", nil, projectUsers)
}

func (c *KanbanController) getColumns(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.HandleError(w, utils.NewBadRequestError(errors.New("project id is required")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	columns, err := service.GetColumns(&id, &user.ID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, columns); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal colums"))
		return
	}
}
