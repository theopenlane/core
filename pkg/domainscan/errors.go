package domainscan

import "errors"

var (
	// ErrInvalidDomain is returned when its not able to determine the apex domain
	ErrInvalidDomain = errors.New("could not determine domain")
)
