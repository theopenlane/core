package webauthn

import (
	"errors"
)

var (
	// ErrUserNotFound is returned when the user couldn't be found
	ErrUserNotFound = errors.New("user not found")

	// ErrSessionNotFound is returned when the session couldn't be found
	ErrSessionNotFound = errors.New("session not found")
)
