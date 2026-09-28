package jsonx

import (
	"encoding/json"
	"errors"
	"maps"
	"testing"
)

func TestPropertyRefs(t *testing.T) {
	t.Parallel()

	t.Run("maps ref properties to their definition and others to empty", func(t *testing.T) {
		t.Parallel()

		got, err := PropertyRefs(SchemaFrom[conformOuter]())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := map[string]string{"name": "", "nested": "conformInner"}
		if !maps.Equal(got, want) {
			t.Fatalf("PropertyRefs() = %v, want %v", got, want)
		}
	})

	t.Run("schema without properties yields an empty map", func(t *testing.T) {
		t.Parallel()

		got, err := PropertyRefs(json.RawMessage(`{"type":"string"}`))
		if err != nil || len(got) != 0 {
			t.Fatalf("PropertyRefs() = %v, %v", got, err)
		}
	})

	t.Run("unresolved root ref", func(t *testing.T) {
		t.Parallel()

		_, err := PropertyRefs(json.RawMessage(`{"$ref":"#/$defs/missing"}`))
		if !errors.Is(err, ErrSchemaRefUnresolved) {
			t.Fatalf("expected %v, got %v", ErrSchemaRefUnresolved, err)
		}
	})
}

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
