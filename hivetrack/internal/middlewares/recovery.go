package middlewares

import (
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
)

// RecoveryMiddleware catches panics and returns 500.
func RecoveryMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered",
						zap.Any("panic", rec),
						zap.String("stack", string(debug.Stack())),
					)
					http.Error(w, `{"errors":[{"code":"internal","message":"internal server error"}]}`, http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
