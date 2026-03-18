package activation

import "errors"

var (
	// ErrHealthCheckFailed indicates the provider health check failed.
	ErrHealthCheckFailed = errors.New("activation: health check failed")
	// ErrStoreRequired indicates the credential store is required.
	ErrStoreRequired = errors.New("activation: credential store required")
	// ErrKeymakerRequired indicates the keymaker dependency is required.
	ErrKeymakerRequired = errors.New("activation: keymaker required")
	// ErrOperationsRequired indicates the operations manager is required for validation.
	ErrOperationsRequired = errors.New("activation: operations manager required")
	// ErrMinterRequired indicates the payload minter is required.
	ErrMinterRequired = errors.New("activation: payload minter required")
)
