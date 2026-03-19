package types

import "errors"

var (
	// ErrClientCastFailed indicates a registered client instance could not be cast to the expected type
	ErrClientCastFailed = errors.New("integrations/types: client cast failed")
)
