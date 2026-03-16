package scim

import "errors"

var (
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("scim: result encode failed")
)
