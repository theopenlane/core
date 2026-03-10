package operations

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	schema := SchemaFrom[schemaSample]()

	if err := ValidateConfig(nil, nil); err != nil {
		t.Fatalf("ValidateConfig() expected nil schema to skip validation, got %v", err)
	}

	if err := ValidateConfig(schema, json.RawMessage(`{"name":"ok"}`)); err != nil {
		t.Fatalf("ValidateConfig() expected valid config, got %v", err)
	}

	err := ValidateConfig(schema, json.RawMessage(`{"name":1}`))
	if err == nil {
		t.Fatalf("ValidateConfig() expected invalid config error")
	}
	if !errors.Is(err, ErrOperationConfigInvalid) {
		t.Fatalf("ValidateConfig() expected ErrOperationConfigInvalid, got %v", err)
	}
}

func TestValidateConfigUsesEmptyObjectForNilConfig(t *testing.T) {
	t.Parallel()

	schema := json.RawMessage(`{
		"type":"object",
		"required":["required_key"],
		"properties":{"required_key":{"type":"string"}}
	}`)

	err := ValidateConfig(schema, nil)
	if err == nil {
		t.Fatalf("ValidateConfig() expected missing required field error for nil config")
	}
	if !errors.Is(err, ErrOperationConfigInvalid) {
		t.Fatalf("ValidateConfig() expected ErrOperationConfigInvalid, got %v", err)
	}
}
