package helpers

import "maps"

// DeepCloneMap creates a deep copy of a map[string]any, recursively cloning nested maps
func DeepCloneMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}

	dst := make(map[string]any, len(src))
	for key, value := range src {
		dst[key] = cloneValue(value)
	}

	return dst
}

// cloneValue recursively clones a value, handling nested maps and slices
func cloneValue(value any) any {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case map[string]any:
		return DeepCloneMap(v)
	case []any:
		cloned := make([]any, len(v))
		for i, item := range v {
			cloned[i] = cloneValue(item)
		}
		return cloned
	case map[string]string:
		return maps.Clone(v)
	case []string:
		return append([]string(nil), v...)
	default:
		return value
	}
}
