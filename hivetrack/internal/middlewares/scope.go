package middlewares

import (
	"context"
	"net/http"

	"github.com/The127/ioc"
	"github.com/gorilla/mux"
)

type scopeKeyType struct{}

// ScopeMiddleware creates a new IoC scope per request and stores it in context.
func ScopeMiddleware(root *ioc.DependencyProvider) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scope := root.NewScope()
			defer func() {
				if err := scope.Close(); err != nil {
					// best effort
					_ = err
				}
			}()

			r = r.WithContext(ContextWithScope(r.Context(), scope))
			next.ServeHTTP(w, r)
		})
	}
}

// ContextWithScope stores the scope in the context.
func ContextWithScope(ctx context.Context, scope *ioc.DependencyProvider) context.Context {
	return context.WithValue(ctx, scopeKeyType{}, scope)
}

// GetScope retrieves the scope from the context. Panics if not set.
func GetScope(ctx context.Context) *ioc.DependencyProvider {
	scope, ok := ctx.Value(scopeKeyType{}).(*ioc.DependencyProvider)
	if !ok || scope == nil {
		panic("IoC scope not found in context")
	}
	return scope
}
