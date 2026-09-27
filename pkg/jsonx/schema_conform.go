package jsonx

import (
	"encoding/json"
	"reflect"

	"github.com/invopop/jsonschema"
)

// ConformToSchema strips undeclared top-level properties and fills defaulted required properties, following the root $ref
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

	out := EditObject(doc, func(object map[string]json.RawMessage) bool {
		changed := false

		if reflect.DeepEqual(root.AdditionalProperties, jsonschema.FalseSchema) {
			for key := range object {
				if _, declared := root.Properties.Get(key); !declared {
					delete(object, key)

					changed = true
				}
			}
		}

		for _, name := range root.Required {
			property, declared := root.Properties.Get(name)
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
