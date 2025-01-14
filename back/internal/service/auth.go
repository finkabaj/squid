package service

import (
	"context"
	"fmt"

	"github.com/finkabaj/squid/back/internal/config"
	"github.com/golang-jwt/jwt/v5"

	"net/http"
	"time"

	"github.com/finkabaj/squid/back/internal/repository"
	"github.com/finkabaj/squid/back/internal/types"
	"github.com/finkabaj/squid/back/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

func Register(user *types.RegisterUser) (types.AuthUser, error) {
	if user == nil {
		return types.AuthUser{}, utils.NewBadRequestError(errors.New("user is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := repository.GetUser(ctx, nil, &user.Email)

	if !errors.Is(err, pgx.ErrNoRows) {
		return types.AuthUser{}, utils.AppError{
			Type: utils.ErrorType{
				Status:  http.StatusConflict,
				Message: "User already exists",
			},
			OriginalError: err,
		}
	}

	passwordHash, err := utils.HashPassword(&user.Password)

	if err != nil {
		return types.AuthUser{}, utils.NewInternalError(err)
	}

	id := uuid.New().String()

	newUser, err := repository.CreateUser(ctx, &id, &passwordHash, user)

	if err != nil {
		return types.AuthUser{}, utils.NewInternalError(err)
	}

	id = uuid.New().String()

	expAt := time.Now().Add(time.Hour * time.Duration(config.Data.RefreshTokenExpHours))

	refreshToken, err := repository.CreateRefreshToken(ctx, &id, &newUser.ID, &expAt)

	if err != nil {
		return types.AuthUser{}, utils.NewInternalError(err)
	}

	jwtPair, err := utils.CreateJWTPair(&newUser, &refreshToken)

	if err != nil {
		return types.AuthUser{}, utils.NewInternalError(err)
	}

	return types.AuthUser{
		User: newUser,
		TokenPair: types.TokenPair{
			AccessToken:  jwtPair["accessToken"],
			RefreshToken: jwtPair["refreshToken"],
		},
	}, nil
}

func Login(login *types.Login) (types.AuthUser, error) {
	if login == nil {
		return types.AuthUser{}, utils.NewBadRequestError(errors.New("login is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := repository.GetUser(ctx, nil, &login.Email)

	if errors.Is(errors.Cause(err), pgx.ErrNoRows) {
		return types.AuthUser{}, utils.AppError{
			Type: utils.ErrorType{
				Status:  http.StatusUnauthorized,
				Message: "Invalid email or password",
			},
			OriginalError: err,
		}
	} else if err != nil {
		return types.AuthUser{}, utils.NewInternalError(err)
	}

	if !utils.CheckPasswordHash(&login.Password, &user.PasswordHash) {
		return types.AuthUser{}, utils.AppError{
			Type: utils.ErrorType{
				Status:  http.StatusUnauthorized,
				Message: "Invalid email or password",
			},
			OriginalError: err,
		}
	}

	err = repository.DeleteRefreshToken(ctx, &user.ID)

	if err != nil && err.Error() != "no rows were deleted" {
		return types.AuthUser{}, utils.NewInternalError(err)
	}

	id := uuid.New().String()

	exp := time.Now().Add(time.Hour * time.Duration(config.Data.RefreshTokenExpHours))

	refreshToken, err := repository.CreateRefreshToken(ctx, &id, &user.ID, &exp)

	if err != nil {
		return types.AuthUser{}, utils.NewInternalError(err)
	}

	jwtPair, err := utils.CreateJWTPair(&user, &refreshToken)

	if err != nil {
		return types.AuthUser{}, utils.NewInternalError(err)
	}

	return types.AuthUser{
		User: user,
		TokenPair: types.TokenPair{
			AccessToken:  jwtPair["accessToken"],
			RefreshToken: jwtPair["refreshToken"],
		},
	}, nil
}

func invalidRefreshToken(err error) utils.AppError {
	return utils.AppError{
		Type: utils.ErrorType{
			Status:  http.StatusUnauthorized,
			Message: "Invalid refresh token",
		},
		OriginalError: err,
	}
}

func RefreshToken(refreshTokenStr *string) (types.TokenPair, error) {
	if refreshTokenStr == nil {
		return types.TokenPair{}, utils.NewBadRequestError(errors.New("refresh token is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token, err := jwt.Parse(*refreshTokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return config.Data.JWTSecret, nil
	})

	if err != nil {
		return types.TokenPair{}, invalidRefreshToken(err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok || !token.Valid {
		return types.TokenPair{}, invalidRefreshToken(err)
	}

	exp, ok := claims["expires_at"].(float64)

	if !ok || time.Now().Unix() > int64(exp) {
		return types.TokenPair{}, invalidRefreshToken(errors.New("refresh token expired"))
	}

	id, ok := claims["id"].(string)

	if !ok {
		return types.TokenPair{}, invalidRefreshToken(err)
	}

	refreshToken, err := repository.GetRefreshToken(ctx, &id)

	if errors.Is(err, pgx.ErrNoRows) {
		return types.TokenPair{}, invalidRefreshToken(err)
	} else if err != nil {
		return types.TokenPair{}, utils.NewInternalError(err)
	}

	user, err := repository.GetUser(ctx, &refreshToken.UserID, nil)

	if errors.Is(err, pgx.ErrNoRows) {
		return types.TokenPair{}, invalidRefreshToken(err)
	} else if err != nil {
		return types.TokenPair{}, utils.NewInternalError(err)
	}

	err = repository.DeleteRefreshToken(ctx, &refreshToken.UserID)
	if err != nil {
		return types.TokenPair{}, utils.NewInternalError(err)
	}

	newTokenID := uuid.New().String()
	expAt := time.Now().Add(time.Hour * time.Duration(config.Data.RefreshTokenExpHours))

	newRefreshToken, err := repository.CreateRefreshToken(ctx, &newTokenID, &user.ID, &expAt)

	if err != nil {
		return types.TokenPair{}, utils.NewInternalError(err)
	}

	jwtPair, err := utils.CreateJWTPair(&user, &newRefreshToken)

	if err != nil {
		return types.TokenPair{}, utils.NewInternalError(err)
	}

	return types.TokenPair{
		AccessToken:  jwtPair["accessToken"],
		RefreshToken: jwtPair["refreshToken"],
	}, nil
}
