package testclient

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

var (
	// ErrFailedToGetOauthToken is returned when the client fails to get an OAuth2 token
	ErrFailedToGetOauthToken = errors.New("failed to get oauth2 token")
	// ErrNoCookieJarSet is returned when the client does not have a cookie jar, cannot set cookies
	ErrNoCookieJarSet = errors.New("client does not have a cookie jar, cannot set cookies")
	// ErrEmptyCSRFToken is returned when an empty CSRF token is received from the server
	ErrEmptyCSRFToken = errors.New("empty csrf token received from server, cannot continue")
	// ErrCSRFCookieNotFound is returned when the CSRF cookie is not found in the cookie jar
	ErrCSRFCookieNotFound = errors.New("csrf cookie not found in cookie jar, cannot continue")
)

// AuthenticationError is returned when a user cannot be authenticated
type AuthenticationError struct {
	// StatusCode is the http response code that was returned
	StatusCode int
	// Body of the response
	Body string
}

// Error returns the AuthenticationError in string format
func (e *AuthenticationError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("unable to authenticate (status %d)", e.StatusCode)
	}

	return fmt.Sprintf("unable to authenticate (status %d): %s", e.StatusCode, strings.ToLower(e.Body))
}

// newAuthenticationError returns an error when authentication fails
func newAuthenticationError(statusCode int, body string) *AuthenticationError {
	return &AuthenticationError{
		StatusCode: statusCode,
		Body:       body,
	}
}

// RequestError is a generic error when a request with the client fails
type RequestError struct {
	// StatusCode is the http response code that was returned
	StatusCode int
	// Body of the response
	Body string
}

// Error returns the RequestError in string format
func (e *RequestError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("unable to process request (status %d)", e.StatusCode)
	}

	return fmt.Sprintf("unable to process request (status %d): %s", e.StatusCode, strings.ToLower(e.Body))
}

// newRequestError returns an error when a client request fails
func newRequestError(statusCode int, body string) *RequestError {
	return &RequestError{
		StatusCode: statusCode,
		Body:       body,
	}
}
