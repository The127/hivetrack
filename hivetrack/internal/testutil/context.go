package testutil

import (
	"context"

	"github.com/the127/hivetrack/internal/authentication"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

// ContextWithDb stores the DbContext directly in the context for use by commands/queries.
func ContextWithDb(db repositories.DbContext) context.Context {
	return repositories.ContextWithDbContext(context.Background(), db)
}

// ContextWithUser adds the given user as the current authenticated user.
func ContextWithUser(ctx context.Context, user models.User) context.Context {
	return authentication.ContextWithCurrentUser(ctx, authentication.CurrentUser{
		ID:      user.ID,
		Sub:     user.Sub,
		Email:   user.Email,
		IsAdmin: user.IsAdmin,
	})
}
