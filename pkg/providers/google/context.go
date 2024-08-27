package google

import (
	"context"

	google "google.golang.org/api/oauth2/v2"
)

// unexported key type prevents collisions
type key int

const (
	userKey  key = iota
	errorKey key = iota
)

// WithUser returns a copy of ctx that stores the Google Userinfo
func WithUser(ctx context.Context, user *google.Userinfo) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext returns the Google Userinfo from the ctx
func UserFromContext(ctx context.Context) (*google.Userinfo, error) {
	user, ok := ctx.Value(userKey).(*google.Userinfo)
	if !ok {
		return nil, ErrContextMissingGoogleUser
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
