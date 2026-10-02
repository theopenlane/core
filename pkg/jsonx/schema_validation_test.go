package jsonx

import (
	"encoding/json"
	"errors"
	"testing"
)

// validateTestInput is the type the ValidateAs tests reflect
type validateTestInput struct {
	Name string `json:"name" jsonschema:"required"`
}

func TestValidate(t *testing.T) {
	t.Parallel()

	schema := json.RawMessage(`{"type":"object","required":["name"],"properties":{"name":{"type":"string"}},"additionalProperties":false}`)
	optional := json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}}}`)

	tests := []struct {
		name       string
		schema     json.RawMessage
		payload    json.RawMessage
		wantErr    bool
		wantSchema bool
	}{
		{name: "empty schema is unconstrained", schema: nil, payload: json.RawMessage(`"anything"`)},
		{name: "valid payload passes", schema: schema, payload: json.RawMessage(`{"name":"alice"}`)},
		{name: "absent payload validates as an empty object", schema: optional, payload: nil},
		{name: "absent payload fails required keys", schema: schema, payload: nil, wantErr: true, wantSchema: true},
		{name: "missing required key fails", schema: schema, payload: json.RawMessage(`{"other":1}`), wantErr: true, wantSchema: true},
		{name: "undeclared key fails when additional properties are closed", schema: schema, payload: json.RawMessage(`{"name":"alice","extra":1}`), wantErr: true, wantSchema: true},
		{name: "type mismatch fails", schema: schema, payload: json.RawMessage(`"alice"`), wantErr: true, wantSchema: true},
		{name: "malformed payload fails as a schema error", schema: schema, payload: json.RawMessage(`{not json`), wantErr: true, wantSchema: true},
		{name: "malformed schema fails without a schema error", schema: json.RawMessage(`{not json`), payload: json.RawMessage(`{"name":"alice"}`), wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := Validate(tc.schema, tc.payload)

			switch {
			case !tc.wantErr && err != nil:
				t.Fatalf("Validate() error = %v, want nil", err)
			case tc.wantErr && err == nil:
				t.Fatal("Validate() error = nil, want error")
			case tc.wantErr && tc.wantSchema != errors.Is(err, ErrSchemaInvalid):
				t.Fatalf("Validate() error = %v, want schema error %v", err, tc.wantSchema)
			}
		})
	}
}

func TestSchemaErrorJoinsIssues(t *testing.T) {
	t.Parallel()

	schema := json.RawMessage(`{"type":"object","required":["a","b"],"properties":{"a":{"type":"string"},"b":{"type":"string"}}}`)

	err := Validate(schema, json.RawMessage(`{}`))

	var schemaErr *SchemaError
	if !errors.As(err, &schemaErr) {
		t.Fatalf("Validate() error = %T, want *SchemaError", err)
	}

	if len(schemaErr.Issues) != 2 {
		t.Fatalf("Issues = %v, want two missing-key issues", schemaErr.Issues)
	}

	if err.Error() != schemaErr.Issues[0]+"; "+schemaErr.Issues[1] {
		t.Fatalf("Error() = %q, want the issues joined", err.Error())
	}
}

func TestValidateAs(t *testing.T) {
	t.Parallel()

	got, err := ValidateAs[validateTestInput](json.RawMessage(`{"name":"alice"}`))
	if err != nil {
		t.Fatalf("ValidateAs() error = %v", err)
	}

	if got.Name != "alice" {
		t.Fatalf("ValidateAs() = %+v, want name alice", got)
	}

	if _, err := ValidateAs[validateTestInput](json.RawMessage(`{}`)); !errors.Is(err, ErrSchemaInvalid) {
		t.Fatalf("ValidateAs() error = %v, want %v", err, ErrSchemaInvalid)
	}
}
