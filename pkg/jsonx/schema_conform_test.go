package jsonx

import (
	"encoding/json"
	"errors"
	"testing"
)

type conformInner struct {
	Mode string `json:"mode" jsonschema:"required,default=fast"`
}

type conformOuter struct {
	Name   string       `json:"name" jsonschema:"required"`
	Nested conformInner `json:"nested"`
}

func TestConformToSchema(t *testing.T) {
	t.Parallel()

	strict := `{"type":"object","additionalProperties":false,"required":["token","region"],"properties":{"token":{"type":"string"},"region":{"type":"string","default":"us"},"extra":{"type":"string"}}}`
	open := `{"type":"object","required":["token"],"properties":{"token":{"type":"string"}}}`

	tests := []struct {
		name   string
		schema string
		doc    string
		want   string
	}{
		{name: "strips undeclared under additionalProperties false", schema: strict, doc: `{"token":"t","region":"eu","legacy":1}`, want: `{"region":"eu","token":"t"}`},
		{name: "keeps undeclared when additional properties are allowed", schema: open, doc: `{"token":"t","legacy":1}`, want: `{"token":"t","legacy":1}`},
		{name: "fills a defaulted required property", schema: strict, doc: `{"token":"t"}`, want: `{"region":"us","token":"t"}`},
		{name: "leaves a required property without default absent", schema: strict, doc: `{"region":"eu"}`, want: `{"region":"eu"}`},
		{name: "leaves optional properties absent", schema: strict, doc: `{"token":"t","region":"eu"}`, want: `{"token":"t","region":"eu"}`},
		{name: "empty document takes defaults", schema: strict, doc: ``, want: `{"region":"us"}`},
		{name: "non-object document passes through", schema: strict, doc: `"scalar"`, want: `"scalar"`},
		{name: "schema without properties passes through", schema: `{"type":"string"}`, doc: `{"a":1}`, want: `{"a":1}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ConformToSchema(json.RawMessage(tc.schema), json.RawMessage(tc.doc))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(got) != tc.want {
				t.Fatalf("got %s, want %s", got, tc.want)
			}
		})
	}
}

func TestConformToSchemaFollowsRefs(t *testing.T) {
	t.Parallel()

	got, err := ConformToSchema(SchemaFrom[conformOuter](), json.RawMessage(`{"name":"n","nested":{"extra":true},"stale":1}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out map[string]json.RawMessage
	if err := json.Unmarshal(got, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if _, ok := out["stale"]; ok {
		t.Fatal("expected the undeclared top-level property to be stripped")
	}

	if string(out["nested"]) != `{"extra":true}` {
		t.Fatalf("expected the nested object to pass through unchanged, got %s", out["nested"])
	}
}

func TestConformToSchemaUnresolvedRef(t *testing.T) {
	t.Parallel()

	_, err := ConformToSchema(json.RawMessage(`{"$ref":"#/$defs/missing"}`), json.RawMessage(`{}`))
	if !errors.Is(err, ErrSchemaRefUnresolved) {
		t.Fatalf("expected %v, got %v", ErrSchemaRefUnresolved, err)
	}
}

func TestSchemaRoot(t *testing.T) {
	t.Parallel()

	t.Run("root without ref", func(t *testing.T) {
		t.Parallel()

		root, defs, err := SchemaRoot(json.RawMessage(`{"type":"object","properties":{"token":{"type":"string"}}}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := root.Properties.Get("token"); !ok || defs != nil {
			t.Fatalf("expected the document itself with no definitions, got %+v defs=%v", root, defs)
		}
	})

	t.Run("root ref resolves into defs", func(t *testing.T) {
		t.Parallel()

		root, defs, err := SchemaRoot(SchemaFrom[conformOuter]())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := root.Properties.Get("name"); !ok {
			t.Fatalf("expected the conformOuter definition, got %+v", root)
		}

		if _, ok := defs["conformInner"]; !ok {
			t.Fatalf("expected definitions to be returned, got %v", defs)
		}
	})

	t.Run("missing def is unresolved", func(t *testing.T) {
		t.Parallel()

		_, _, err := SchemaRoot(json.RawMessage(`{"$ref":"#/$defs/missing"}`))
		if !errors.Is(err, ErrSchemaRefUnresolved) {
			t.Fatalf("expected %v, got %v", ErrSchemaRefUnresolved, err)
		}
	})

	t.Run("empty document yields empty schema", func(t *testing.T) {
		t.Parallel()

		root, defs, err := SchemaRoot(nil)
		if err != nil || root == nil || root.Properties != nil || defs != nil {
			t.Fatalf("expected an empty schema, got %+v defs=%v err=%v", root, defs, err)
		}
	})
}
