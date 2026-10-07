package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// finalizeDefinition wraps stored-input validation
func finalizeDefinition(def types.Definition) (types.Definition, error) {
	def.Operations = finalizeOperations(def)

	return def, nil
}

// finalizeOperations wraps each stored-input operation's validation with the uniform filter expression check
func finalizeOperations(def types.Definition) []types.OperationRegistration {
	operations := slices.Clone(def.Operations)

	for i := range operations {
		operation := &operations[i]

		if !operation.Stored {
			continue
		}

		operation.Input.Validate = validateOperationInput(operation.Input.Validate)
	}

	return operations
}

// validateOperationInput compiles the stored filter expression before running the definition's own validation
func validateOperationInput(next types.ValidateFunc) types.ValidateFunc {
	return func(ctx context.Context, req types.InstallationRequest, payload json.RawMessage) error {
		settings, err := types.OperationSettingsFrom(payload)
		if err != nil {
			return err
		}

		if settings.FilterExpr != "" {
			if err := providerkit.ValidateExpr(settings.FilterExpr); err != nil {
				return fmt.Errorf("%w: %w", ErrOperationFilterExprInvalid, err)
			}
		}

		if next == nil {
			return nil
		}

		return next(ctx, req, payload)
	}
}
