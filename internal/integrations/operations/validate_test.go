package operations

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func TestValidateInput(t *testing.T) {
	t.Parallel()

	schema := json.RawMessage(`{"type":"object","required":["name"],"properties":{"name":{"type":"string"}}}`)
	schemaNoRequired := json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}}}`)
	schemaNoAdditional := json.RawMessage(`{"type":"object","required":["name"],"properties":{"name":{"type":"string"}},"additionalProperties":false}`)
	errSemantic := errors.New("name is reserved")
	rejectAdmin := func(_ context.Context, _ types.InstallationRequest, payload json.RawMessage) error {
		if string(payload) == `{"name":"admin"}` {
			return errSemantic
		}

		return nil
	}

	tests := []struct {
		name     string
		schema   json.RawMessage
		validate types.ValidateFunc
		payload  json.RawMessage
		sentinel error
		wantErr  error
		wantRaw  bool
	}{
		{name: "nil schema passes", payload: json.RawMessage(`{"name":"alice"}`), sentinel: types.ErrOperationConfigInvalid},
		{name: "valid payload against schema", schema: schema, payload: json.RawMessage(`{"name":"alice"}`), sentinel: types.ErrOperationConfigInvalid},
		{name: "missing required field fails with the sentinel", schema: schema, payload: json.RawMessage(`{"other":"value"}`), sentinel: types.ErrOperationConfigInvalid, wantErr: types.ErrOperationConfigInvalid},
		{name: "invalid JSON payload fails with the sentinel", schema: schema, payload: json.RawMessage(`{not json`), sentinel: types.ErrOperationConfigInvalid, wantErr: types.ErrOperationConfigInvalid},
		{name: "empty payload against required fields fails", schema: schema, payload: json.RawMessage(`{}`), sentinel: types.ErrOperationConfigInvalid, wantErr: types.ErrOperationConfigInvalid},
		{name: "nil payload against optional fields passes", schema: schemaNoRequired, payload: nil, sentinel: types.ErrOperationConfigInvalid},
		{name: "extra fields pass when additional properties are open", schema: schema, payload: json.RawMessage(`{"name":"alice","extra":"field"}`), sentinel: types.ErrOperationConfigInvalid},
		{name: "extra fields fail when additional properties are closed", schema: schemaNoAdditional, payload: json.RawMessage(`{"name":"alice","extra":"field"}`), sentinel: types.ErrOperationConfigInvalid, wantErr: types.ErrOperationConfigInvalid},
		{name: "type mismatch fails with the given sentinel", schema: json.RawMessage(`{"type":"object"}`), payload: json.RawMessage(`"a string"`), sentinel: ErrDispatchInputInvalid, wantErr: ErrDispatchInputInvalid},
		{name: "malformed schema surfaces the raw error", schema: json.RawMessage(`{not valid json`), payload: json.RawMessage(`{"name":"alice"}`), sentinel: types.ErrOperationConfigInvalid, wantRaw: true},
		{name: "semantic validation runs after the schema passes", schema: schema, validate: rejectAdmin, payload: json.RawMessage(`{"name":"admin"}`), sentinel: types.ErrOperationConfigInvalid, wantErr: errSemantic},
		{name: "semantic validation wraps the sentinel", schema: schema, validate: rejectAdmin, payload: json.RawMessage(`{"name":"admin"}`), sentinel: types.ErrOperationConfigInvalid, wantErr: types.ErrOperationConfigInvalid},
		{name: "semantic validation is skipped when the schema fails", schema: schema, validate: rejectAdmin, payload: json.RawMessage(`{}`), sentinel: types.ErrOperationConfigInvalid, wantErr: types.ErrOperationConfigInvalid},
		{name: "semantic validation passes", schema: schema, validate: rejectAdmin, payload: json.RawMessage(`{"name":"alice"}`), sentinel: types.ErrOperationConfigInvalid},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateInput(context.Background(), types.InstallationRequest{}, tc.schema, tc.validate, tc.payload, tc.sentinel)

			switch {
			case tc.wantRaw && err == nil:
				t.Fatal("expected a raw error, got nil")
			case tc.wantRaw && errors.Is(err, tc.sentinel):
				t.Fatalf("expected a raw error, got the sentinel %v", err)
			case tc.wantRaw:
			case tc.wantErr == nil && err != nil:
				t.Fatalf("expected no error, got %v", err)
			case tc.wantErr != nil && !errors.Is(err, tc.wantErr):
				t.Fatalf("expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}
