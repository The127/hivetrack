package middlewares

import (
	"net/http"
	"strings"
)

// CORSMiddleware adds CORS headers for the given allowed origins.
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", strings.Join([]string{
				"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
			}, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join([]string{
				"Content-Type", "Authorization",
			}, ", "))
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
