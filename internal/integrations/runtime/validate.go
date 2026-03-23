package runtime

import (
	"encoding/json"

	"github.com/theopenlane/core/pkg/jsonx"
)

// validatePayload validates data against a JSON schema, returning the sentinel error when validation fails
func validatePayload(schema, data json.RawMessage, sentinel error) error {
	if len(schema) == 0 {
		return nil
	}

	result, err := jsonx.ValidateSchema(schema, data)
	if err != nil {
		return err
	}

	if !result.Valid() {
		return sentinel
	}

	return nil
}
