package jsonx

import (
	"encoding/json"
	"testing"
)

func TestValidateSchema(t *testing.T) {
	t.Parallel()

	schema := json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["name"],
		"properties":{"name":{"type":"string"}}
	}`)

	validDoc := map[string]any{"name": "ok"}
	result, err := ValidateSchema(schema, validDoc)
	if err != nil {
		t.Fatalf("ValidateSchema() unexpected error: %v", err)
	}
	if !result.Valid() {
		t.Fatalf("expected valid result, got %v", ValidationErrorStrings(result))
	}

	invalidDoc := map[string]any{"name": 1}
	result, err = ValidateSchema(schema, invalidDoc)
	if err != nil {
		t.Fatalf("ValidateSchema() unexpected invalid-doc error: %v", err)
	}
	if result.Valid() {
		t.Fatalf("expected invalid result")
	}

	issues := ValidationErrorStrings(result)
	if len(issues) == 0 {
		t.Fatalf("expected validation errors")
	}
}

func TestValidateSchemaUsesBytesLoader(t *testing.T) {
	t.Parallel()

	schema := []byte(`{"type":"object","required":["enabled"],"properties":{"enabled":{"type":"boolean"}}}`)
	doc := []byte(`{"enabled":true}`)

	result, err := ValidateSchema(schema, doc)
	if err != nil {
		t.Fatalf("ValidateSchema() unexpected error: %v", err)
	}
	if !result.Valid() {
		t.Fatalf("expected valid bytes result, got %v", ValidationErrorStrings(result))
	}
}
