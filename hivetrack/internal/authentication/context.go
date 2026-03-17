package authentication

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type currentUserKeyType struct{}

// CurrentUser holds the authenticated user extracted from the JWT.
type CurrentUser struct {
	ID      uuid.UUID
	Sub     string
	Email   string
	IsAdmin bool
}

// ContextWithCurrentUser stores the current user in the context.
func ContextWithCurrentUser(ctx context.Context, user CurrentUser) context.Context {
	return context.WithValue(ctx, currentUserKeyType{}, user)
}

// GetCurrentUser retrieves the current user from context. Returns zero value if not set.
func GetCurrentUser(ctx context.Context) (CurrentUser, bool) {
	u, ok := ctx.Value(currentUserKeyType{}).(CurrentUser)
	return u, ok
}

// MustGetCurrentUser retrieves the current user or panics if not set.
func MustGetCurrentUser(ctx context.Context) CurrentUser {
	u, ok := GetCurrentUser(ctx)
	if !ok {
		panic(fmt.Errorf("current user not found in context"))
	}
	return u
}
