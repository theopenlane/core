package oauth

import "errors"

var (
	// ErrSpecOAuthRequired indicates the provider spec is missing OAuth configuration
	ErrSpecOAuthRequired = errors.New("providers/oauth: oauth spec required")
	// ErrAuthTypeMismatch indicates the spec auth type does not support interactive OAuth flows
	ErrAuthTypeMismatch = errors.New("providers/oauth: auth type does not support interactive flow")
	// ErrRelyingPartyInit indicates Zitadel RP construction failed
	ErrRelyingPartyInit = errors.New("providers/oauth: relying party initialization failed")
	// ErrStateGeneration indicates random state generation failed
	ErrStateGeneration = errors.New("providers/oauth: state generation failed")
	// ErrCodeExchange indicates the authorization code exchange failed
	ErrCodeExchange = errors.New("providers/oauth: code exchange failed")
	// ErrTokenUnavailable indicates the stored credential did not include an oauth2 token
	ErrTokenUnavailable = errors.New("providers/oauth: oauth token unavailable")
	// ErrTokenRefresh indicates token refresh failed
	ErrTokenRefresh = errors.New("providers/oauth: token refresh failed")
	// ErrClaimsEncode indicates claims could not be serialized to a map
	ErrClaimsEncode = errors.New("providers/oauth: claims encoding failed")
)
