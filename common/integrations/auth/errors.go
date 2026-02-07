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
	// URL is the URL that was requested
	URL string
	// Status is the HTTP status text returned by the request
	Status string
	// StatusCode is the HTTP status code returned by the request
	StatusCode int
	// Body is the response body returned by the request, if any
	Body string
}

// Error returns a formatted error message for the HTTP request failure
func (e *HTTPRequestError) Error() string {
	return ErrHTTPRequestFailed.Error()
}

// Unwrap allows errors.Is and errors.As to work with HTTPRequestError
func (e *HTTPRequestError) Unwrap() error {
	return ErrHTTPRequestFailed
}
