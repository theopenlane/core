package graphapi

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
)

var (
	// ErrInternalServerError is returned when an internal error occurs.
	ErrInternalServerError = errors.New("internal server error")

	// ErrCascadeDelete is returned when an error occurs while performing cascade deletes on associated objects
	ErrCascadeDelete = errors.New("error deleting associated objects")

	// ErrSearchFailed is returned when the search operation fails
	ErrSearchFailed = errors.New("search failed, please try again")

	// ErrSearchQueryTooShort is returned when the search query is too short
	ErrSearchQueryTooShort = errors.New("search query is too short, please enter a longer search query")

	// ErrNoOrganizationID is returned when the organization ID is not provided
	ErrNoOrganizationID = errors.New("unable to determine organization ID in request")

	// ErrUnableToDetermineObjectType is returned when the object type up the parent upload object cannot be determined
	ErrUnableToDetermineObjectType = errors.New("unable to determine parent object type")

	// ErrResourceNotAccessibleWithToken is returned when a resource is not accessible with a personal access token or api token
	ErrResourceNotAccessibleWithToken = errors.New("resource is not accessible with token authentication")
)

var _ gqlerrors.CustomErrorType = (*NotFoundError)(nil)

// NotFoundError is returned when the requested object is not found
type NotFoundError struct {
	ObjectType string
}

// Code returns the NotFoundError code
func (e NotFoundError) Code() string {
	return gqlerrors.NotFoundErrorCode
}

// Error returns the NotFoundError in string format
func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.ObjectType)
}

// Message returns the NotFoundError in string format
func (e NotFoundError) Message() string {
	return fmt.Sprintf("%s not found", e.ObjectType)
}

// newPermissionDeniedError returns a NotFoundError
func newNotFoundError(o string) NotFoundError {
	return NotFoundError{
		ObjectType: o,
	}
}

var _ gqlerrors.CustomErrorType = (*NotAuthorizedError)(nil)

// NotAuthorizedError is returned when the user is not authorized to perform the action
type NotAuthorizedError struct{}

// Code returns the NotAuthorizedError code
func (e NotAuthorizedError) Code() string {
	return gqlerrors.UnauthorizedErrorCode
}

// Error returns the NotAuthorizedError in string format
func (e NotAuthorizedError) Error() string {
	return generated.ErrPermissionDenied.Error()
}

// Message returns the NotAuthorizedError in string format
func (e NotAuthorizedError) Message() string {
	return "you do not have permission to perform this action, please contact your organization owner"
}

// newPermissionDeniedError returns a NotAuthorizedError
func newPermissionDeniedError() NotAuthorizedError {
	return NotAuthorizedError{}
}

func newCascadeDeleteError(err error) error {
	return fmt.Errorf("%w: %v", ErrCascadeDelete, err)
}

var _ gqlerrors.CustomErrorType = (*AlreadyExistsError)(nil)

// AlreadyExistsError is returned when an object already exists
type AlreadyExistsError struct {
	ObjectType string
}

// Code returns the AlreadyExistsError code
func (e *AlreadyExistsError) Code() string {
	return gqlerrors.AlreadyExistsErrorCode
}

// Message returns the AlreadyExistsError message
func (e *AlreadyExistsError) Message() string {
	return fmt.Sprintf("%s already exists in the system", e.ObjectType)
}

// Error returns the AlreadyExistsError in string format
func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("%s already exists", e.ObjectType)
}

// newAlreadyExistsError returns a AlreadyExistsError
func newAlreadyExistsError(o string) *AlreadyExistsError {
	return &AlreadyExistsError{
		ObjectType: o,
	}
}

type action struct {
	object string
	action string
}

var _ gqlerrors.CustomErrorType = (*ForeignKeyError)(nil)

// ForeignKeyError is returned when an object does not exist in the related table
type ForeignKeyError struct {
	Action     string
	ObjectType string
}

// Code returns the ForeignKeyError code
func (e *ForeignKeyError) Code() string {
	return gqlerrors.ConflictErrorCode
}

// Message returns the ForeignKeyError message
func (e *ForeignKeyError) Message() string {
	return "invalid input provided, unable complete the request"
}

// Error returns the ForeignKeyError in string format
func (e *ForeignKeyError) Error() string {
	if e.ObjectType == "" {
		return fmt.Sprintf("constraint failed, unable to complete the %s", e.Action)
	}

	return fmt.Sprintf("constraint failed, unable to complete the %s because the '%s' record does not exist", e.Action, e.ObjectType)
}

// newForeignKeyError returns a ForeignKeyError
func newForeignKeyError(action, objecttype string) *ForeignKeyError {
	return &ForeignKeyError{
		Action:     action,
		ObjectType: objecttype,
	}
}

var _ gqlerrors.CustomErrorType = (*ValidationError)(nil)

// ValidationError is returned when a field fails validation
type ValidationError struct {
	ErrMsg string
}

// Code returns the ValidationError code
func (e *ValidationError) Code() string {
	return gqlerrors.ValidationErrorCode
}

// Message returns the ValidationError message
func (e *ValidationError) Message() string {
	return fmt.Sprintf("invalid input provided: %s", e.ErrMsg)
}

// Error returns the ValidationError in string format, by removing the "generated: " prefix
func (e *ValidationError) Error() string {
	return strings.ReplaceAll(e.ErrMsg, "generated: ", "")
}

// newValidationError returns a ValidationError
func newValidationError(errMsg string) *ValidationError {
	return &ValidationError{
		ErrMsg: errMsg,
	}
}

// parseRequestError logs and parses the error and returns the appropriate error type for the client
func parseRequestError(err error, a action) error {
	// log the error for debugging
	log.Error().
		Err(err).
		Str("action", a.action).
		Str("object", a.object).
		Msg("error processing request")

	switch {
	case generated.IsValidationError(err):
		validationError := err.(*generated.ValidationError)

		log.Debug().
			Err(validationError).
			Str("field", validationError.Name).
			Msg("validation error")

		return newValidationError(validationError.Error())
	case generated.IsConstraintError(err):
		constraintError := err.(*generated.ConstraintError)

		log.Debug().Err(constraintError).Msg("constraint error")

		// Check for unique constraint error
		if strings.Contains(strings.ToLower(constraintError.Error()), "unique") {
			return newAlreadyExistsError(a.object)
		}

		// Check for foreign key constraint error
		if rout.IsForeignKeyConstraintError(constraintError) {
			object := getConstraintField(constraintError, a.object)

			return newForeignKeyError(a.action, object)
		}

		return constraintError
	case generated.IsNotFound(err):
		log.Debug().Err(err).Msg("request object was not found")

		return newNotFoundError(a.object)
	case errors.Is(err, privacy.Deny):
		log.Debug().Err(err).Msg("user has no access to the requested object")

		return newNotFoundError(a.object)
	case errors.Is(err, generated.ErrPermissionDenied):
		log.Debug().Err(err).Msg("user has no access to the requested object")

		return newPermissionDeniedError()
	default:
		log.Error().Err(err).Msg("unexpected error occurred")

		return err
	}
}

// getConstraintField returns the field that caused the constraint error
// this currently only works for postgres errors, which is the supported database of this project
func getConstraintField(err error, object string) string {
	unwrappedErr := errors.Unwrap(err)     // Unwrap the error to get the underlying error
	dbError := errors.Unwrap(unwrappedErr) // Unwrap again to get the postgres error

	if postgresError, ok := dbError.(*pq.Error); ok {
		// the Table will be the object_edge so we need to remove the object_ prefix
		return strings.ReplaceAll(postgresError.Table, object+"_", "")
	}

	return ""
}
