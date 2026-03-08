package oauth

import "errors"

var (
	// ErrAuthTypeMismatch indicates the provider spec auth type is not oauth2/oidc for interactive OAuth flows.
	ErrAuthTypeMismatch = errors.New("oauth: auth type mismatch")
)
