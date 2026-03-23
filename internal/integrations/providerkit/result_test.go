package providerkit

import (
	"encoding/json"
	"errors"
	"testing"
)

var errEncodeTest = errors.New("encode failed")

func TestEncodeResult(t *testing.T) {
	t.Parallel()

	t.Run("success serialization", func(t *testing.T) {
		t.Parallel()

		got, err := EncodeResult(map[string]string{"key": "value"}, errEncodeTest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var m map[string]string
		if err := json.Unmarshal(got, &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if m["key"] != "value" {
			t.Fatalf("expected key=value, got %q", m["key"])
		}
	})

	t.Run("unencodable value returns custom error", func(t *testing.T) {
		t.Parallel()

		// Channels cannot be marshaled to JSON
		_, err := EncodeResult(make(chan int), errEncodeTest)
		if !errors.Is(err, errEncodeTest) {
			t.Fatalf("expected errEncodeTest, got %v", err)
		}
	})
}

func TestMarshalEnvelope(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		env, err := MarshalEnvelope("repos", map[string]int{"count": 3}, errEncodeTest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if env.Resource != "repos" {
			t.Fatalf("expected resource=repos, got %q", env.Resource)
		}

		if env.Variant != "" {
			t.Fatalf("expected empty variant, got %q", env.Variant)
		}

		var m map[string]float64
		if err := json.Unmarshal(env.Payload, &m); err != nil {
			t.Fatalf("unmarshal payload: %v", err)
		}

		if m["count"] != 3 {
			t.Fatalf("expected count=3, got %v", m["count"])
		}
	})
}

func TestMarshalEnvelopeVariant(t *testing.T) {
	t.Parallel()

	t.Run("success with variant", func(t *testing.T) {
		t.Parallel()

		env, err := MarshalEnvelopeVariant("alert", "vulns", map[string]bool{"critical": true}, errEncodeTest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if env.Variant != "alert" {
			t.Fatalf("expected variant=alert, got %q", env.Variant)
		}

		if env.Resource != "vulns" {
			t.Fatalf("expected resource=vulns, got %q", env.Resource)
		}

		var m map[string]bool
		if err := json.Unmarshal(env.Payload, &m); err != nil {
			t.Fatalf("unmarshal payload: %v", err)
		}

		if !m["critical"] {
			t.Fatal("expected critical=true")
		}
	})
}

func TestRawEnvelope(t *testing.T) {
	t.Parallel()

	t.Run("wraps raw payload", func(t *testing.T) {
		t.Parallel()

		raw := json.RawMessage(`{"id":1}`)
		env := RawEnvelope("items", raw)

		if env.Resource != "items" {
			t.Fatalf("expected resource=items, got %q", env.Resource)
		}

		if env.Variant != "" {
			t.Fatalf("expected empty variant, got %q", env.Variant)
		}

		if string(env.Payload) != `{"id":1}` {
			t.Fatalf("expected payload {\"id\":1}, got %s", env.Payload)
		}
	})
}

func TestRawEnvelopeVariant(t *testing.T) {
	t.Parallel()

	t.Run("wraps with variant", func(t *testing.T) {
		t.Parallel()

		raw := json.RawMessage(`{"id":2}`)
		env := RawEnvelopeVariant("warning", "alerts", raw)

		if env.Variant != "warning" {
			t.Fatalf("expected variant=warning, got %q", env.Variant)
		}

		if env.Resource != "alerts" {
			t.Fatalf("expected resource=alerts, got %q", env.Resource)
		}

		if string(env.Payload) != `{"id":2}` {
			t.Fatalf("expected payload {\"id\":2}, got %s", env.Payload)
		}
	})
}

func TestSchemaFrom(t *testing.T) {
	t.Parallel()

	t.Run("produces valid JSON from struct", func(t *testing.T) {
		t.Parallel()

		type sample struct {
			Name  string `json:"name" jsonschema:"required"`
			Count int    `json:"count"`
		}

		raw := SchemaFrom[sample]()
		if raw == nil {
			t.Fatal("expected non-nil schema")
		}

		var schema map[string]any
		if err := json.Unmarshal(raw, &schema); err != nil {
			t.Fatalf("schema is not valid JSON: %v", err)
		}

		// The reflector emits a $ref + $defs structure; verify the definition exists
		defs, ok := schema["$defs"]
		if !ok {
			t.Fatal("expected $defs key in schema")
		}

		defsMap, ok := defs.(map[string]any)
		if !ok {
			t.Fatalf("expected $defs to be map, got %T", defs)
		}

		sampleDef, ok := defsMap["sample"]
		if !ok {
			t.Fatal("expected sample definition in $defs")
		}

		defMap, ok := sampleDef.(map[string]any)
		if !ok {
			t.Fatalf("expected sample def to be map, got %T", sampleDef)
		}

		props, ok := defMap["properties"]
		if !ok {
			t.Fatal("expected properties in sample definition")
		}

		propsMap, ok := props.(map[string]any)
		if !ok {
			t.Fatalf("expected properties to be map, got %T", props)
		}

		if _, ok := propsMap["name"]; !ok {
			t.Fatal("expected name property in schema")
		}

		if _, ok := propsMap["count"]; !ok {
			t.Fatal("expected count property in schema")
		}
	})
}
