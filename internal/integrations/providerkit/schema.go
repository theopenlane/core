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

// WebhookEventSchema reflects a webhook event payload type and returns both the JSON schema
// and a typed webhook event ref whose name is derived from the schema definition key
func WebhookEventSchema[T any]() (json.RawMessage, types.WebhookEventRef[T]) {
	schema := SchemaFrom[T]()

	return schema, types.NewWebhookEventRef[T](SchemaID(schema))
}
