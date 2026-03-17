package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/authentication"
	"github.com/the127/hivetrack/internal/config"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
	"go.uber.org/zap"
)

// AuthMiddleware extracts the Bearer token, verifies it, upserts the user, and
// injects CurrentUser into the context.
func AuthMiddleware(verifier *authentication.OIDCVerifier, logger *zap.Logger, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"errors":[{"code":"unauthorized","message":"missing bearer token"}]}`, http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := verifier.VerifyToken(r.Context(), token)
			if err != nil {
				logger.Warn("token verification failed", zap.Error(err))
				http.Error(w, `{"errors":[{"code":"unauthorized","message":"invalid token"}]}`, http.StatusUnauthorized)
				return
			}

			// Get DbContext from the request context
			db := repositories.GetDbContext(r.Context())

			user, err := db.Users().GetBySub(r.Context(), claims.Sub)
			if err != nil {
				logger.Error("failed to get user by sub", zap.Error(err))
				http.Error(w, `{"errors":[{"code":"internal","message":"internal server error"}]}`, http.StatusInternalServerError)
				return
			}

			now := time.Now()
			if user == nil {
				isAdmin := cfg.InitialAdmin.Email != "" && claims.Email == cfg.InitialAdmin.Email
				user = &models.User{
					ID:          uuid.New(),
					Sub:         claims.Sub,
					Email:       claims.Email,
					DisplayName: claims.Name,
					IsAdmin:     isAdmin,
					CreatedAt:   now,
					LastLoginAt: now,
				}
			} else {
				user.Email = claims.Email
				user.DisplayName = claims.Name
				user.LastLoginAt = now
			}

			if err := db.Users().Upsert(r.Context(), user); err != nil {
				logger.Error("failed to upsert user", zap.Error(err))
				http.Error(w, `{"errors":[{"code":"internal","message":"internal server error"}]}`, http.StatusInternalServerError)
				return
			}
			// Commit immediately — user upsert is independent of the request operation.
			if err := db.Commit(r.Context()); err != nil {
				logger.Error("failed to commit user upsert", zap.Error(err))
				http.Error(w, `{"errors":[{"code":"internal","message":"internal server error"}]}`, http.StatusInternalServerError)
				return
			}

			ctx := authentication.ContextWithCurrentUser(r.Context(), authentication.CurrentUser{
				ID:      user.ID,
				Sub:     user.Sub,
				Email:   user.Email,
				IsAdmin: user.IsAdmin,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
