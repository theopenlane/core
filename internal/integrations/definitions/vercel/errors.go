package vercel

import "errors"

var (
	// ErrAPITokenMissing indicates the Vercel API token is missing from the credential
	ErrAPITokenMissing = errors.New("vercel: api token missing")
	// ErrClientType indicates the provided client is not the expected HTTP client
	ErrClientType = errors.New("vercel: unexpected client type")
)
