package domainscan

import "errors"

var (
	// ErrUnexpectedErrorCode is returned when a non-200 error code is returned from an upstream API
	ErrUnexpectedErrorCode = errors.New("unexpected error code")
	// ErrInvalidDomain is returned when its not able to determine the apex domain
	ErrInvalidDomain = errors.New("could not determine domain")
)
