package auth

import "errors"

var (
	// ErrOAuthRelyingPartyInit indicates Zitadel relying party construction failed
	ErrOAuthRelyingPartyInit = errors.New("auth: oauth relying party initialization failed")
	// ErrOAuthStateGeneration indicates random CSRF state generation failed
	ErrOAuthStateGeneration = errors.New("auth: oauth state generation failed")
	// ErrOAuthStateInvalid indicates the stored oauth start state could not be decoded
	ErrOAuthStateInvalid = errors.New("auth: oauth state invalid")
	// ErrOAuthStateMissing indicates the callback omitted the required CSRF state
	ErrOAuthStateMissing = errors.New("auth: oauth callback state missing")
	// ErrOAuthStateMismatch indicates the callback state does not match the stored CSRF state
	ErrOAuthStateMismatch = errors.New("auth: oauth state mismatch")
	// ErrOAuthCodeMissing indicates the authorization code is absent from the callback input
	ErrOAuthCodeMissing = errors.New("auth: oauth callback code missing")
	// ErrOAuthCodeExchange indicates the authorization code exchange failed
	ErrOAuthCodeExchange = errors.New("auth: oauth code exchange failed")
	// ErrOAuthClaimsEncode indicates OIDC claims could not be serialized to a map
	ErrOAuthClaimsEncode = errors.New("auth: oauth claims encoding failed")
)
