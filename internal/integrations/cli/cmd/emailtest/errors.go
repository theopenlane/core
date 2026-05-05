//go:build examples

package emailtest

import "errors"

var (
	// ErrRecipientRequired is returned when --to is not provided
	ErrRecipientRequired = errors.New("--to is required")
	// ErrNameRequired is returned when --name is not provided for the send command
	ErrNameRequired = errors.New("--name is required")
	// ErrHostRequired is returned when no server host is configured
	ErrHostRequired = errors.New("no server host configured; set --host or OPENLANE_HOST")
	// ErrServerError is returned when the server responds with a non-200 status
	ErrServerError = errors.New("server returned error")
)
