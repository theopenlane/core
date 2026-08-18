package jsonx

import "encoding/json"

// Stringify renders an arbitrary decoded JSON value as its JSON string form for
// downstream text processing. Strings pass through unchanged; nil, empty slices, JSON
// null, and unmarshalable values render as the empty string
func Stringify(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case []any:
		if len(typed) == 0 {
			return ""
		}
	case []string:
		if len(typed) == 0 {
			return ""
		}
	}

	raw, err := json.Marshal(value)
	if err != nil || len(raw) == 0 || string(raw) == "null" {
		return ""
	}

	return string(raw)
}
