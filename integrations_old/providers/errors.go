package providers

import "errors"

var (
	// ErrProviderNil indicates a builder returned a nil provider without an error
	ErrProviderNil = errors.New("integrationsv2/providers: provider is nil")
	// ErrBuilderNil indicates a BuilderFunc has a nil BuildFunc field
	ErrBuilderNil = errors.New("integrationsv2/providers: builder func is nil")
	// ErrAuthTypeMismatch indicates the provider spec auth type does not match the expected auth kind
	ErrAuthTypeMismatch = errors.New("integrationsv2/providers: auth type mismatch")
	// ErrSpecOAuthRequired indicates the provider spec is missing OAuth configuration
	ErrSpecOAuthRequired = errors.New("integrationsv2/providers: oauth spec required")
	// ErrSpecWorkloadIdentityRequired indicates the provider spec is missing workload identity configuration
	ErrSpecWorkloadIdentityRequired = errors.New("integrationsv2/providers: workload identity spec required")
	// ErrTokenUnavailable indicates the stored credential did not include an oauth2 token
	ErrTokenUnavailable = errors.New("integrationsv2/providers: oauth token unavailable")
	// ErrStateGeneration indicates random state generation failed
	ErrStateGeneration = errors.New("integrationsv2/providers: state generation failed")
	// ErrCodeExchange indicates the authorization code exchange failed
	ErrCodeExchange = errors.New("integrationsv2/providers: code exchange failed")
	// ErrTokenRefresh indicates token refresh failed
	ErrTokenRefresh = errors.New("integrationsv2/providers: token refresh failed")
)
