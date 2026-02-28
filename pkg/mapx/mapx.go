package mapx

import (
	"maps"
	"reflect"
	"strings"
)

// DeepCloneMapAny creates a deep copy of a map[string]any
func DeepCloneMapAny(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}

	dst := make(map[string]any, len(src))
	for key, value := range src {
		dst[key] = cloneAny(value)
	}

	return dst
}

// cloneAny creates a deep copy of any value, handling maps and slices recursively
func cloneAny(value any) any {
	switch v := value.(type) {
	case nil:
		return nil
	case map[string]any:
		return DeepCloneMapAny(v)
	case []any:
		cloned := make([]any, len(v))
		for i, item := range v {
			cloned[i] = cloneAny(item)
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

// CloneMapStringSlice clones a map[string][]string, skipping blank keys
func CloneMapStringSlice(src map[string][]string) map[string][]string {
	if len(src) == 0 {
		return nil
	}

	dst := make(map[string][]string, len(src))
	for key, value := range src {
		if strings.TrimSpace(key) == "" {
			continue
		}

		dst[key] = append([]string(nil), value...)
	}

	if len(dst) == 0 {
		return nil
	}

	return dst
}

// PruneMapZeroAny removes zero-value leaves from a nested map[string]any
func PruneMapZeroAny(src map[string]any) map[string]any {
	pruned := make(map[string]any, len(src))
	for key, value := range src {
		nested, ok := value.(map[string]any)
		if ok {
			nested = PruneMapZeroAny(nested)
			if len(nested) > 0 {
				pruned[key] = nested
			}

			continue
		}

		if value == nil {
			continue
		}

		if _, isBool := value.(bool); isBool || !reflect.ValueOf(value).IsZero() {
			pruned[key] = value
		}
	}

	return pruned
}

// DeepMergeMapAny deep-merges override onto base, recursing into nested map[string]any values
func DeepMergeMapAny(base, override map[string]any) map[string]any {
	merged := make(map[string]any, len(base)+len(override))
	maps.Copy(merged, base)

	for key, overrideValue := range override {
		baseNested, baseIsMap := merged[key].(map[string]any)
		overrideNested, overrideIsMap := overrideValue.(map[string]any)
		if baseIsMap && overrideIsMap {
			merged[key] = DeepMergeMapAny(baseNested, overrideNested)
			continue
		}

		merged[key] = overrideValue
	}

	return merged
}

// MapSetFromSlice converts a slice into a set represented as map[T]struct{}
func MapSetFromSlice[T comparable](items []T) map[T]struct{} {
	set := make(map[T]struct{}, len(items))
	for _, item := range items {
		set[item] = struct{}{}
	}

	return set
}

// MapIntersectionUnique returns items present in both slices, in right-hand order, with duplicates removed
func MapIntersectionUnique[T comparable](left, right []T) []T {
	leftSet := MapSetFromSlice(left)
	seen := make(map[T]struct{}, len(right))
	intersection := make([]T, 0)

	for _, item := range right {
		if _, exists := leftSet[item]; !exists {
			continue
		}

		if _, alreadySeen := seen[item]; alreadySeen {
			continue
		}

		seen[item] = struct{}{}
		intersection = append(intersection, item)
	}

	return intersection
}
