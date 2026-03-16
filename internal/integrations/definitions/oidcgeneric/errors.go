package oidcgeneric

import "errors"

var (
	// ErrOAuthTokenMissing indicates the OAuth access token is missing
	ErrOAuthTokenMissing = errors.New("oidcgeneric: oauth token missing")
	// ErrClientType indicates the provided client is not the expected type
	ErrClientType = errors.New("oidcgeneric: unexpected client type")
	// ErrUserinfoCallFailed indicates the userinfo call failed
	ErrUserinfoCallFailed = errors.New("oidcgeneric: userinfo call failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("oidcgeneric: result encode failed")
)
