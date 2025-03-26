// Package middleware provides middleware for http Handlers.
package middleware

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// A Chain is a middleware chain use for http request processing.
type Chain struct {
	mw []MiddlewareFunc
}

// MiddlewareFunc is a function that acts as middleware for http Handlers.
type MiddlewareFunc func(next http.Handler) http.Handler

// NewChain creates a new Middleware chain
func NewChain(middlewares ...MiddlewareFunc) Chain {
	return Chain{
		mw: append([]MiddlewareFunc{}, middlewares...),
	}
}

// Chain returns a http.Handler that chains the middleware onion-style around the handler.
func (c Chain) Chain(handler http.Handler) http.Handler {
	for i := len(c.mw) - 1; i >= 0; i-- {
		handler = c.mw[i](handler)
	}

	return handler
}

// Conditional is a middleware that only executes middleware if the condition
// returns true for the request. If the condition returns false, the middleware
// is skipped, and request handling moves on to the next handler in the chain.
func Conditional(middleware MiddlewareFunc, condition func(r *http.Request) bool) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler := next
			if condition(r) {
				handler = middleware(next)
			}

			handler.ServeHTTP(w, r)
		})
	}
}

// Conditional applies the middleware if the condition is true
func EchoConditional(condition bool, middleware echo.MiddlewareFunc) echo.MiddlewareFunc {
	if condition {
		return middleware
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}

// Chain chains multiple middleware functions together
func EchoChain(middlewares ...echo.MiddlewareFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
