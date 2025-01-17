package controller

import (
	"github.com/finkabaj/squid/back/internal/middleware"
	"github.com/finkabaj/squid/back/internal/types"
	"github.com/go-chi/chi/v5"
	"net/http"
)

var kanbanControllerInitialized = false

func RegisterKanbanRoutes(r *chi.Mux) {
	if !kanbanControllerInitialized {
		return
	}

	r.Route("/", func(r chi.Router) {
		r.With(middleware.ValidateJWT).With(middleware.ValidateJson[types.KanbanColumn]())
	})

	kanbanControllerInitialized = true
}

func createList(w http.ResponseWriter, r *http.Request) {

}
