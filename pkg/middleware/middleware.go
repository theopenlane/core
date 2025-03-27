// Package middleware provides middleware for http Handlers.
package middleware

import "net/http"

// A Chain is a middleware chain use for http request processing.
type Chain struct {
	mw []Func
}

// Func is a function that acts as middleware for http Handlers.
type Func func(next http.Handler) http.Handler

// NewChain creates a new Middleware chain
func NewChain(middlewares ...Func) Chain {
	return Chain{
		mw: append([]Func{}, middlewares...),
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
func Conditional(middleware Func, condition func(r *http.Request) bool) Func {
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
