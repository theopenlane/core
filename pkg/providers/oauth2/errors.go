package oauth2

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrContextMissingToken is returned when the context is missing the token value
	ErrContextMissingToken = errors.New("oauth2: context missing token")

	// ErrContextMissingStateValue is returned when the context is missing the state value
	ErrContextMissingStateValue = errors.New("oauth2: context missing state value")

	// ErrInvalidState is returned when the state parameter is invalid
	ErrInvalidState = errors.New("oauth2: invalid oauth2 state parameter")

	// ErrFailedToGenerateToken is returned when a token cannot be generated
	ErrFailedToGenerateToken = errors.New("failed to generate token")

	// ErrMissingCodeOrState is returned when the request is missing the code or state query string parameter
	ErrMissingCodeOrState = errors.New("oauth2: request missing code or state")

	// ErrContextMissingErrorValue is returned when the context does not have an error value
	ErrContextMissingErrorValue = fmt.Errorf("context missing error value")
)

// DefaultFailureHandler responds with a 400 status code and message parsed from the context
var DefaultFailureHandler = http.HandlerFunc(failureHandler)

func failureHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	err := ErrorFromContext(ctx)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// ErrorFromContext always returns some non-nil error
	http.Error(w, "", http.StatusBadRequest)
}
