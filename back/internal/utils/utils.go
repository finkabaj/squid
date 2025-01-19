package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"time"

	"github.com/finkabaj/squid/back/internal/config"

	"golang.org/x/crypto/bcrypt"

	"github.com/finkabaj/squid/back/internal/types"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

type OkResponse struct {
	Message string `json:"message"`
}

// MarshalBody Writes json body to w, sends status code
func MarshalBody(w http.ResponseWriter, status int, v any) (err error) {
	(w).Header().Set("Content-Type", "application/json; charset=utf8")
	(w).WriteHeader(status)
	err = json.NewEncoder(w).Encode(v)

	return
}

// UnmarshalBody Reads json body to v. Body is ReadCloser
func UnmarshalBody(body io.ReadCloser, v any) (err error) {
	err = json.NewDecoder(body).Decode(v)

	return
}

func ValidateSliceOrStruct(w http.ResponseWriter, validate *validator.Validate, v any) (haveError bool) {
	isSlice := reflect.TypeOf(v).Kind() == reflect.Slice

	var err error

	if isSlice {
		err = validate.Var(v, "required,dive")
	} else {
		err = validate.Struct(v)
	}

	if err != nil {
		var invalidValidationError *validator.InvalidValidationError
		if errors.As(err, &invalidValidationError) {
			// if you see this error that means that it's time to correct validate_json implementation (or you fucked up json)
			HandleError(w, err)
			return true
		}
		validationErrors := make(map[string]string)
		for _, e := range err.(validator.ValidationErrors) {
			validationErrors[e.Field()] = e.Tag()
		}

		HandleError(w, NewValidationError(validationErrors))

		return true
	}

	return
}

func HashPassword(password *string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(*password), config.Data.SaltRounds)
	return string(bytes), err
}

func CheckPasswordHash(password, hash *string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(*hash), []byte(*password))
	return err == nil
}

func CreateJWTRefresh(refreshToken *types.RefreshToken) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         refreshToken.ID,
		"user_id":    refreshToken.UserID,
		"created_at": refreshToken.CreatedAt.Unix(),
		"expires_at": refreshToken.ExpiresAt.Unix(),
	})

	tokenStr, err := token.SignedString(config.Data.JWTSecret)

	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func CreateJWT(user *types.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user.ID,
		"email":      user.Email,
		"created_at": time.Now().Unix(),
		"expires_at": time.Now().Add(time.Minute * time.Duration(config.Data.AccessTokenExpMinutes)).Unix(),
	})

	tokenStr, err := token.SignedString(config.Data.JWTSecret)

	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func CreateJWTPair(user *types.User, refreshToken *types.RefreshToken) (map[string]string, error) {
	if refreshToken == nil || user == nil {
		return nil, errors.New("refresh token or user is nil")
	}

	refreshTokenStr, err := CreateJWTRefresh(refreshToken)

	if err != nil {
		return nil, err
	}

	accessTokenStr, err := CreateJWT(user)

	if err != nil {
		return nil, err
	}

	return map[string]string{
		"refreshToken": refreshTokenStr,
		"accessToken":  accessTokenStr,
	}, nil
}

// DO NOT USE if don't know what it for!!!
func UpdateSelector[T any](update *T, current *T) *T {
	if update != nil {
		return update
	}

	return current
}
