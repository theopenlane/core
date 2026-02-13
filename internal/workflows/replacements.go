package workflows

import (
	"fmt"

	"github.com/theopenlane/core/pkg/jsonx"
)

// BuildObjectReplacements builds string replacements from object fields for template substitution
// Keys include both top-level field names and "object.<field>" paths
func BuildObjectReplacements(obj *Object) map[string]string {
	if obj == nil {
		return nil
	}

	replacements := map[string]string{
		"object_id":   obj.ID,
		"object_type": obj.Type.String(),
	}

	node := obj.Node
	if node == nil {
		node = obj.CELValue()
	}

	payload, err := jsonx.ToMap(node)
	if err != nil {
		return replacements
	}

	// Add top-level fields without prefix.
	for key, value := range payload {
		addReplacementValue(replacements, key, value)
	}

	// Add object.<field> paths.
	addReplacementValue(replacements, "object", payload)

	return replacements
}

// addReplacementValue adds replacement entries for a given key and value
func addReplacementValue(out map[string]string, key string, value any) {
	switch typed := value.(type) {
	case map[string]any:
		for nestedKey, nestedVal := range typed {
			addReplacementValue(out, key+"."+nestedKey, nestedVal)
		}
	case []any:
		// Skip slices for now to avoid ambiguous stringification nasties
		return
	case nil:
		return
	default:
		out[key] = fmt.Sprintf("%v", typed)
	}
}
