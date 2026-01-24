package workflows

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Avoid reflect-based nil checks on hot paths; treat marshaled "null" as a typed-nil sentinel
var jsonNull = []byte("null")

// StringField extracts a string field by name from a struct or pointer to struct
// It returns an error when decoding fails
func StringField(node any, field string) (string, error) {
	if node == nil {
		return "", ErrStringFieldNil
	}

	data, err := json.Marshal(node)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrStringFieldMarshal, err)
	}
	if bytes.Equal(bytes.TrimSpace(data), jsonNull) {
		return "", ErrStringFieldNil
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return "", fmt.Errorf("%w: %w", ErrStringFieldUnmarshal, err)
	}

	raw, ok := m[field]
	if !ok || raw == nil {
		return "", nil
	}

	switch v := raw.(type) {
	case string:
		return v, nil
	case *string:
		if v == nil {
			return "", nil
		}
		return *v, nil
	case []byte:
		return string(v), nil
	default:
		return "", nil
	}
}

// StringSliceField extracts a string slice field by name from a struct or pointer to struct
// It returns an error when decoding fails
func StringSliceField(node any, field string) ([]string, error) {
	if node == nil {
		return nil, ErrStringFieldNil
	}

	data, err := json.Marshal(node)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStringFieldMarshal, err)
	}
	if bytes.Equal(bytes.TrimSpace(data), jsonNull) {
		return nil, ErrStringFieldNil
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStringFieldUnmarshal, err)
	}

	raw, ok := m[field]
	if !ok || raw == nil {
		return nil, nil
	}

	items, ok := raw.([]any)
	if !ok {
		return nil, nil
	}

	out := make([]string, 0, len(items))
	for _, item := range items {
		value, ok := item.(string)
		if !ok {
			return nil, ErrStringSliceFieldInvalid
		}
		out = append(out, value)
	}

	return out, nil
}
