package operations

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
)

var schemaReflector = &jsonschema.Reflector{
	AllowAdditionalProperties:  true,
	RequiredFromJSONSchemaTags: true,
}

// SchemaFrom reflects a JSON schema from a Go type and returns it as a map
func SchemaFrom[T any]() map[string]any {
	schema := schemaReflector.Reflect(new(T))

	data, err := json.Marshal(schema)
	if err != nil {
		return nil
	}

	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}

	return out
}
