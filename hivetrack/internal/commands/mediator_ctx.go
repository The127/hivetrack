package commands

import (
	"context"

	"github.com/The127/mediatr"
)

type mediatorKey struct{}

// ContextWithMediator stores a mediator in the context for use by command handlers.
func ContextWithMediator(ctx context.Context, m mediatr.Mediator) context.Context {
	return context.WithValue(ctx, mediatorKey{}, m)
}

func getMediatorFromContext(ctx context.Context) (mediatr.Mediator, bool) {
	m, ok := ctx.Value(mediatorKey{}).(mediatr.Mediator)
	return m, ok
}
