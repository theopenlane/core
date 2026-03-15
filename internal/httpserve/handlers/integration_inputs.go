package handlers

import (
	"context"
	"encoding/json"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/httpserve/handlers/internal/jsonschemautil"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

func normalizeUserInput(ctx context.Context, registration *types.UserInputRegistration, raw json.RawMessage) (json.RawMessage, error) {
	if registration == nil || len(raw) == 0 {
		return jsonx.CloneRawMessage(raw), nil
	}

	normalized := jsonx.CloneRawMessage(raw)
	if registration.Normalize != nil {
		next, err := registration.Normalize(ctx, jsonx.CloneRawMessage(raw))
		if err != nil {
			return nil, err
		}
		if len(next) > 0 {
			normalized = next
		}
	}

	if err := validateRawSchema(registration.Schema, normalized); err != nil {
		return nil, err
	}

	if registration.Validate == nil {
		return normalized, nil
	}

	validated, err := registration.Validate(ctx, jsonx.CloneRawMessage(normalized))
	if err != nil {
		return nil, err
	}
	if len(validated) == 0 {
		return normalized, nil
	}

	return validated, nil
}

func normalizeCredential(ctx context.Context, registration *types.CredentialRegistration, installation *ent.Integration, raw json.RawMessage) (types.CredentialSet, error) {
	credential := types.CredentialSet{ProviderData: jsonx.CloneRawMessage(raw)}
	if registration == nil {
		return credential, nil
	}

	var err error
	if registration.Normalize != nil {
		credential, err = registration.Normalize(ctx, installation, credential)
		if err != nil {
			return types.CredentialSet{}, err
		}
	}

	if err := validateRawSchema(registration.Schema, credential.ProviderData); err != nil {
		return types.CredentialSet{}, err
	}

	if registration.Validate != nil {
		credential, err = registration.Validate(ctx, installation, credential)
		if err != nil {
			return types.CredentialSet{}, err
		}
	}

	return credential, nil
}

func validateRawSchema(schema json.RawMessage, raw json.RawMessage) error {
	if len(schema) == 0 {
		return nil
	}

	document := map[string]any{}
	if len(raw) > 0 {
		decoded, err := jsonx.ToMap(raw)
		if err != nil {
			return err
		}
		if decoded != nil {
			document = decoded
		}
	}

	result, err := jsonx.ValidateSchema(schema, document)
	if err != nil {
		return err
	}

	fieldErrs := jsonschemautil.FieldErrorsFromResult(result)
	if len(fieldErrs) > 0 {
		return fieldErrs[0]
	}

	return nil
}
