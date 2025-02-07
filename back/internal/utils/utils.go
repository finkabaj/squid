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
func UnmarshalBody(body io.ReadCloser, v any) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
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

func CreateJWTRefresh(refreshToken *types.RefreshToken) (string, time.Time, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         refreshToken.ID,
		"user_id":    refreshToken.UserID,
		"created_at": refreshToken.CreatedAt.Unix(),
		"expires_at": refreshToken.ExpiresAt.Unix(),
	})

	tokenStr, err := token.SignedString(config.Data.JWTSecret)

	if err != nil {
		return "", time.Time{}, err
	}

	return tokenStr, refreshToken.ExpiresAt, nil
}

func CreateJWT(user *types.User) (string, time.Time, error) {
	expAt := time.Now().Add(time.Minute * time.Duration(config.Data.AccessTokenExpMinutes))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user.ID,
		"email":      user.Email,
		"created_at": time.Now().Unix(),
		"expires_at": expAt.Unix(),
	})

	tokenStr, err := token.SignedString(config.Data.JWTSecret)

	if err != nil {
		return "", time.Time{}, err
	}

	return tokenStr, expAt, nil
}

func CreateJWTPair(user *types.User, refreshToken *types.RefreshToken) (map[string]string, map[string]time.Time, error) {
	if refreshToken == nil || user == nil {
		return nil, nil, errors.New("refresh token or user is nil")
	}

	refreshTokenStr, refreshTokenExpAt, err := CreateJWTRefresh(refreshToken)

	if err != nil {
		return nil, nil, err
	}

	accessTokenStr, accessTokenExpAt, err := CreateJWT(user)

	if err != nil {
		return nil, nil, err
	}

	return map[string]string{
			"refreshToken": refreshTokenStr,
			"accessToken":  accessTokenStr,
		}, map[string]time.Time{
			"refreshToken": refreshTokenExpAt,
			"accessToken":  accessTokenExpAt,
		}, nil
}

func Map[T any, F any](mapper func(int, T) F, values []T) []F {
	result := make([]F, len(values))

	for i, v := range values {
		result[i] = mapper(i, v)
	}

	return result
}

func Have[T any](haveF func(int, T) bool, data []T) bool {
	for i, v := range data {
		if haveF(i, v) {
			return true
		}
	}

	return false
}

func SetTokenCookie(w http.ResponseWriter, name, token string, expiry time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    token,
		Expires:  expiry,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
}
