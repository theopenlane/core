package workflows

import "encoding/json"

// StringField extracts a string field by name from a struct or pointer to struct
func StringField(node any, field string) string {
	if node == nil {
		return ""
	}

	data, err := json.Marshal(node)
	if err != nil {
		return ""
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return ""
	}

	raw, ok := m[field]
	if !ok || raw == nil {
		return ""
	}

	switch v := raw.(type) {
	case string:
		return v
	case *string:
		return *v
	case []byte:
		return string(v)
	default:
		return ""
	}
}
