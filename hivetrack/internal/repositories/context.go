package repositories

import (
	"context"
	"fmt"
)

type dbContextKeyType struct{}

// ContextWithDbContext stores the DbContext in the context.
func ContextWithDbContext(ctx context.Context, db DbContext) context.Context {
	return context.WithValue(ctx, dbContextKeyType{}, db)
}

// GetDbContext retrieves the DbContext from the context.
// Panics if the DbContext is not set.
func GetDbContext(ctx context.Context) DbContext {
	db, ok := ctx.Value(dbContextKeyType{}).(DbContext)
	if !ok || db == nil {
		panic(fmt.Errorf("DbContext not found in context — did you forget to register it in the IoC scope?"))
	}
	return db
}
