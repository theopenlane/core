package graphapi

import (
	"errors"
	"fmt"
	"strings"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/utils/rout"
	"go.uber.org/zap"
)

var (
	// ErrInternalServerError is returned when an internal error occurs.
	ErrInternalServerError = errors.New("internal server error")

	// ErrPermissionDenied is returned when the user is not authorized to perform the requested query or mutation
	ErrPermissionDenied = errors.New("you are not authorized to perform this action")

	// ErrCascadeDelete is returned when an error occurs while performing cascade deletes on associated objects
	ErrCascadeDelete = errors.New("error deleting associated objects")

	// ErrSubscriberNotFound is returned when a subscriber is not found
	ErrSubscriberNotFound = errors.New("subscriber not found")

	// ErrSearchFailed is returned when the search operation fails
	ErrSearchFailed = errors.New("search failed, please try again")
)

// PermissionDeniedError is returned when user is not authorized to perform the requested query or mutation
type PermissionDeniedError struct {
	Action     string
	ObjectType string
}

// Error returns the PermissionDeniedError in string format
func (e *PermissionDeniedError) Error() string {
	return fmt.Sprintf("you are not authorized to perform this action: %s on %s", e.Action, e.ObjectType)
}

// newPermissionDeniedError returns a PermissionDeniedError
func newPermissionDeniedError(a string, o string) *PermissionDeniedError {
	return &PermissionDeniedError{
		Action:     a,
		ObjectType: o,
	}
}

func newCascadeDeleteError(err error) error {
	return fmt.Errorf("%w: %v", ErrCascadeDelete, err)
}

// AlreadyExistsError is returned when an object already exists
type AlreadyExistsError struct {
	ObjectType string
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

// ForeignKeyError is returned when an object does not exist in the related table
type ForeignKeyError struct {
	Action     string
	ObjectType string
}

// Error returns the ForeignKeyError in string format
func (e *ForeignKeyError) Error() string {
	return fmt.Sprintf("constraint failed, unable to complete the action '%s' because the record '%s' does not exist. please try again", e.Action, e.ObjectType)
}

// newForeignKeyError returns a ForeignKeyError
func newForeignKeyError(action, objecttype string) *ForeignKeyError {
	return &ForeignKeyError{
		Action:     action,
		ObjectType: objecttype,
	}
}

// parseRequestError logs and parses the error and returns the appropriate error type for the client
// TODO: cleanup return error messages
func parseRequestError(err error, a action, logger *zap.SugaredLogger) error {
	// log the error for debugging
	logger.Errorw("error processing request", "action", a.action, "object", a.object, "error", err)

	switch {
	case generated.IsValidationError(err):
		validationError := err.(*generated.ValidationError)

		logger.Debugw("validation error", "field", validationError.Name, "error", validationError.Error())

		return validationError
	case generated.IsConstraintError(err):
		constraintError := err.(*generated.ConstraintError)

		logger.Debugw("constraint error", "error", constraintError.Error())

		// Check for unique (or UNIQUE in sqlite) constraint error
		if strings.Contains(strings.ToLower(constraintError.Error()), "unique") {
			return newAlreadyExistsError(a.object)
		}

		// Check for foreign key constraint error
		if rout.IsForeignKeyConstraintError(constraintError) {
			return newForeignKeyError(a.action, a.object)
		}

		return constraintError
	case generated.IsNotFound(err):
		logger.Debugw("not found", "error", err.Error())

		return err
	case errors.Is(err, privacy.Deny):
		logger.Debugw("permission denied", "error", err.Error())

		return newPermissionDeniedError(a.action, a.object)
	default:
		logger.Errorw("unexpected error", "error", err.Error())

		return err
	}
}
