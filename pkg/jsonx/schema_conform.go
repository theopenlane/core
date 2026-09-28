package jsonx

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/invopop/jsonschema"
)

// ConformToSchema strips undeclared properties and fills defaulted required properties, following the root $ref and recursing into nested object properties present in the document
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
		changed := false

		if reflect.DeepEqual(node.AdditionalProperties, jsonschema.FalseSchema) {
			for key := range object {
				if _, declared := node.Properties.Get(key); !declared {
					delete(object, key)

					changed = true
				}
			}
		}

		for pair := node.Properties.Oldest(); pair != nil; pair = pair.Next() {
			value, present := object[pair.Key]
			if !present {
				continue
			}

			var property *jsonschema.Schema
			if property, err = followSchemaRef(pair.Value, defs); err != nil {
				return false
			}

			if property.Properties == nil {
				continue
			}

			var nested json.RawMessage
			if nested, err = conformObject(property, defs, value); err != nil {
				return false
			}

			if !bytes.Equal(nested, value) {
				object[pair.Key] = nested
				changed = true
			}
		}

		for _, name := range node.Required {
			property, declared := node.Properties.Get(name)
			if _, present := object[name]; present || !declared {
				continue
			}

			if property, err = followSchemaRef(property, defs); err != nil {
				return false
			}

			if property.Default == nil {
				continue
			}

			var value json.RawMessage
			if value, err = json.Marshal(property.Default); err != nil {
				return false
			}

			object[name] = value
			changed = true
		}

		return changed
	})

	if err != nil {
		return nil, err
	}

	return out, nil
}
