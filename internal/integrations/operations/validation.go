package operations

import (
	"encoding/json"
	"strings"

	"github.com/theopenlane/core/pkg/jsonx"
)

var emptyJSONObject = json.RawMessage(`{}`)

// ConfigValidationError captures JSON schema validation details
type ConfigValidationError struct {
	// Issues contains the validation error messages
	Issues []string
}

// Error formats operation config validation failures
func (e *ConfigValidationError) Error() string {
	if e == nil || len(e.Issues) == 0 {
		return ErrOperationConfigInvalid.Error()
	}

	return ErrOperationConfigInvalid.Error() + ": " + strings.Join(e.Issues, "; ")
}

// Unwrap enables errors.Is(err, ErrOperationConfigInvalid)
func (e *ConfigValidationError) Unwrap() error {
	return ErrOperationConfigInvalid
}

// ValidateConfig validates operation config against a descriptor-provided JSON schema
func ValidateConfig(schema json.RawMessage, config json.RawMessage) error {
	if len(schema) == 0 {
		return nil
	}

	doc := config
	if len(doc) == 0 {
		doc = emptyJSONObject
	}

	result, err := jsonx.ValidateSchema(schema, doc)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	}

	return &ConfigValidationError{
		Issues: jsonx.ValidationErrorStrings(result),
	}
}
