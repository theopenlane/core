package jsonx

import (
	"encoding/json"
	"testing"
)

func TestSchemaID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		schema json.RawMessage
		want   string
	}{
		{
			name:   "extracts definition key from ref path",
			schema: json.RawMessage(`{"$ref":"#/$defs/MyType"}`),
			want:   "MyType",
		},
		{
			name:   "nested ref path returns base",
			schema: json.RawMessage(`{"$ref":"#/$defs/deep/Nested"}`),
			want:   "Nested",
		},
		{
			name:   "invalid JSON returns empty string",
			schema: json.RawMessage(`not-json`),
			want:   "",
		},
		{
			name:   "missing ref returns dot",
			schema: json.RawMessage(`{}`),
			want:   ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := SchemaID(tc.schema)
			if got != tc.want {
				t.Fatalf("jsonx.SchemaID() = %q, want %q", got, tc.want)
			}
		})
	}
}
