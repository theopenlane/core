package shortlinks

import "errors"

var (
	// ErrMissingAuthenticationParams indicates a missing client ID or secret
	ErrMissingAuthenticationParams = errors.New("shortlinks: missing client ID")
	// ErrMissingURL indicates a missing URL
	ErrMissingURL = errors.New("shortlinks: missing URL")
)
