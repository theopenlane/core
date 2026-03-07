package activation

import "errors"

var (
	// ErrHealthCheckFailed indicates the provider health check failed.
	ErrHealthCheckFailed = errors.New("activation: health check failed")
	// ErrStoreRequired indicates the credential store is required.
	ErrStoreRequired = errors.New("activation: credential store required")
	// ErrHealthValidatorRequired indicates a health validator dependency is required for validation.
	ErrHealthValidatorRequired = errors.New("activation: health validator required")
	// ErrMinterRequired indicates the payload minter is required.
	ErrMinterRequired = errors.New("activation: payload minter required")
)
