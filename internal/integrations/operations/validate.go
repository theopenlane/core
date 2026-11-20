package operations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/logx"
	"github.com/theopenlane/utils/jsonx"
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

// ConformStored upgrades a stored document from the name it was persisted under, strips and defaults it against the layout schema, then validates it
func ConformStored(ctx context.Context, req types.InstallationRequest, layout types.InputRegistration, from string, stored json.RawMessage, sentinel error) (json.RawMessage, error) {
	document := stored

	if layout.Upgrade != nil {
		upgraded, err := layout.Upgrade(ctx, req, from, stored)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed upgrading stored document")

			return nil, fmt.Errorf("%w: upgrade: %w", sentinel, err)
		}

		document = upgraded
	}

	conformed, err := jsonx.ConformToSchema(layout.Schema, document)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed conforming stored document to schema")

		return nil, fmt.Errorf("%w: conform to schema: %w", sentinel, err)
	}

	if err := ValidateInput(ctx, req, layout.Schema, layout.Validate, conformed, sentinel); err != nil {
		return nil, err
	}

	return conformed, nil
}
