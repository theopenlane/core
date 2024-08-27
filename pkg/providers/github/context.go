package github

import (
	"context"

	"github.com/google/go-github/v63/github"
)

// unexported key type prevents collisions
type key int

const (
	userKey  key = iota
	errorKey key = iota
)

// WithUser returns a copy of context that stores the GitHub User
func WithUser(ctx context.Context, user *github.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext returns the GitHub User from the context
func UserFromContext(ctx context.Context) (*github.User, error) {
	user, ok := ctx.Value(userKey).(*github.User)
	if !ok {
		return nil, ErrContextMissingGithubUser
	}

	return user, nil
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
