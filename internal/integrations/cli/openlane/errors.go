//go:build examples

package openlane

import "errors"

var (
	// ErrLogin is returned when user login fails
	ErrLogin = errors.New("login failed")
	// ErrHostRequired is returned when host is empty
	ErrHostRequired = errors.New("openlane host is required")
	// ErrTokenRequired is returned when auth-mode=token but no token is provided
	ErrTokenRequired = errors.New("token is required for auth-mode=token")
	// ErrUnsupportedAuthMode is returned when an unknown auth-mode is specified
	ErrUnsupportedAuthMode = errors.New("unsupported auth-mode (expected auto, token, or credentials)")
	// ErrCredentialsRequired is returned when email or password is missing in credentials mode
	ErrCredentialsRequired = errors.New("email and password are required for auth-mode=credentials")
)
