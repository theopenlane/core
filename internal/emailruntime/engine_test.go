package emailruntime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateTemplateData_EmptySchema(t *testing.T) {
	err := validateTemplateData(nil, map[string]any{"foo": "bar"})
	require.NoError(t, err)
}

func TestValidateTemplateData_EmptyPayload(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
	}

	err := validateTemplateData(schema, map[string]any{})
	require.NoError(t, err)
}

func TestValidateTemplateData_ValidData(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
		"required": []any{"name"},
	}

	err := validateTemplateData(schema, map[string]any{"name": "Alice"})
	require.NoError(t, err)
}

func TestValidateTemplateData_MissingRequired(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
		"required": []any{"name"},
	}

	err := validateTemplateData(schema, map[string]any{})
	require.ErrorIs(t, err, ErrTemplateDataInvalid)
}

func TestValidateTemplateData_WrongType(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"count": map[string]any{"type": "integer"},
		},
		"required": []any{"count"},
	}

	err := validateTemplateData(schema, map[string]any{"count": "not-an-int"})
	require.ErrorIs(t, err, ErrTemplateDataInvalid)
}
