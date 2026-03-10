package operations

import (
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

type schemaSample struct {
	Name string `json:"name"`
}

func TestSchemaFrom(t *testing.T) {
	schema := SchemaFrom[schemaSample]()
	if schema == nil {
		t.Fatalf("expected raw schema")
	}

	result, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema),
		gojsonschema.NewBytesLoader([]byte(`{"name":"value"}`)),
	)
	if err != nil {
		t.Fatalf("expected valid schema, got error: %v", err)
	}
	if !result.Valid() {
		t.Fatalf("expected payload to validate: %v", result.Errors())
	}

	invalidResult, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema),
		gojsonschema.NewBytesLoader([]byte(`{"name":123}`)),
	)
	if err != nil {
		t.Fatalf("expected invalid payload validation to run, got error: %v", err)
	}
	if invalidResult.Valid() {
		t.Fatalf("expected payload type mismatch to fail validation")
	}

	unknownFieldResult, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema),
		gojsonschema.NewBytesLoader([]byte(`{"name":"value","unknown":true}`)),
	)
	if err != nil {
		t.Fatalf("expected unknown-field validation to run, got error: %v", err)
	}
	if unknownFieldResult.Valid() {
		t.Fatalf("expected additional properties to be rejected")
	}
}
