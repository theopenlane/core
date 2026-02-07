package auth

import "errors"

var (
	// ErrAuthenticatedClientNil indicates the authenticated client is nil.
	ErrAuthenticatedClientNil = errors.New("auth: authenticated client is nil")
	// ErrOAuthTokenMissing indicates the OAuth token is not present in the credential payload.
	ErrOAuthTokenMissing = errors.New("auth: oauth token missing")
	// ErrAccessTokenEmpty indicates the access token field is empty.
	ErrAccessTokenEmpty = errors.New("auth: access token empty")
	// ErrAPITokenMissing indicates the API token is not present in the credential payload.
	ErrAPITokenMissing = errors.New("auth: api token missing")
	// ErrHTTPRequestFailed indicates an HTTP request returned a non-2xx status.
	ErrHTTPRequestFailed = errors.New("auth: http request failed")
	// ErrRandomStateGeneration indicates random state generation failed.
	ErrRandomStateGeneration = errors.New("auth: random state generation failed")
	// ErrDecodeProviderDataTargetNil indicates provider data decoding target is nil.
	ErrDecodeProviderDataTargetNil = errors.New("auth: decode provider data target is nil")
)

// HTTPRequestError captures metadata for failed HTTP requests.
type HTTPRequestError struct {
	URL        string
	Status     string
	StatusCode int
	Body       string
}

func (e *HTTPRequestError) Error() string {
	return ErrHTTPRequestFailed.Error()
}

func (e *HTTPRequestError) Unwrap() error {
	return ErrHTTPRequestFailed
}
