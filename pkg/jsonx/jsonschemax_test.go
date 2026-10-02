package jsonx

import (
	"encoding/json"
	"errors"
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

// mergeSettings is the base type the merge tests reflect
type mergeSettings struct {
	Disable bool `json:"disable,omitempty"`
}

// mergeNested is a nested type carried through the merge via its definition
type mergeNested struct {
	Mode string `json:"mode"`
}

// mergeConfig is the extra type the merge tests reflect
type mergeConfig struct {
	Limit  int         `json:"limit" jsonschema:"required"`
	Nested mergeNested `json:"nested,omitempty"`
}

// mergeConflict declares a property the base already declares
type mergeConflict struct {
	Disable bool `json:"disable,omitempty"`
}

// mergeEmpty has no properties
type mergeEmpty struct{}

func TestMergeSchemas(t *testing.T) {
	t.Parallel()

	t.Run("appends properties, required names, and definitions after the base", func(t *testing.T) {
		t.Parallel()

		merged, err := MergeSchemas(SchemaFrom[mergeSettings](), SchemaFrom[mergeConfig]())
		if err != nil {
			t.Fatalf("MergeSchemas() error = %v", err)
		}

		root, defs, err := SchemaRoot(merged)
		if err != nil {
			t.Fatalf("SchemaRoot() error = %v", err)
		}

		var keys []string
		for pair := root.Properties.Oldest(); pair != nil; pair = pair.Next() {
			keys = append(keys, pair.Key)
		}

		if len(keys) != 3 || keys[0] != "disable" || keys[1] != "limit" || keys[2] != "nested" {
			t.Fatalf("properties = %v, want base properties first", keys)
		}

		if len(root.Required) != 1 || root.Required[0] != "limit" {
			t.Fatalf("required = %v, want the extra's required names", root.Required)
		}

		if _, ok := defs["mergeNested"]; !ok {
			t.Fatalf("expected the nested definition carried, got %v", defs)
		}

		if SchemaID(merged) != "mergeSettings" {
			t.Fatalf("expected the base root retained, got %s", SchemaID(merged))
		}

		result, err := ValidateSchema(merged, json.RawMessage(`{"disable":true,"limit":2,"nested":{"mode":"fast"}}`))
		if err != nil || !result.Valid() {
			t.Fatalf("expected the merged schema to validate a combined document, got %v %v", err, ValidationErrorStrings(result))
		}

		result, err = ValidateSchema(merged, json.RawMessage(`{"disable":true,"limit":2,"stray":1}`))
		if err != nil || result.Valid() {
			t.Fatal("expected the merged schema to keep additionalProperties false")
		}
	})

	t.Run("rejects a property declared by both schemas", func(t *testing.T) {
		t.Parallel()

		if _, err := MergeSchemas(SchemaFrom[mergeSettings](), SchemaFrom[mergeConflict]()); !errors.Is(err, ErrSchemaPropertyConflict) {
			t.Fatalf("MergeSchemas() error = %v, want %v", err, ErrSchemaPropertyConflict)
		}
	})

	t.Run("extra without properties leaves the base properties", func(t *testing.T) {
		t.Parallel()

		merged, err := MergeSchemas(SchemaFrom[mergeSettings](), SchemaFrom[mergeEmpty]())
		if err != nil {
			t.Fatalf("MergeSchemas() error = %v", err)
		}

		root, _, err := SchemaRoot(merged)
		if err != nil {
			t.Fatalf("SchemaRoot() error = %v", err)
		}

		if root.Properties.Len() != 1 {
			t.Fatalf("expected only the base property, got %d", root.Properties.Len())
		}
	})

	t.Run("empty extra leaves the base unchanged", func(t *testing.T) {
		t.Parallel()

		merged, err := MergeSchemas(SchemaFrom[mergeSettings](), nil)
		if err != nil {
			t.Fatalf("MergeSchemas() error = %v", err)
		}

		root, _, err := SchemaRoot(merged)
		if err != nil {
			t.Fatalf("SchemaRoot() error = %v", err)
		}

		if root.Properties.Len() != 1 {
			t.Fatalf("expected only the base property, got %d", root.Properties.Len())
		}
	})
}
