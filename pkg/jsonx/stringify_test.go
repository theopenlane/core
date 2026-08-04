package jsonx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringify(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{name: "nil", value: nil, expected: ""},
		{name: "string passthrough", value: `[{"type":"p"}]`, expected: `[{"type":"p"}]`},
		{name: "empty string", value: "", expected: ""},
		{name: "empty any slice", value: []any{}, expected: ""},
		{name: "empty string slice", value: []string{}, expected: ""},
		{name: "any slice", value: []any{map[string]any{"type": "p"}}, expected: `[{"type":"p"}]`},
		{name: "string slice", value: []string{"a", "b"}, expected: `["a","b"]`},
		{name: "map", value: map[string]any{"k": "v"}, expected: `{"k":"v"}`},
		{name: "number", value: 42, expected: "42"},
		{name: "unmarshalable", value: func() {}, expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Stringify(tt.value))
		})
	}
}
