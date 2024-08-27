package oauth2

import (
	"context"

	"golang.org/x/oauth2"
)

// unexported key type prevents collisions
type key int

const (
	tokenKey    key = iota
	stateKey    key = iota
	errorKey    key = iota
	redirectKey key = iota
)

// WithState returns a copy of ctx that stores the state value
func WithState(ctx context.Context, state string) context.Context {
	return context.WithValue(ctx, stateKey, state)
}

// StateFromContext returns the state value from the ctx
func StateFromContext(ctx context.Context) (string, error) {
	state, ok := ctx.Value(stateKey).(string)

	if !ok {
		return "", ErrContextMissingStateValue
	}

	return state, nil
}

// WithToken returns a copy of ctx that stores the Token
func WithToken(ctx context.Context, token *oauth2.Token) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

// TokenFromContext returns the Token from the ctx
func TokenFromContext(ctx context.Context) (*oauth2.Token, error) {
	token, ok := ctx.Value(tokenKey).(*oauth2.Token)

	if !ok {
		return nil, ErrContextMissingToken
	}

	return token, nil
}

// WithError returns a copy of context that stores the given error value
func WithError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, errorKey, err)
}

// ErrorFromContext returns the error value from the ctx or an error that the
// context was missing an error value
func ErrorFromContext(ctx context.Context) error {
	err, ok := ctx.Value(errorKey).(error)
	if !ok {
		return ErrContextMissingErrorValue
	}

	return err
}

// WithRedirectURL returns a copy of ctx that stores the redirect value
func WithRedirectURL(ctx context.Context, redirect string) context.Context {
	return context.WithValue(ctx, redirectKey, redirect)
}

// RedirectFromContext returns the redirect value from the ctx
func RedirectFromContext(ctx context.Context) string {
	redirect, ok := ctx.Value(redirectKey).(string)

	if !ok {
		return ""
	}

	return redirect
}
