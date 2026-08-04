package system

import "errors"

var (
	// ErrResultEncode indicates an operation result payload could not be encoded
	ErrResultEncode = errors.New("system: failed to encode operation result")
	// ErrOperationConfigInvalid indicates an operation config payload could not be decoded
	ErrOperationConfigInvalid = errors.New("system: operation config invalid")
)
