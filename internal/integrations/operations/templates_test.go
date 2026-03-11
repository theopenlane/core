package operations_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestResolveOperationConfig_NoTemplate(t *testing.T) {
	config := types.IntegrationConfig{}

	result, err := operations.ResolveOperationConfig(config, "health.default", nil)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestResolveOperationConfig_NoTemplateWithOverrides(t *testing.T) {
	config := types.IntegrationConfig{}
	overrides := json.RawMessage(`{"key":"value"}`)

	result, err := operations.ResolveOperationConfig(config, "health.default", overrides)
	require.NoError(t, err)
	assert.Equal(t, overrides, result)
}

func TestResolveOperationConfig_WithTemplate(t *testing.T) {
	config := types.IntegrationConfig{
		OperationTemplates: map[string]types.OperationTemplate{
			"health.default": {
				Config: json.RawMessage(`{"timeout":30}`),
			},
		},
	}

	result, err := operations.ResolveOperationConfig(config, "health.default", nil)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestResolveOperationConfig_WithTemplateAndAllowedOverride(t *testing.T) {
	config := types.IntegrationConfig{
		OperationTemplates: map[string]types.OperationTemplate{
			"health.default": {
				Config:         json.RawMessage(`{"timeout":30}`),
				AllowOverrides: []string{"timeout"},
			},
		},
	}
	overrides := json.RawMessage(`{"timeout":60}`)

	result, err := operations.ResolveOperationConfig(config, "health.default", overrides)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestResolveOperationConfig_WithTemplateAndDisallowedOverride(t *testing.T) {
	config := types.IntegrationConfig{
		OperationTemplates: map[string]types.OperationTemplate{
			"health.default": {
				Config:         json.RawMessage(`{"timeout":30}`),
				AllowOverrides: []string{"timeout"},
			},
		},
	}
	overrides := json.RawMessage(`{"secret":"value"}`)

	_, err := operations.ResolveOperationConfig(config, "health.default", overrides)
	assert.ErrorIs(t, err, operations.ErrOperationTemplateOverrideNotAllowed)
}

func TestResolveOperationConfig_TemplateNoOverridesAllowed(t *testing.T) {
	config := types.IntegrationConfig{
		OperationTemplates: map[string]types.OperationTemplate{
			"health.default": {
				Config: json.RawMessage(`{"timeout":30}`),
			},
		},
	}
	overrides := json.RawMessage(`{"timeout":60}`)

	_, err := operations.ResolveOperationConfig(config, "health.default", overrides)
	assert.ErrorIs(t, err, operations.ErrOperationTemplateOverridesNotAllowed)
}
