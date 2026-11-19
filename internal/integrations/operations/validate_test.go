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

func TestConformStored(t *testing.T) {
	t.Parallel()

	const (
		retiredName  = "retiredName"
		upgradedBody = `{"name":"upgraded"}`
	)

	schema := json.RawMessage(`{"type":"object","required":["name"],"properties":{"name":{"type":"string"}},"additionalProperties":false}`)
	defaulted := json.RawMessage(`{"type":"object","required":["name"],"properties":{"name":{"type":"string","default":"anon"}},"additionalProperties":false}`)
	malformed := json.RawMessage(`{not valid json`)
	errUpgrade := errors.New("upgrade failed")
	errSemantic := errors.New("name is reserved")

	upgradeTo := func(wantFrom string) types.UpgradeFunc {
		return func(_ context.Context, _ types.InstallationRequest, from string, _ json.RawMessage) (json.RawMessage, error) {
			if from != wantFrom {
				return nil, errUpgrade
			}

			return json.RawMessage(upgradedBody), nil
		}
	}

	tests := []struct {
		name    string
		layout  types.InputRegistration
		from    string
		payload string
		want    string
		wantErr []error
	}{
		{name: "valid document is unchanged", layout: types.InputRegistration{Schema: schema}, payload: `{"name":"alice"}`, want: `{"name":"alice"}`},
		{name: "undeclared key is stripped", layout: types.InputRegistration{Schema: schema}, payload: `{"name":"alice","extra":1}`, want: `{"name":"alice"}`},
		{name: "required default is filled", layout: types.InputRegistration{Schema: defaulted}, payload: `{}`, want: `{"name":"anon"}`},
		{name: "upgrade receives the stored name and replaces the document", layout: types.InputRegistration{Schema: schema, Upgrade: upgradeTo(retiredName)}, from: retiredName, payload: `{"name":"old"}`, want: upgradedBody},
		{name: "upgrade error wraps the sentinel and the cause", layout: types.InputRegistration{Schema: schema, Upgrade: upgradeTo(retiredName)}, from: "other", payload: `{"name":"old"}`, wantErr: []error{types.ErrOperationConfigInvalid, errUpgrade}},
		{name: "schema conform error wraps the sentinel", layout: types.InputRegistration{Schema: malformed}, payload: `{"name":"alice"}`, wantErr: []error{types.ErrOperationConfigInvalid}},
		{name: "schema violation after conform wraps the sentinel", layout: types.InputRegistration{Schema: schema}, payload: `{}`, wantErr: []error{types.ErrOperationConfigInvalid}},
		{
			name: "semantic validation error wraps the sentinel and the cause",
			layout: types.InputRegistration{Schema: schema, Validate: func(context.Context, types.InstallationRequest, json.RawMessage) error {
				return errSemantic
			}},
			payload: `{"name":"admin"}`,
			wantErr: []error{types.ErrOperationConfigInvalid, errSemantic},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ConformStored(context.Background(), types.InstallationRequest{}, tc.layout, tc.from, json.RawMessage(tc.payload), types.ErrOperationConfigInvalid)

			if len(tc.wantErr) > 0 {
				for _, want := range tc.wantErr {
					if !errors.Is(err, want) {
						t.Fatalf("expected error %v, got %v", want, err)
					}
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if string(got) != tc.want {
				t.Fatalf("ConformStored() = %s, want %s", got, tc.want)
			}
		})
	}
}
