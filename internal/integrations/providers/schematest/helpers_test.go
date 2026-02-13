package schematest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProperties(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"age":  map[string]any{"type": "integer"},
		},
	}

	props := Properties(t, schema)
	require.Contains(t, props, "name")
	require.Contains(t, props, "age")
}

func TestPropertiesNested(t *testing.T) {
	schema := map[string]any{
		"allOf": []any{
			map[string]any{
				"properties": map[string]any{
					"nested": map[string]any{"type": "string"},
				},
			},
		},
	}

	props := Properties(t, schema)
	require.Contains(t, props, "nested")
}

func TestRequired(t *testing.T) {
	schema := map[string]any{
		"type":     "object",
		"required": []any{"name", "email"},
	}

	required := Required(t, schema)
	require.Contains(t, required, "name")
	require.Contains(t, required, "email")
}

func TestRequiredNested(t *testing.T) {
	schema := map[string]any{
		"allOf": []any{
			map[string]any{
				"required": []any{"nested_field"},
			},
		},
	}

	required := Required(t, schema)
	require.Contains(t, required, "nested_field")
}

func TestFindMapNotFound(t *testing.T) {
	schema := map[string]any{
		"type": "object",
	}

	result := FindMap(schema, "properties")
	require.Nil(t, result)
}

func TestFindSliceNotFound(t *testing.T) {
	schema := map[string]any{
		"type": "object",
	}

	result := FindSlice(schema, "required")
	require.Nil(t, result)
}
