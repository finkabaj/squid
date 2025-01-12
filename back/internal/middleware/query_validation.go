package middleware

import (
	"context"
	"net/http"

	"github.com/finkabaj/squid/back/internal/utils"
)

type validateQueryCtxKey struct{}

// ValidateQuery Validation middleware for query params. Decoder function should decode query based on RawQuery.
func ValidateQuery[T any](decoder func(q string) T) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.RawQuery
			dQuery := decoder(query)

			if haveError := utils.ValidateSliceOrStruct(w, validate, dQuery); haveError {
				return
			}

			ctx := context.WithValue(r.Context(), validateQueryCtxKey{}, dQuery)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func QueryFromContext(ctx context.Context) any {
	return ctx.Value(validateQueryCtxKey{})
}
