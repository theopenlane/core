package helpers

import "errors"

var (
	// ErrOAuthTokenMissing indicates the OAuth token is not present in the credential payload
	ErrOAuthTokenMissing = errors.New("helpers: oauth token missing")
	// ErrAccessTokenEmpty indicates the access token field is empty
	ErrAccessTokenEmpty = errors.New("helpers: access token empty")
	// ErrAPITokenMissing indicates the API token is not present in the credential payload
	ErrAPITokenMissing = errors.New("helpers: api token missing")
	// ErrHTTPRequestFailed indicates an HTTP request returned a non-2xx status
	ErrHTTPRequestFailed = errors.New("helpers: http request failed")
)
