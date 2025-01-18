package controller

import (
	"github.com/finkabaj/squid/back/internal/service"
	"github.com/pkg/errors"

	"net/http"
	"net/url"

	"github.com/finkabaj/squid/back/internal/middleware"
	"github.com/finkabaj/squid/back/internal/types"
	"github.com/finkabaj/squid/back/internal/utils"
	"github.com/go-chi/chi/v5"
)

var kanbanControllerInitialized = false

func RegisterKanbanRoutes(r *chi.Mux) {
	if !kanbanControllerInitialized {
		return
	}

	r.Route("/kanban", func(r chi.Router) {
		r.With(middleware.ValidateJWT).With(middleware.ValidateJson[types.CreateProject]()).Post("/project", createProject)
		r.With(middleware.ValidateJWT).Get("/project", getProject)
	})

	kanbanControllerInitialized = true
}

func createProject(w http.ResponseWriter, r *http.Request) {
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
	}
}

func getProject(w http.ResponseWriter, r *http.Request) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	id := values.Get("id")
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
