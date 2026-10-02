package operations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// ValidateInput checks a payload against its schema and then its semantic validation, wrapping schema failures in sentinel
func ValidateInput(ctx context.Context, req types.InstallationRequest, schema json.RawMessage, validate types.ValidateFunc, payload json.RawMessage, sentinel error) error {
	err := jsonx.Validate(schema, payload)

	var schemaErr *jsonx.SchemaError

	switch {
	case errors.As(err, &schemaErr):
		logx.FromContext(ctx).Info().Strs("issues", schemaErr.Issues).Msg("schema validation failed")

		return fmt.Errorf("%w: %w", sentinel, err)
	case err != nil:
		return err
	case validate == nil:
		return nil
	}

	if err := validate(ctx, req, payload); err != nil {
		return fmt.Errorf("%w: %w", sentinel, err)
	}

	return nil
}
