package providerkit

import (
	"encoding/json"
	"testing"

	"github.com/theopenlane/core/v2/pkg/jsonx"
)

func TestSchemaID_RoundTrip(t *testing.T) {
	t.Parallel()

	type cred struct {
		Token string `json:"token" jsonschema:"required"`
	}

	schema := jsonx.SchemaFrom[cred]()
	id := jsonx.SchemaID(schema)

	if id == "" || id == "." {
		t.Fatalf("expected non-empty schema ID from reflected type, got %q", id)
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
