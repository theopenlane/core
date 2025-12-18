package graphapihistory

import (
	"context"
	"errors"
	"strings"

	generated "github.com/theopenlane/core/internal/ent/historygenerated"
	"github.com/theopenlane/core/internal/ent/historygenerated/privacy"
	"github.com/theopenlane/core/internal/graphapi/common"
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
	default:
		logx.FromContext(ctx).Error().Err(err).Msg("unexpected error occurred")

		return err
	}
}
