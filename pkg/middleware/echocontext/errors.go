package echocontext

import (
	"errors"
)

var (
	// ErrUnableToRetrieveEchoContext is returned when the echo context is unable to be parsed from parent context
	ErrUnableToRetrieveEchoContext = errors.New("unable to retrieve echo context")
)
