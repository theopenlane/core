package github

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrServerError returns a generic server error
	ErrServerError = errors.New("server error")

	// ErrContextMissingGithubUser is returned when the GitHub user is missing from the context
	ErrContextMissingGithubUser = errors.New("context missing github user")

	// ErrFailedConstructingEndpointURL is returned when URL is invalid and unable to be parsed
	ErrFailedConstructingEndpointURL = errors.New("error constructing URL")

	// ErrCreatingGithubClient is returned when the GitHub client cannot be created
	ErrCreatingGithubClient = errors.New("error creating github client")

	// ErrUnableToGetGithubUser when the user cannot be retrieved from GitHub
	ErrUnableToGetGithubUser = errors.New("unable to get github user")

	// ErrPrimaryEmailNotFound when the user's primary email cannot be retrieved from GitHub
	ErrPrimaryEmailNotFound = errors.New("unable to get primary email address")

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
