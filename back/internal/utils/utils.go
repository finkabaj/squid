package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"

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

var secret = os.Getenv("JWT_SECRET")

func CreateJWTAuth(username, email *string) (string, error) {
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

func HashPassword(password *string) (string, error) {
	saltRounds, err := strconv.Atoi(os.Getenv("SALT_ROUNDS"))
	if err != nil {
		return "", errors.New("SALT_ROUNDS env var not set")
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(*password), saltRounds)
	return string(bytes), err
}

func CheckPasswordHash(password, hash *string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(*hash), []byte(*password))
	return err == nil
}

func CreateJWTRefresh(refresh types.RefreshToken) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{})

	tokenStr, err := token.SignedString(secret)

	if err != nil {
		return "", err
	}

	return tokenStr, nil
}
