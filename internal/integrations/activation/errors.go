package activation

import "errors"

var (
	// ErrOrgIDRequired indicates the organization identifier was omitted.
	ErrOrgIDRequired = errors.New("integrations: org id required")
	// ErrProviderRequired indicates the provider type was not specified.
	ErrProviderRequired = errors.New("activation: provider required")
	// ErrHealthCheckFailed indicates the provider health check failed.
	ErrHealthCheckFailed = errors.New("activation: health check failed")
	// ErrStoreRequired indicates the credential store dependency is missing.
	ErrStoreRequired = errors.New("activation: credential store required")
	// ErrHealthValidatorRequired indicates a health validator dependency is required for validation.
	ErrHealthValidatorRequired = errors.New("activation: health validator required")
	// ErrMinterRequired indicates the credential minter dependency is missing.
	ErrMinterRequired = errors.New("activation: credential minter required")
)
