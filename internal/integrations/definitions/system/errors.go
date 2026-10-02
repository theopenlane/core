package system

import "errors"

var (
	// ErrResultEncode indicates an operation result payload could not be encoded
	ErrResultEncode = errors.New("system: failed to encode operation result")
)
