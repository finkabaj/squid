package middleware

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/finkabaj/squid/back/internal/utils"
	"github.com/go-playground/validator/v10"
)

type validateJsonCtxKey struct{}

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})
}

// ValidateJson is a middleware that validates the json body of a request and respond
// with an error if json body is empty or if the json body is invalid.
// To use it, you must pass a struct to generic type T that has json and validate tags and call it as a function.
func ValidateJson[T any]() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body := r.Body
			defer body.Close()

			var v T

			if err := utils.UnmarshalBody(body, &v); err == io.EOF {
				utils.NewErrorResponseBuilder(utils.ErrEmptyBody).
					SetStatus(http.StatusBadRequest).
					SetMessage("empty request body").
					Send(w)
				return
			}

			if haveError := utils.ValidateSliceOrStruct(w, validate, v); haveError {
				return
			}

			ctx := context.WithValue(r.Context(), validateJsonCtxKey{}, v)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func JsonFromContext(ctx context.Context) any {
	return ctx.Value(validateJsonCtxKey{})
}
