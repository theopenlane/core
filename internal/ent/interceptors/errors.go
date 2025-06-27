package interceptors

import (
	"errors"
)

var (
	// ErrInternalServerError is returned when an internal error occurs.
	ErrInternalServerError = errors.New("internal server error")
	// ErrUnableToRetrieveUserID is returned when the user cannot be retrieved from the context
	ErrUnableToRetrieveUserID = errors.New("unable to retrieve user from context")
	// ErrRetrievingObjects is returned when an error occurs while retrieving objects
	ErrRetrievingObjects = errors.New("error retrieving objects")
	// ErrFeatureNotEnabled is returned when a requested feature is not enabled for the organization
	ErrFeatureNotEnabled = errors.New("feature not enabled for organization")
)
