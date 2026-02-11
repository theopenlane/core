package schematest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Properties extracts schema properties and fails the test if missing
func Properties(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()

	if props := FindMap(schema, "properties"); props != nil {
		return props
	}

	require.Fail(t, "expected properties to be a map")
	return nil
}

// Required extracts required fields from a schema and returns them as strings
func Required(t *testing.T, schema map[string]any) []string {
	t.Helper()

	raw := FindSlice(schema, "required")
	if raw == nil {
		return nil
	}

	values, ok := raw.([]any)
	if !ok {
		return nil
	}

	out := make([]string, 0, len(values))
	for _, value := range values {
		if name, ok := value.(string); ok {
			out = append(out, name)
		}
	}

	return out
}

// FindMap searches nested schema content for a map value
func FindMap(node any, key string) map[string]any {
	switch typed := node.(type) {
	case map[string]any:
		if val, ok := typed[key].(map[string]any); ok {
			return val
		}

		for _, value := range typed {
			if found := FindMap(value, key); found != nil {
				return found
			}
		}
	case []any:
		for _, value := range typed {
			if found := FindMap(value, key); found != nil {
				return found
			}
		}
	}

	return nil
}

// FindSlice searches nested schema content for a slice value
func FindSlice(node any, key string) any {
	switch typed := node.(type) {
	case map[string]any:
		if val, ok := typed[key]; ok {
			return val
		}

		for _, value := range typed {
			if found := FindSlice(value, key); found != nil {
				return found
			}
		}
	case []any:
		for _, value := range typed {
			if found := FindSlice(value, key); found != nil {
				return found
			}
		}
	}

	return nil
}
