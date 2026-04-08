package oidclocal

import "errors"

var (
	// ErrOAuthTokenMissing indicates the OAuth access token is missing from the auth-managed credential
	ErrOAuthTokenMissing = errors.New("oidclocal: oauth token missing")
	// ErrSubjectMissing indicates the OIDC subject claim is missing from the auth-managed credential
	ErrSubjectMissing = errors.New("oidclocal: subject claim missing")
	// ErrCredentialEncode indicates the credential could not be serialized
	ErrCredentialEncode = errors.New("oidclocal: credential encode failed")
	// ErrCredentialDecode indicates the credential could not be deserialized
	ErrCredentialDecode = errors.New("oidclocal: credential decode failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("oidclocal: result encode failed")
)
