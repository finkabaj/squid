package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/finkabaj/squid/back/internal/config"
	"github.com/finkabaj/squid/back/internal/repository"
	"github.com/finkabaj/squid/back/internal/types"
	"github.com/finkabaj/squid/back/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

type ValidateJWTCtxKey struct{}

func ValidateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.HandleError(w, utils.NewUnauthorizedError(errors.New("invalid authorization header")))
			return
		}

		tokenString := authHeader[7:]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return config.Data.JWTSecret, nil
		})

		if err != nil {
			utils.HandleError(w, utils.NewUnauthorizedError(err))
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok || !token.Valid {
			utils.HandleError(w, utils.NewUnauthorizedError(err))
			return
		}

		exp, ok := claims["expires_at"].(float64)

		if !ok || time.Now().Unix() > int64(exp) {
			utils.HandleError(w, utils.NewUnauthorizedError(errors.New("token expired")))
			return
		}

		userID, ok := claims["user_id"].(string)

		if !ok {
			utils.HandleError(w, utils.NewUnauthorizedError(errors.New("no id found in token")))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		user, err := repository.GetUser(ctx, &userID, nil)

		if err != nil {
			utils.HandleError(w, utils.NewUnauthorizedError(err))
			return
		}

		newCtx := context.WithValue(r.Context(), ValidateJWTCtxKey{}, user)
		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func UserFromContext(ctx context.Context) types.User {
	return ctx.Value(ValidateJWTCtxKey{}).(types.User)
}
