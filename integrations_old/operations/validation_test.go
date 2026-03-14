package operations_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/operations"
)

func TestValidateConfig_NoSchema(t *testing.T) {
	err := operations.ValidateConfig(nil, json.RawMessage(`{"key":"value"}`))
	require.NoError(t, err)
}

func TestValidateConfig_ValidConfig(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}`)
	config := json.RawMessage(`{"name":"test"}`)

	err := operations.ValidateConfig(schema, config)
	require.NoError(t, err)
}

func TestValidateConfig_InvalidConfig(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}`)
	config := json.RawMessage(`{}`)

	err := operations.ValidateConfig(schema, config)
	assert.ErrorIs(t, err, operations.ErrOperationConfigInvalid)

	var validationErr *operations.ConfigValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.NotEmpty(t, validationErr.Issues)
}

func TestValidateConfig_EmptyConfig(t *testing.T) {
	schema := json.RawMessage(`{"type":"object"}`)

	err := operations.ValidateConfig(schema, nil)
	require.NoError(t, err)
}
