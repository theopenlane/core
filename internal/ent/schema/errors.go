package schema

import (
	"errors"
)

var (
	// ErrInvalidTokenSize is returned when session token size is invalid
	ErrInvalidTokenSize = errors.New("invalid token size")

	// ErrContainsSpaces is returned when field contains spaces
	ErrContainsSpaces = errors.New("field should not contain spaces")

	// ErrUnexpectedMutationType is returned when an unexpected mutation type is encountered
	ErrUnexpectedMutationType = errors.New("unexpected mutation type")

	// ErrInternalServerError is returned when an error occurs that should not be exposed to the user and is not the user's fault, such as an error writing to the database or authz system
	ErrInternalServerError = errors.New("internal server error")
)
