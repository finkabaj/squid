package controller

import (
	"github.com/finkabaj/squid/back/internal/service"
	"github.com/finkabaj/squid/back/internal/websocket"
	"github.com/pkg/errors"

	"github.com/finkabaj/squid/back/internal/middleware"
	"github.com/finkabaj/squid/back/internal/types"
	"github.com/finkabaj/squid/back/internal/utils"
	"github.com/go-chi/chi/v5"
	"net/http"
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
	})

	kanbanControllerInitialized = true
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
