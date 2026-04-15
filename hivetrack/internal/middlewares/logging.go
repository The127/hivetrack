package middlewares

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Ensure responseWriter implements http.Flusher so SSE handlers can flush.
var _ http.Flusher = (*responseWriter)(nil)

// LoggingMiddleware logs each HTTP request with method, path, status and duration.
// 5xx responses are logged at Error level, 4xx at Warn, everything else at Info.
func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", rw.statusCode),
				zap.Duration("duration", time.Since(start)),
			}
			if q := r.URL.RawQuery; q != "" {
				fields = append(fields, zap.String("query", q))
			}

			switch {
			case rw.statusCode >= 500:
				logger.Error("request", fields...)
			case rw.statusCode >= 400:
				logger.Warn("request", fields...)
			default:
				logger.Info("request", fields...)
			}
		})
	}
}
