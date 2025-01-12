package controller

import (
	"github.com/pkg/errors"
	"net/http"

	"github.com/finkabaj/squid/back/internal/service"
	"github.com/finkabaj/squid/back/internal/utils"

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

func register(w http.ResponseWriter, r *http.Request) {
	register, ok := middleware.JsonFromContext(r.Context()).(types.RegisterUser)

	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get register user from context")))
		return
	}

	user, err := service.Register(&register)

	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, user); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal user"))
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	login, ok := middleware.JsonFromContext(r.Context()).(types.Login)

	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get login user from context")))
		return
	}

	user, err := service.Login(&login)

	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, user); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal user"))
	}
}

func refreshToken(w http.ResponseWriter, r *http.Request) {}

func checkToken(w http.ResponseWriter, r *http.Request) {}
