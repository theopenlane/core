package jsonx

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/invopop/jsonschema"
)

// ConformToSchema returns doc with undeclared properties stripped and required defaults filled
func ConformToSchema(schema, doc json.RawMessage) (json.RawMessage, error) {
	root, defs, err := SchemaRoot(schema)
	if err != nil {
		return nil, err
	}

	if root.Properties == nil {
		return doc, nil
	}

	if IsEmptyRawMessage(doc) {
		doc = json.RawMessage("{}")
	}

	return conformObject(root, defs, doc)
}

// conformObject conforms one object document recursing into present properties
func conformObject(node *jsonschema.Schema, defs jsonschema.Definitions, doc json.RawMessage) (json.RawMessage, error) {
	var err error

	out := EditObject(doc, func(object map[string]json.RawMessage) bool {
		stripped := stripUndeclared(node, object)

		var nested, filled bool

		if nested, err = conformNested(node, defs, object); err != nil {
			return false
		}

		if filled, err = fillRequiredDefaults(node, defs, object); err != nil {
			return false
		}

		return stripped || nested || filled
	})

	if err != nil {
		return nil, err
	}

	return out, nil
}

// stripUndeclared removes keys the schema does not declare when additional properties are forbidden
func stripUndeclared(node *jsonschema.Schema, object map[string]json.RawMessage) bool {
	if !reflect.DeepEqual(node.AdditionalProperties, jsonschema.FalseSchema) {
		return false
	}

	changed := false

	for key := range object {
		if _, declared := node.Properties.Get(key); !declared {
			delete(object, key)

			changed = true
		}
	}

	return changed
}

// conformNested recurses into each present object-valued property
func conformNested(node *jsonschema.Schema, defs jsonschema.Definitions, object map[string]json.RawMessage) (bool, error) {
	changed := false

	for pair := node.Properties.Oldest(); pair != nil; pair = pair.Next() {
		value, present := object[pair.Key]
		if !present {
			continue
		}

		property, err := followSchemaRef(pair.Value, defs)
		if err != nil {
			return false, err
		}

		if property.Properties == nil {
			continue
		}

		nested, err := conformObject(property, defs, value)
		if err != nil {
			return false, err
		}

		if !bytes.Equal(nested, value) {
			object[pair.Key] = nested
			changed = true
		}
	}

	return changed, nil
}

// fillRequiredDefaults sets each missing required property that declares a default
func fillRequiredDefaults(node *jsonschema.Schema, defs jsonschema.Definitions, object map[string]json.RawMessage) (bool, error) {
	changed := false

	for _, name := range node.Required {
		property, declared := node.Properties.Get(name)
		if _, present := object[name]; present || !declared {
			continue
		}

		property, err := followSchemaRef(property, defs)
		if err != nil {
			return false, err
		}

		if property.Default == nil {
			continue
		}

		value, err := json.Marshal(property.Default)
		if err != nil {
			return false, err
		}

		object[name] = value
		changed = true
	}

	return changed, nil
}
