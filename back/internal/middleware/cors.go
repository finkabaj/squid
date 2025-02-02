package middleware

import (
	"net/http"
)

func Cors(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			for _, allowed := range allowedOrigins {
				if allowed == origin || allowed == "*" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
