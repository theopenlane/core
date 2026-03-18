package apikey

import "errors"

var (
	// ErrProviderMetadataRequired indicates provider metadata is required but not supplied
	ErrProviderMetadataRequired = errors.New("apikey: provider metadata required")
	// ErrTokenFieldRequired indicates the configured token field is missing from metadata
	ErrTokenFieldRequired = errors.New("apikey: token field required")
	// ErrAuthTypeMismatch indicates the provider spec specifies an incompatible auth type
	ErrAuthTypeMismatch = errors.New("apikey: auth type mismatch")
	// ErrBeginAuthNotSupported indicates BeginAuth is not supported for API key providers
	ErrBeginAuthNotSupported = errors.New("apikey: BeginAuth is not supported; configure credentials via metadata")
)
