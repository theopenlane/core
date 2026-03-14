package providerkit

import "errors"

var (
	// ErrFilterExprInvalid is returned when a filter CEL expression cannot be compiled or the evaluator cannot be initialized
	ErrFilterExprInvalid = errors.New("filter expression invalid")
	// ErrMapExprInvalid is returned when a map CEL expression cannot be compiled or the evaluator cannot be initialized
	ErrMapExprInvalid = errors.New("map expression invalid")
	// ErrFilterExprEval is returned when a filter CEL expression fails during evaluation
	ErrFilterExprEval = errors.New("filter expression evaluation failed")
	// ErrMapExprEval is returned when a map CEL expression fails during evaluation
	ErrMapExprEval = errors.New("map expression evaluation failed")
	// ErrOAuthTokenMissing indicates the OAuth token is not present in the credential set
	ErrOAuthTokenMissing = errors.New("providerkit: oauth token missing")
	// ErrAccessTokenEmpty indicates the access token field is empty
	ErrAccessTokenEmpty = errors.New("providerkit: access token empty")
	// ErrAPITokenMissing indicates the API token is not present in the credential set
	ErrAPITokenMissing = errors.New("providerkit: api token missing")
	// ErrHTTPRequestFailed indicates an HTTP request returned a non-2xx status
	ErrHTTPRequestFailed = errors.New("providerkit: http request failed")
	// ErrOAuthRelyingPartyInit indicates Zitadel relying party construction failed
	ErrOAuthRelyingPartyInit = errors.New("providerkit: oauth relying party initialization failed")
	// ErrOAuthStateGeneration indicates random CSRF state generation failed
	ErrOAuthStateGeneration = errors.New("providerkit: oauth state generation failed")
	// ErrOAuthStateInvalid indicates the stored oauth start state could not be decoded
	ErrOAuthStateInvalid = errors.New("providerkit: oauth state invalid")
	// ErrOAuthStateMismatch indicates the callback state does not match the stored CSRF state
	ErrOAuthStateMismatch = errors.New("providerkit: oauth state mismatch")
	// ErrOAuthCodeMissing indicates the authorization code is absent from the callback input
	ErrOAuthCodeMissing = errors.New("providerkit: oauth callback code missing")
	// ErrOAuthCallbackInputInvalid indicates the callback input could not be decoded
	ErrOAuthCallbackInputInvalid = errors.New("providerkit: oauth callback input invalid")
	// ErrOAuthCodeExchange indicates the authorization code exchange failed
	ErrOAuthCodeExchange = errors.New("providerkit: oauth code exchange failed")
	// ErrOAuthClaimsEncode indicates OIDC claims could not be serialized to a map
	ErrOAuthClaimsEncode = errors.New("providerkit: oauth claims encoding failed")
)

// HTTPRequestError captures metadata for failed HTTP requests
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
