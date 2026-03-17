package providerkit

import (
	"encoding/json"

	"github.com/invopop/jsonschema"

	"github.com/theopenlane/core/pkg/jsonx"
)

// schemaReflector is the shared JSON schema reflector used by SchemaFrom
var schemaReflector = &jsonschema.Reflector{
	AllowAdditionalProperties:  false,
	RequiredFromJSONSchemaTags: true,
}

// SchemaFrom reflects a JSON schema from a Go type and returns it as raw JSON
// Returns nil if schema marshaling fails.
func SchemaFrom[T any]() json.RawMessage {
	schema := schemaReflector.Reflect(new(T))

	out, err := jsonx.ToRawMessage(schema)
	if err != nil {
		return nil
	}

	return out
}
