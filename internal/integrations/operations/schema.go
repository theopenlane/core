package operations

import (
	"encoding/json"

	"github.com/invopop/jsonschema"

	"github.com/theopenlane/core/pkg/jsonx"
)

var schemaReflector = &jsonschema.Reflector{
	AllowAdditionalProperties:  false,
	RequiredFromJSONSchemaTags: true,
}

// SchemaFrom reflects a JSON schema from a Go type and returns it as raw JSON.
func SchemaFrom[T any]() json.RawMessage {
	schema := schemaReflector.Reflect(new(T))

	out, err := jsonx.ToRawMessage(schema)
	if err != nil {
		return nil
	}

	return out
}
