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
		r.With(middleware.ValidateJson[types.RefreshTokenRequest]()).Post("/refresh", refreshToken)
		r.With(middleware.ValidateJWT).Post("/check", checkToken)
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

	if err = utils.MarshalBody(w, http.StatusCreated, user); err != nil {
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

func refreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, ok := middleware.JsonFromContext(r.Context()).(types.RefreshTokenRequest)

	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get refresh token from context")))
		return
	}

	tokens, err := service.RefreshToken(&refreshToken.RefreshToken)

	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, tokens); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal user"))
	}
}

func checkToken(w http.ResponseWriter, r *http.Request) {
	if err := utils.MarshalBody(w, http.StatusOK, utils.OkResponse{
		Message: "Token is valid",
	}); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal ok response"))
	}
}
