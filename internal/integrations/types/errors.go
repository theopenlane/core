package types //nolint:revive

import "errors"

var (
	// ErrClientCastFailed indicates a registered client instance could not be cast to the expected type
	ErrClientCastFailed = errors.New("integrations: client cast failed")
	// ErrCredentialRefNotFound indicates the requested credential ref was not found in the definition
	ErrCredentialRefNotFound = errors.New("integrations: credential ref not found")
	// ErrConnectionRefNotFound indicates the requested connection credential ref was not found in the definition
	ErrConnectionRefNotFound = errors.New("integrations: connection credential ref not found")
)
