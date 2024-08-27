package google

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrServerError returns a generic server error
	ErrServerError = errors.New("server error")

	// ErrContextMissingGoogleUser  is returned when the Google user is missing from the context
	ErrContextMissingGoogleUser = errors.New("context missing google user")

	// ErrFailedConstructingEndpointURL is returned when URL is invalid and unable to be parsed
	ErrFailedConstructingEndpointURL = errors.New("error constructing URL")

	// ErrUnableToGetGoogleUser when the user cannot be retrieved from Google
	ErrUnableToGetGoogleUser = errors.New("unable to get google user")

	// ErrCannotValidateGoogleUser when the Google user is invalid
	ErrCannotValidateGoogleUser = errors.New("could not validate google user")

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
