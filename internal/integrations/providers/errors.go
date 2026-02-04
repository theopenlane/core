package providers

import "errors"

var (
	// ErrSpecOAuthRequired indicates the provider spec is missing OAuth configuration
	ErrSpecOAuthRequired = errors.New("integrations/providers: oauth spec required")
	// ErrSpecWorkloadIdentityRequired indicates the provider spec is missing workload identity configuration
	ErrSpecWorkloadIdentityRequired = errors.New("integrations/providers: workload identity spec required")
	// ErrTokenUnavailable indicates the stored credential did not include an oauth2 token
	ErrTokenUnavailable = errors.New("integrations/providers: oauth token unavailable")
	// ErrRelyingPartyInit indicates Zitadel RP construction failed
	ErrRelyingPartyInit = errors.New("integrations/providers: relying party initialization failed")
	// ErrStateGeneration indicates random state generation failed
	ErrStateGeneration = errors.New("integrations/providers: state generation failed")
	// ErrCodeExchange indicates the authorization code exchange failed
	ErrCodeExchange = errors.New("integrations/providers: code exchange failed")
	// ErrTokenRefresh indicates token refresh failed
	ErrTokenRefresh = errors.New("integrations/providers: token refresh failed")
	// ErrProviderNil indicates a builder returned a nil provider without an error
	ErrProviderNil = errors.New("integrations/providers: provider is nil")
)
