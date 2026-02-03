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
	// ErrRandomStateGeneration indicates random state generation failed
	ErrRandomStateGeneration = errors.New("helpers: random state generation failed")
	// ErrDecodeConfigTargetNil indicates DecodeConfig was called with a nil target
	ErrDecodeConfigTargetNil = errors.New("helpers: decode config target is nil")
	// ErrOperationTemplateRequired indicates the operation requires a stored template configuration
	ErrOperationTemplateRequired = errors.New("helpers: operation template required")
	// ErrOperationTemplateOverridesNotAllowed indicates overrides are not permitted for a template
	ErrOperationTemplateOverridesNotAllowed = errors.New("helpers: operation template overrides not allowed")
	// ErrOperationTemplateOverrideNotAllowed indicates a provided override key is not permitted
	ErrOperationTemplateOverrideNotAllowed = errors.New("helpers: operation template override not allowed")
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
