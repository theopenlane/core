package providerkit

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
				t.Fatalf("SchemaID() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSchemaID_RoundTrip(t *testing.T) {
	t.Parallel()

	type cred struct {
		Token string `json:"token" jsonschema:"required"`
	}

	schema := SchemaFrom[cred]()
	id := SchemaID(schema)

	if id == "" || id == "." {
		t.Fatalf("expected non-empty schema ID from reflected type, got %q", id)
	}
}

func TestCredentialSchema(t *testing.T) {
	t.Parallel()

	type testCred struct {
		APIKey string `json:"api_key" jsonschema:"required"`
	}

	schema, ref := CredentialSchema[testCred]()

	if schema == nil {
		t.Fatal("expected non-nil schema")
	}

	var doc map[string]any
	if err := json.Unmarshal(schema, &doc); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}

	if ref.String() == "" {
		t.Fatal("expected non-empty credential ref identity")
	}
}

func TestOperationSchema(t *testing.T) {
	t.Parallel()

	type testOpConfig struct {
		Limit int `json:"limit"`
	}

	schema, ref := OperationSchema[testOpConfig]()

	if schema == nil {
		t.Fatal("expected non-nil schema")
	}

	var doc map[string]any
	if err := json.Unmarshal(schema, &doc); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}

	if ref.Name() == "" {
		t.Fatal("expected non-empty operation ref name")
	}
}

func TestWebhookEventSchema(t *testing.T) {
	t.Parallel()

	type testEvent struct {
		Action string `json:"action"`
	}

	schema, ref := WebhookEventSchema[testEvent]()

	if schema == nil {
		t.Fatal("expected non-nil schema")
	}

	var doc map[string]any
	if err := json.Unmarshal(schema, &doc); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}

	if ref.Name() == "" {
		t.Fatal("expected non-empty webhook event ref name")
	}
}
