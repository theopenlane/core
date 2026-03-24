package runtime

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestValidatePayloadEmptySchemaSkips(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("should not surface")
	if err := validatePayload(nil, json.RawMessage(`{"x":1}`), sentinel); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidatePayloadValidData(t *testing.T) {
	t.Parallel()

	schema := json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}`)
	data := json.RawMessage(`{"name":"test"}`)

	if err := validatePayload(schema, data, ErrUserInputInvalid); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidatePayloadInvalidDataReturnsSentinel(t *testing.T) {
	t.Parallel()

	schema := json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}`)
	data := json.RawMessage(`{}`)

	err := validatePayload(schema, data, ErrUserInputInvalid)
	if !errors.Is(err, ErrUserInputInvalid) {
		t.Fatalf("expected ErrUserInputInvalid, got %v", err)
	}
}

func TestValidatePayloadMalformedSchemaReturnsError(t *testing.T) {
	t.Parallel()

	schema := json.RawMessage(`{not valid json`)
	data := json.RawMessage(`{"name":"test"}`)

	if err := validatePayload(schema, data, ErrUserInputInvalid); err == nil {
		t.Fatal("expected error for malformed schema")
	}
}

func TestValidatePayloadTypeMismatchReturnsSentinel(t *testing.T) {
	t.Parallel()

	schema := json.RawMessage(`{"type":"object"}`)
	data := json.RawMessage(`"a string"`)

	err := validatePayload(schema, data, ErrCredentialInvalid)
	if !errors.Is(err, ErrCredentialInvalid) {
		t.Fatalf("expected ErrCredentialInvalid, got %v", err)
	}
}
