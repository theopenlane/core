package oidcgeneric

import "errors"

var (
	// ErrOAuthTokenMissing indicates the OAuth access token is missing
	ErrOAuthTokenMissing = errors.New("oidcgeneric: oauth token missing")
	// ErrClientType indicates the provided client is not the expected type
	ErrClientType = errors.New("oidcgeneric: unexpected client type")
)
