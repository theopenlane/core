package echocontext

import (
	"context"

	echo "github.com/theopenlane/echox"
)

// EchoContextKey is the context key for the echo.Context
var EchoContextKey = &ContextKey{"EchoContextKey"}

// ContextKey is the key name for the additional context
type ContextKey struct {
	name string
}

// CustomContext contains the echo.Context and request context.Context
type CustomContext struct {
	echo.Context
	ctx context.Context
}

// EchoContextToContextMiddleware is the middleware that adds the echo.Context to the parent context
func EchoContextToContextMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := context.WithValue(c.Request().Context(), EchoContextKey, c)

			c.SetRequest(c.Request().WithContext(ctx))

			cc := &CustomContext{c, ctx}

			return next(cc)
		}
	}
}

// EchoContextFromContext gets the echo.Context from the parent context
func EchoContextFromContext(ctx context.Context) (echo.Context, error) {
	// retrieve echo.Context from provided context
	echoContext := ctx.Value(EchoContextKey)
	if echoContext == nil {
		return nil, ErrUnableToRetrieveEchoContext
	}

	// type cast the context to ensure echo.Context
	ec, ok := echoContext.(echo.Context)
	if !ok {
		return ec, ErrUnableToRetrieveEchoContext
	}

	return ec, nil
}
