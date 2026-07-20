package runtime

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// validatePayload validates data against a JSON schema, returning the sentinel error when validation fails
func validatePayload(ctx context.Context, schema, data json.RawMessage, sentinel error) error {
	if len(schema) == 0 {
		return nil
	}

	result, err := jsonx.ValidateSchema(schema, data)
	if err != nil {
		return err
	}

	if !result.Valid() {
		logger := logx.FromContext(ctx).Info()
		for _, resultErr := range result.Errors() {
			logger = logger.Str(resultErr.Field(), resultErr.Description())
		}

		logger.Msg("schema validation failed")

		return sentinel
	}

	return nil
}
