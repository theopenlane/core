package operations

import (
	"github.com/invopop/jsonschema"

	"github.com/theopenlane/core/pkg/jsonx"
)

var schemaReflector = &jsonschema.Reflector{
	AllowAdditionalProperties:  true,
	RequiredFromJSONSchemaTags: true,
}

// SchemaFrom reflects a JSON schema from a Go type and returns it as a map
func SchemaFrom[T any]() map[string]any {
	schema := schemaReflector.Reflect(new(T))

	out, err := jsonx.ToMap(schema)
	if err != nil {
		return nil
	}

	return out
}
