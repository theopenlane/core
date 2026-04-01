package operations

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	schema := json.RawMessage(`{"type":"object","required":["name"],"properties":{"name":{"type":"string"}}}`)
	schemaNoRequired := json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}}}`)
	schemaNoAdditional := json.RawMessage(`{"type":"object","required":["name"],"properties":{"name":{"type":"string"}},"additionalProperties":false}`)

	tests := []struct {
		name    string
		schema  json.RawMessage
		value   json.RawMessage
		wantErr error
	}{
		{
			name:    "nil schema passes",
			schema:  nil,
			value:   json.RawMessage(`{"name":"alice"}`),
			wantErr: nil,
		},
		{
			name:    "valid config against schema",
			schema:  schema,
			value:   json.RawMessage(`{"name":"alice"}`),
			wantErr: nil,
		},
		{
			name:    "missing required field fails",
			schema:  schema,
			value:   json.RawMessage(`{"other":"value"}`),
			wantErr: ErrOperationConfigInvalid,
		},
		{
			name:    "invalid JSON value fails",
			schema:  schema,
			value:   json.RawMessage(`{not json`),
			wantErr: ErrOperationConfigInvalid,
		},
		{
			name:    "empty config against schema with required fields",
			schema:  schema,
			value:   json.RawMessage(`{}`),
			wantErr: ErrOperationConfigInvalid,
		},
		{
			name:    "nil config against schema with no required fields",
			schema:  schemaNoRequired,
			value:   nil,
			wantErr: nil,
		},
		{
			name:    "extra fields with additionalProperties allowed",
			schema:  schema,
			value:   json.RawMessage(`{"name":"alice","extra":"field"}`),
			wantErr: nil,
		},
		{
			name:    "extra fields with additionalProperties false",
			schema:  schemaNoAdditional,
			value:   json.RawMessage(`{"name":"alice","extra":"field"}`),
			wantErr: ErrOperationConfigInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateConfig(tc.schema, tc.value)

			switch {
			case tc.wantErr == nil && err != nil:
				t.Fatalf("expected no error, got %v", err)
			case tc.wantErr != nil && err == nil:
				t.Fatalf("expected error %v, got nil", tc.wantErr)
			case tc.wantErr != nil && !errors.Is(err, tc.wantErr):
				t.Fatalf("expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}
