package controller

import (
	"net/http"

	"github.com/pkg/errors"

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
		r.With(middleware.ValidateJWT, middleware.ValidateJson[types.UpdateUser]()).Patch("/user", updateUser)
		r.With(middleware.ValidateJWT, middleware.ValidateJson[types.UpdatePassword]()).Patch("/password", updatePassword)
		r.With(middleware.ValidateJWT).Get("/user/{id}", getUser)
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

func updateUser(w http.ResponseWriter, r *http.Request) {
	updatedUser, ok := middleware.JsonFromContext(r.Context()).(types.UpdateUser)
	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get update user from context")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	busser, err := service.UpdateUser(&user, &updatedUser)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, busser); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal user"))
	}
}

func getUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	user, err := service.GetUserById(&userID)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, user); err != nil {
		utils.HandleError(w, errors.New("Failed to marshal user"))
	}
}

func updatePassword(w http.ResponseWriter, r *http.Request) {
	newPassword, ok := middleware.JsonFromContext(r.Context()).(types.UpdatePassword)
	if !ok {
		utils.HandleError(w, utils.NewInternalError(errors.New("Failed to get update user from context")))
		return
	}

	user := middleware.UserFromContext(r.Context())

	busser, err := service.UpdateUserPassword(&user, &newPassword)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if err = utils.MarshalBody(w, http.StatusOK, busser); err != nil {
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
