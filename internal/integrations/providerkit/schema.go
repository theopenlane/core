package providerkit

import (
	"encoding/json"
	"path"

	"github.com/invopop/jsonschema"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// schemaReflector is the shared JSON schema reflector used by SchemaFrom and SchemaID
var schemaReflector = &jsonschema.Reflector{
	AllowAdditionalProperties:  false,
	RequiredFromJSONSchemaTags: true,
}

// SchemaFrom reflects a JSON schema from a Go type and returns it as raw JSON
func SchemaFrom[T any]() json.RawMessage {
	schema := schemaReflector.Reflect(new(T))

	out, err := jsonx.ToRawMessage(schema)
	if err != nil {
		return nil
	}

	return out
}

// PropertyNames reflects a Go type and returns the top-level JSON property names
// from the generated JSON schema. Properties from embedded structs are promoted
// by the reflector and appear as top-level names
func PropertyNames[T any]() []string {
	schema := schemaReflector.Reflect(new(T))

	if schema.Ref != "" {
		defKey := path.Base(schema.Ref)
		if def, ok := schema.Definitions[defKey]; ok {
			schema = def
		}
	}

	if schema.Properties == nil {
		return nil
	}

	names := make([]string, 0, schema.Properties.Len())

	for pair := schema.Properties.Oldest(); pair != nil; pair = pair.Next() {
		names = append(names, pair.Key)
	}

	return names
}

// PropertyDescriptor is a top-level JSON schema property with its name and description
type PropertyDescriptor struct {
	// Name is the JSON property key as it appears in the reflected schema
	Name string
	// Description is the human-readable description extracted from the jsonschema description tag
	Description string
}

// PropertyDescriptors reflects a Go type and returns its top-level JSON properties
// with names and descriptions from the generated JSON schema. Properties from embedded
// structs are promoted by the reflector and appear as top-level entries
func PropertyDescriptors[T any]() []PropertyDescriptor {
	schema := schemaReflector.Reflect(new(T))

	if schema.Ref != "" {
		defKey := path.Base(schema.Ref)
		if def, ok := schema.Definitions[defKey]; ok {
			schema = def
		}
	}

	if schema.Properties == nil {
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

// CredentialSchema reflects a credential schema type and returns both the JSON schema
// and a typed credential ref whose slot identity is derived from the schema definition key
func CredentialSchema[T any]() (json.RawMessage, types.CredentialRef[T]) {
	schema := SchemaFrom[T]()

	return schema, types.NewCredentialRef[T](SchemaID(schema))
}

// OperationSchema reflects an operation config type and returns both the JSON schema
// and a typed operation ref whose name is derived from the schema definition key
func OperationSchema[T any]() (json.RawMessage, types.OperationRef[T]) {
	schema := SchemaFrom[T]()

	return schema, types.NewOperationRef[T](SchemaID(schema))
}

// OperationSchemaVariant reflects the same schema as OperationSchema but registers
// the operation under a composite key of TypeName.variant, allowing the same input
// type to back multiple catalog entries (e.g. themed visual variants)
func OperationSchemaVariant[T any](variant string) (json.RawMessage, types.OperationRef[T]) {
	schema := SchemaFrom[T]()

	return schema, types.NewOperationRef[T](SchemaID(schema) + "." + variant)
}

// WebhookEventSchema reflects a webhook event payload type and returns both the JSON schema
// and a typed webhook event ref whose name is derived from the schema definition key
func WebhookEventSchema[T any]() (json.RawMessage, types.WebhookEventRef[T]) {
	schema := SchemaFrom[T]()

	return schema, types.NewWebhookEventRef[T](SchemaID(schema))
}

// RuntimeSchema reflects a runtime integration config type and returns both the
// JSON schema and a typed runtime integration ref whose identity is derived from the schema definition key
func RuntimeSchema[T any]() (json.RawMessage, types.RuntimeRef[T]) {
	schema := SchemaFrom[T]()

	return schema, types.NewRuntimeRef[T](SchemaID(schema), schema)
}
