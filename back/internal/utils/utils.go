package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"reflect"

	"github.com/finkabaj/squid/back/internal/types"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrEmptyBody  = errors.New("empty request body")
	ErrValidation = errors.New("validation error")
)

type OkResponse struct {
	Message string `json:"message"`
}

// Reads json body to v. Body is ReadCloser
func UnmarshalBody(body io.ReadCloser, v any) (err error) {
	err = json.NewDecoder(body).Decode(v)

	return
}

// Writes json body to w, sends status code
func MarshalBody(w http.ResponseWriter, status int, v any) (err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf8")
	w.WriteHeader(status)
	err = json.NewEncoder(w).Encode(v)

	return
}

// Use this function if you have UnmarshalJSON method in your struct
func UnmarshalBodyBytes(body []byte, v any) (err error) {
	if string(body) == "[]" {
		// If the JSON string is an empty array, set the target to an empty slice
		reflect.ValueOf(v).Elem().Set(reflect.MakeSlice(reflect.TypeOf(v).Elem(), 0, 0))
		return nil
	}

	err = json.Unmarshal(body, v)

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
		if _, ok := err.(*validator.InvalidValidationError); ok {
			// if you see this error that means that it's time to correct validate_json implementation (or you fucked up json)
			SendBadRequestError(w, "Invalid json while validation body")
			return true
		}
		validationErrors := make(map[string]string)
		for _, e := range err.(validator.ValidationErrors) {
			validationErrors[e.Field()] = e.Tag()
		}

		SendValidationError(w, validationErrors)

		return true
	}

	return
}

var secret = os.Getenv("JWT_SECRET")

func CreateJWTAuth(username *string, email *string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": *username,
		"email":    *email,
	})

	tokenStr, err := token.SignedString(secret)

	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func CreateJWTRefresh(refresh types.JWTRefreshToken) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{})

	tokenStr, err := token.SignedString(secret)

	if err != nil {
		return "", err
	}

	return tokenStr, nil
}
