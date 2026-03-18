package shortlinks

import "errors"

var (
	// ErrMissingAuthenticationParams indicates a missing client ID or secret
	ErrMissingAuthenticationParams = errors.New("shortlinks: missing client ID or secret")
	// ErrMissingURL indicates a missing URL
	ErrMissingURL = errors.New("shortlinks: missing URL")
	// ErrEmptyResponse indicates the API returned a nil response
	ErrEmptyResponse = errors.New("shortlinks: empty response")
	// ErrEmptyResponseBody indicates the API returned an empty body
	ErrEmptyResponseBody = errors.New("shortlinks: empty response body")
	// ErrMissingShortURL indicates the response did not contain a short URL
	ErrMissingShortURL = errors.New("shortlinks: missing short URL in response")
	// ErrClientNotConfigured indicates the shortlinks client is not configured
	ErrClientNotConfigured = errors.New("shortlinks: client not configured")
)
