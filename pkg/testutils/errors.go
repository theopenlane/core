package testutils

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidDBURI is returned when an invalid DB URI is used
	ErrInvalidDBURI = errors.New("invalid DB URI")
)

func newURIError(uri string) error {
	return fmt.Errorf("%w: %s", ErrInvalidDBURI, uri)
}
