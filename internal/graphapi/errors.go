package graphapi

import (
	"context"
	"errors"
	"strings"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/utils/rout"
)

// parseRequestError logs and parses the error and returns the appropriate error type for the client
func parseRequestError(ctx context.Context, err error, a common.Action) error {
	// log the error for debugging
	logx.FromContext(ctx).Error().
		Err(err).
		Str("action", a.Action).
		Str("object", a.Object).
		Msg("error processing request")

	switch {
	case generated.IsValidationError(err):
		validationError := err.(*generated.ValidationError)

		logx.FromContext(ctx).Info().
			Err(validationError).
			Str("field", validationError.Name).
			Msg("validation error")

		if strings.Contains(strings.ToLower(validationError.Error()), "generated:") {
			numParts := 2

			errMsg := strings.SplitN(validationError.Error(), "generated: ", numParts)
			if len(errMsg) == numParts {
				return common.NewValidationError(errMsg[1])
			}
		}

		return common.NewValidationError(validationError.Error())
	case generated.IsConstraintError(err):
		constraintError := err.(*generated.ConstraintError)

		logx.FromContext(ctx).Info().Err(constraintError).Msg("constraint error")

		// Check for unique constraint error
		if strings.Contains(strings.ToLower(constraintError.Error()), "unique") {
			return common.NewAlreadyExistsError(a.Object)
		}

		// Check for foreign key constraint error
		if rout.IsForeignKeyConstraintError(constraintError) {
			object := common.GetConstraintField(constraintError, a.Object)

			return common.NewForeignKeyError(a.Action, object)
		}

		return constraintError
	case generated.IsNotFound(err):
		logx.FromContext(ctx).Info().Err(err).Msg("request object was not found")

		return common.NewNotFoundError(a.Object)
	case errors.Is(err, privacy.Deny):
		logx.FromContext(ctx).Info().Err(err).Msg("user has no access to the requested object due to privacy rules")

		return common.NewNotFoundError(a.Object)
	case errors.Is(err, generated.ErrPermissionDenied):
		logx.FromContext(ctx).Info().Err(err).Msg("user has no access to the requested object due to permissions")

		return newPermissionDeniedError()
	default:
		logx.FromContext(ctx).Error().Err(err).Msg("unexpected error occurred")

		return err
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

// Module implements the CustomErrorType interface
func (e NotAuthorizedError) Module() models.OrgModule {
	return ""
}

// newPermissionDeniedError returns a NotAuthorizedError
func newPermissionDeniedError() NotAuthorizedError {
	return NotAuthorizedError{}
}

var ErrStandardNotFound = errors.New("standard not found")
