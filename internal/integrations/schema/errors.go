package schema

import "errors"

var (
	// ErrProviderStateDecode is returned when provider state cannot be decoded
	ErrProviderStateDecode = errors.New("failed to decode provider state")
)
