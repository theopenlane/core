package helpers

import (
	"reflect"
	"testing"
)

func TestDeepCloneMap(t *testing.T) {
	t.Run("nil map returns nil", func(t *testing.T) {
		result := DeepCloneMap(nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("empty map returns empty map", func(t *testing.T) {
		src := map[string]any{}
		result := DeepCloneMap(src)

		if result == nil {
			t.Error("expected non-nil map")
		}
		if len(result) != 0 {
			t.Errorf("expected empty map, got %v", result)
		}
	})

	t.Run("clones simple values", func(t *testing.T) {
		src := map[string]any{
			"string": "value",
			"int":    42,
			"bool":   true,
			"nil":    nil,
		}

		result := DeepCloneMap(src)

		if !reflect.DeepEqual(src, result) {
			t.Errorf("expected %v, got %v", src, result)
		}

		result["string"] = "modified"
		if src["string"] == "modified" {
			t.Error("modifying clone should not affect original")
		}
	})

	t.Run("deep clones nested maps", func(t *testing.T) {
		src := map[string]any{
			"nested": map[string]any{
				"key": "value",
			},
		}

		result := DeepCloneMap(src)

		nested := result["nested"].(map[string]any)
		nested["key"] = "modified"

		original := src["nested"].(map[string]any)
		if original["key"] == "modified" {
			t.Error("modifying nested clone should not affect original")
		}
	})

	t.Run("deep clones slices", func(t *testing.T) {
		src := map[string]any{
			"slice": []any{"a", "b", "c"},
		}

		result := DeepCloneMap(src)

		slice := result["slice"].([]any)
		slice[0] = "modified"

		original := src["slice"].([]any)
		if original[0] == "modified" {
			t.Error("modifying cloned slice should not affect original")
		}
	})

	t.Run("deep clones nested slices with maps", func(t *testing.T) {
		src := map[string]any{
			"items": []any{
				map[string]any{"id": 1},
				map[string]any{"id": 2},
			},
		}

		result := DeepCloneMap(src)

		items := result["items"].([]any)
		firstItem := items[0].(map[string]any)
		firstItem["id"] = 999

		original := src["items"].([]any)
		originalFirst := original[0].(map[string]any)
		if originalFirst["id"] == 999 {
			t.Error("modifying cloned nested structure should not affect original")
		}
	})

	t.Run("handles map[string]string", func(t *testing.T) {
		src := map[string]any{
			"labels": map[string]string{
				"env": "prod",
				"app": "api",
			},
		}

		result := DeepCloneMap(src)

		labels := result["labels"].(map[string]string)
		labels["env"] = "dev"

		original := src["labels"].(map[string]string)
		if original["env"] == "dev" {
			t.Error("modifying cloned map[string]string should not affect original")
		}
	})

	t.Run("handles []string", func(t *testing.T) {
		src := map[string]any{
			"tags": []string{"tag1", "tag2", "tag3"},
		}

		result := DeepCloneMap(src)

		tags := result["tags"].([]string)
		tags[0] = "modified"

		original := src["tags"].([]string)
		if original[0] == "modified" {
			t.Error("modifying cloned []string should not affect original")
		}
	})
}
