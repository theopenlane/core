package cloudflare

import "errors"

var (
	// ErrAPITokenMissing indicates the Cloudflare API token is missing from the credential
	ErrAPITokenMissing = errors.New("cloudflare: api token missing")
	// ErrTokenVerificationFailed indicates the Cloudflare token verification failed
	ErrTokenVerificationFailed = errors.New("cloudflare: token verification failed")
	// ErrTokenNotActive indicates the Cloudflare token is not in an active state
	ErrTokenNotActive = errors.New("cloudflare: token is not active")
	// ErrClientType indicates the provided client is not a Cloudflare client
	ErrClientType = errors.New("cloudflare: unexpected client type")
)
