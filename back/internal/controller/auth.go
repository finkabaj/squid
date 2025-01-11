package controller

import (
	"net/http"

	"github.com/finkabaj/squid/back/internal/middleware"
	"github.com/finkabaj/squid/back/internal/types"
	"github.com/go-chi/chi/v5"
)

var authControllerInitialized bool

func RegisterAuthRoutes(r *chi.Mux) {
	if authControllerInitialized {
		return
	}

	r.Route("/auth", func(r chi.Router) {
		r.With(middleware.ValidateJson[types.Login]()).Post("/login", login)
		r.With(middleware.ValidateJson[types.RegisterUser]()).Post("/register", register)
		r.With(middleware.ValidateJWTRefresh).Post("/refresh", refreshToken)
		r.With(middleware.ValidateJWTAuth).Post("/check", checkToken)
	})

	authControllerInitialized = true
}

func register(w http.ResponseWriter, r *http.Request) {}

func login(w http.ResponseWriter, r *http.Request) {}

func refreshToken(w http.ResponseWriter, r *http.Request) {}

func checkToken(w http.ResponseWriter, r *http.Request) {}
