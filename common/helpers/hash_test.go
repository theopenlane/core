package helpers

import (
	"testing"
	"time"
)

func TestHashBuilder_WriteStrings(t *testing.T) {
	t.Run("identical inputs produce identical hashes", func(t *testing.T) {
		hash1 := NewHashBuilder().WriteStrings("foo", "bar", "baz").Hex()
		hash2 := NewHashBuilder().WriteStrings("foo", "bar", "baz").Hex()

		if hash1 != hash2 {
			t.Errorf("expected identical hashes, got %s and %s", hash1, hash2)
		}
	})

	t.Run("different inputs produce different hashes", func(t *testing.T) {
		hash1 := NewHashBuilder().WriteStrings("foo", "bar").Hex()
		hash2 := NewHashBuilder().WriteStrings("foo", "baz").Hex()

		if hash1 == hash2 {
			t.Errorf("expected different hashes, got %s for both", hash1)
		}
	})

	t.Run("empty strings are skipped", func(t *testing.T) {
		hash1 := NewHashBuilder().WriteStrings("foo", "", "bar").Hex()
		hash2 := NewHashBuilder().WriteStrings("foo", "bar").Hex()

		if hash1 != hash2 {
			t.Errorf("expected identical hashes (empty strings skipped), got %s and %s", hash1, hash2)
		}
	})
}

func TestHashBuilder_WriteTime(t *testing.T) {
	t.Run("identical times produce identical hashes", func(t *testing.T) {
		ts := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

		hash1 := NewHashBuilder().WriteTime(ts).Hex()
		hash2 := NewHashBuilder().WriteTime(ts).Hex()

		if hash1 != hash2 {
			t.Errorf("expected identical hashes, got %s and %s", hash1, hash2)
		}
	})

	t.Run("zero time is skipped", func(t *testing.T) {
		hash1 := NewHashBuilder().WriteStrings("test").WriteTime(time.Time{}).Hex()
		hash2 := NewHashBuilder().WriteStrings("test").Hex()

		if hash1 != hash2 {
			t.Errorf("expected identical hashes (zero time skipped), got %s and %s", hash1, hash2)
		}
	})
}

func TestHashBuilder_WriteSortedMap(t *testing.T) {
	t.Run("maps with same content produce identical hashes regardless of iteration order", func(t *testing.T) {
		map1 := map[string]any{"z": "value1", "a": "value2", "m": "value3"}
		map2 := map[string]any{"a": "value2", "m": "value3", "z": "value1"}

		hash1 := NewHashBuilder().WriteSortedMap(map1).Hex()
		hash2 := NewHashBuilder().WriteSortedMap(map2).Hex()

		if hash1 != hash2 {
			t.Errorf("expected identical hashes for same map content, got %s and %s", hash1, hash2)
		}
	})

	t.Run("handles complex values", func(t *testing.T) {
		m := map[string]any{
			"string":  "value",
			"number":  42,
			"boolean": true,
			"nested":  map[string]string{"key": "val"},
		}

		hash := NewHashBuilder().WriteSortedMap(m).Hex()
		if hash == "" {
			t.Error("expected non-empty hash")
		}
	})
}

func TestHashBuilder_Chaining(t *testing.T) {
	t.Run("builder supports method chaining", func(t *testing.T) {
		ts := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		m := map[string]any{"key": "value"}

		hash := NewHashBuilder().
			WriteStrings("foo", "bar").
			WriteTime(ts).
			WriteSortedMap(m).
			Hex()

		if hash == "" {
			t.Error("expected non-empty hash from chained operations")
		}
	})
}
