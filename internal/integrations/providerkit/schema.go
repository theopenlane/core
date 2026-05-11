package providerkit

import (
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// CredentialSchema reflects a credential schema type and returns both the JSON schema
// and a typed credential ref whose slot identity is derived from the schema definition key
func CredentialSchema[T any]() (json.RawMessage, types.CredentialRef[T]) {
	schema := jsonx.SchemaFrom[T]()

	return schema, types.NewCredentialRef[T](jsonx.SchemaID(schema))
}

// OperationSchema reflects an operation config type and returns both the JSON schema
// and a typed operation ref whose name is derived from the schema definition key
func OperationSchema[T any]() (json.RawMessage, types.OperationRef[T]) {
	schema := jsonx.SchemaFrom[T]()

	return schema, types.NewOperationRef[T](jsonx.SchemaID(schema))
}

// OperationSchemaVariant reflects the same schema as OperationSchema but registers
// the operation under a composite key of TypeName.variant, allowing the same input
// type to back multiple catalog entries (e.g. themed visual variants)
func OperationSchemaVariant[T any](variant string) (json.RawMessage, types.OperationRef[T]) {
	schema := jsonx.SchemaFrom[T]()

	return schema, types.NewOperationRef[T](jsonx.SchemaID(schema) + "." + variant)
}

// WebhookEventSchema reflects a webhook event payload type and returns both the JSON schema
// and a typed webhook event ref whose name is derived from the schema definition key
func WebhookEventSchema[T any]() (json.RawMessage, types.WebhookEventRef[T]) {
	schema := jsonx.SchemaFrom[T]()

	return schema, types.NewWebhookEventRef[T](jsonx.SchemaID(schema))
}

// RuntimeSchema reflects a runtime integration config type and returns both the
// JSON schema and a typed runtime integration ref whose identity is derived from the schema definition key
func RuntimeSchema[T any]() (json.RawMessage, types.RuntimeRef[T]) {
	schema := jsonx.SchemaFrom[T]()

	return schema, types.NewRuntimeRef[T](jsonx.SchemaID(schema), schema)
}
