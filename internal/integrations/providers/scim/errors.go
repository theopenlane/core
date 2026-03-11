package scim

import "errors"

var (
	// ErrBeginAuthNotSupported is returned when BeginAuth is called on the SCIM provider
	ErrBeginAuthNotSupported = errors.New("integrationsv2/providers/scim: BeginAuth is not supported for push-based SCIM provider")
	// ErrMintNotSupported is returned when Mint is called on the SCIM provider
	ErrMintNotSupported = errors.New("integrationsv2/providers/scim: Mint is not supported; SCIM credentials are org API tokens")
)
