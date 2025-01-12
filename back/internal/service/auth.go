package service

import (
	"context"

	"github.com/finkabaj/squid/back/internal/repository"
	"github.com/finkabaj/squid/back/internal/types"
	"github.com/finkabaj/squid/back/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

func Register(user *types.RegisterUser) (types.User, error) {
	if user == nil {
		return types.User{}, utils.NewBadRequestError(errors.New("user is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := repository.GetUser(ctx, nil, &user.Email)

	if !errors.Is(err, pgx.ErrNoRows) {
		return types.User{}, utils.AppError{
			Type: utils.ErrorType{
				Status:  http.StatusConflict,
				Message: "User already exists",
			},
			OriginalError: err,
		}
	}

	passwordHash, err := utils.HashPassword(&user.Password)

	if err != nil {
		return types.User{}, utils.NewInternalError(err)
	}

	id := uuid.New().String()

	newUser, err := repository.CreateUser(ctx, &id, &passwordHash, user)

	if err != nil {
		return types.User{}, utils.NewInternalError(err)
	}

	return newUser, nil
}

func Login(login *types.Login) (types.User, error) {
	if login == nil {
		return types.User{}, utils.NewBadRequestError(errors.New("login is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := repository.GetUser(ctx, nil, &login.Email)

	if errors.Is(errors.Cause(err), pgx.ErrNoRows) {
		return types.User{}, utils.AppError{
			Type: utils.ErrorType{
				Status:  http.StatusUnauthorized,
				Message: "Invalid email or password",
			},
			OriginalError: err,
		}
	} else if err != nil {
		return types.User{}, utils.NewInternalError(err)
	}

	if !utils.CheckPasswordHash(&login.Password, &user.PasswordHash) {
		return types.User{}, utils.AppError{
			Type: utils.ErrorType{
				Status:  http.StatusUnauthorized,
				Message: "Invalid email or password",
			},
			OriginalError: err,
		}
	}

	return user, nil
}
