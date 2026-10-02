package jsonx

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/invopop/jsonschema"
)

// reflector is the shared JSON schema reflector used by SchemaFrom
var reflector = &jsonschema.Reflector{
	AllowAdditionalProperties:  false,
	RequiredFromJSONSchemaTags: true,
}

// SchemaFrom reflects a JSON schema from a Go type and returns it as raw JSON
func SchemaFrom[T any]() json.RawMessage {
	out, err := ToRawMessage(reflector.Reflect(new(T)))
	if err != nil {
		return nil
	}

	return out
}

// PropertyDescriptor is a top-level JSON schema property with its name and description
type PropertyDescriptor struct {
	// Name is the JSON property key as it appears in the reflected schema
	Name string
	// Description is the human-readable description extracted from the jsonschema description tag
	Description string
}

// PropertyDescriptors returns a Go type's top-level JSON schema properties
func PropertyDescriptors[T any]() []PropertyDescriptor {
	schema, _, err := SchemaRoot(SchemaFrom[T]())
	if err != nil || schema.Properties == nil {
		return nil
	}

	out := make([]PropertyDescriptor, 0, schema.Properties.Len())

	for pair := schema.Properties.Oldest(); pair != nil; pair = pair.Next() {
		out = append(out, PropertyDescriptor{
			Name:        pair.Key,
			Description: pair.Value.Description,
		})
	}

	return out
}

// SchemaRoot decodes a raw schema and follows its root $ref into its definitions
func SchemaRoot(schema json.RawMessage) (*jsonschema.Schema, jsonschema.Definitions, error) {
	var doc jsonschema.Schema
	if err := UnmarshalIfPresent(schema, &doc); err != nil {
		return nil, nil, err
	}

	root, err := followSchemaRef(&doc, doc.Definitions)
	if err != nil {
		return nil, nil, err
	}

	return root, doc.Definitions, nil
}

// followSchemaRef returns the definition a node's $ref names, or the node itself
func followSchemaRef(node *jsonschema.Schema, defs jsonschema.Definitions) (*jsonschema.Schema, error) {
	if node.Ref == "" {
		return node, nil
	}

	target, ok := defs[path.Base(node.Ref)]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrSchemaRefUnresolved, node.Ref)
	}

	return target, nil
}

// MergeSchemas returns base with extra's root properties, required names, and definitions appended
func MergeSchemas(base, extra json.RawMessage) (json.RawMessage, error) {
	var doc jsonschema.Schema
	if err := UnmarshalIfPresent(base, &doc); err != nil {
		return nil, err
	}

	root, err := followSchemaRef(&doc, doc.Definitions)
	if err != nil {
		return nil, err
	}

	extraRoot, extraDefs, err := SchemaRoot(extra)
	if err != nil {
		return nil, err
	}

	if root.Properties == nil {
		root.Properties = jsonschema.NewProperties()
	}

	if extraRoot.Properties != nil {
		for pair := extraRoot.Properties.Oldest(); pair != nil; pair = pair.Next() {
			if _, exists := root.Properties.Get(pair.Key); exists {
				return nil, fmt.Errorf("%w: %s", ErrSchemaPropertyConflict, pair.Key)
			}

			root.Properties.Set(pair.Key, pair.Value)
		}
	}

	root.Required = append(root.Required, extraRoot.Required...)

	if doc.Definitions == nil {
		doc.Definitions = jsonschema.Definitions{}
	}

	rootName := path.Base(doc.Ref)

	for name, definition := range extraDefs {
		if name == rootName {
			continue
		}

		doc.Definitions[name] = definition
	}

	return ToRawMessage(&doc)
}

// SchemaID extracts the definition key from a reflected JSON schema's $ref path
func SchemaID(schema json.RawMessage) string {
	var doc struct {
		Ref string `json:"$ref"`
	}

	if err := json.Unmarshal(schema, &doc); err != nil {
		return ""
	}

	return path.Base(doc.Ref)
}

// InjectDefaults returns the schema with stored values injected as defaults
func InjectDefaults(schema json.RawMessage, defaults map[string]any) (json.RawMessage, error) {
	var doc jsonschema.Schema
	if err := json.Unmarshal(schema, &doc); err != nil {
		return schema, err
	}

	root, err := followSchemaRef(&doc, doc.Definitions)
	if err != nil || root.Properties == nil {
		return schema, nil
	}

	injectSchemaDefaults(root, doc.Definitions, defaults)

	out, err := json.Marshal(&doc)
	if err != nil || out == nil {
		return schema, nil
	}

	return out, nil
}

// injectSchemaDefaults injects stored values as defaults on each property
func injectSchemaDefaults(typeDef *jsonschema.Schema, defs jsonschema.Definitions, stored map[string]any) {
	for pair := typeDef.Properties.Oldest(); pair != nil; pair = pair.Next() {
		k, prop := pair.Key, pair.Value

		if storedVal, ok := stored[k]; ok {
			prop.Default = storedVal

			if prop.Ref != "" {
				if storedSubMap, ok := storedVal.(map[string]any); ok {
					if nestedDef, ok := defs[path.Base(prop.Ref)]; ok && nestedDef.Properties != nil {
						injectSchemaDefaults(nestedDef, defs, storedSubMap)
					}
				}
			}
		}
	}
}
