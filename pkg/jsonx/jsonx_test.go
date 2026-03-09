package jsonx

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestDeepMerge(t *testing.T) {
	t.Parallel()

	t.Run("empty patch returns base unchanged", func(t *testing.T) {
		t.Parallel()
		base := json.RawMessage(`{"a":1}`)
		merged, changed, err := DeepMerge(base, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if changed {
			t.Fatal("expected no change")
		}
		if string(merged) != string(base) {
			t.Fatalf("expected base unchanged, got %s", merged)
		}
	})

	t.Run("patch adds new key", func(t *testing.T) {
		t.Parallel()
		base := json.RawMessage(`{"a":1}`)
		patch := json.RawMessage(`{"b":2}`)
		merged, changed, err := DeepMerge(base, patch)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !changed {
			t.Fatal("expected change")
		}
		var out map[string]any
		if err := json.Unmarshal(merged, &out); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if out["a"] != float64(1) || out["b"] != float64(2) {
			t.Fatalf("unexpected merged output: %v", out)
		}
	})

	t.Run("patch overwrites existing key", func(t *testing.T) {
		t.Parallel()
		base := json.RawMessage(`{"a":1}`)
		patch := json.RawMessage(`{"a":99}`)
		merged, changed, err := DeepMerge(base, patch)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !changed {
			t.Fatal("expected change")
		}
		var out map[string]any
		if err := json.Unmarshal(merged, &out); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if out["a"] != float64(99) {
			t.Fatalf("expected a=99, got %v", out["a"])
		}
	})

	t.Run("patch with same values reports no change", func(t *testing.T) {
		t.Parallel()
		base := json.RawMessage(`{"a":1}`)
		patch := json.RawMessage(`{"a":1}`)
		_, changed, err := DeepMerge(base, patch)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if changed {
			t.Fatal("expected no change for identical values")
		}
	})

	t.Run("nested deep merge", func(t *testing.T) {
		t.Parallel()
		base := json.RawMessage(`{"a":{"x":1,"y":2}}`)
		patch := json.RawMessage(`{"a":{"y":99},"b":3}`)
		merged, changed, err := DeepMerge(base, patch)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !changed {
			t.Fatal("expected change")
		}
		var out map[string]any
		if err := json.Unmarshal(merged, &out); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		nested, ok := out["a"].(map[string]any)
		if !ok {
			t.Fatalf("expected nested map, got %T", out["a"])
		}
		if nested["x"] != float64(1) || nested["y"] != float64(99) {
			t.Fatalf("unexpected nested values: %v", nested)
		}
		if out["b"] != float64(3) {
			t.Fatalf("expected b=3, got %v", out["b"])
		}
	})

	t.Run("nil base with patch", func(t *testing.T) {
		t.Parallel()
		patch := json.RawMessage(`{"a":1}`)
		merged, changed, err := DeepMerge(nil, patch)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !changed {
			t.Fatal("expected change")
		}
		var out map[string]any
		if err := json.Unmarshal(merged, &out); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if out["a"] != float64(1) {
			t.Fatalf("expected a=1, got %v", out["a"])
		}
	})
}

type sample struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestRoundTripStruct(t *testing.T) {
	input := sample{Name: "ok", Count: 2}
	var out sample
	if err := RoundTrip(input, &out); err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	if out != input {
		t.Fatalf("RoundTrip mismatch: got %+v want %+v", out, input)
	}
}

func TestRoundTripBytes(t *testing.T) {
	input := []byte(`{"name":"ok","count":3}`)
	var out sample
	if err := RoundTrip(input, &out); err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	if out.Name != "ok" || out.Count != 3 {
		t.Fatalf("RoundTrip mismatch: got %+v", out)
	}
}

func TestRoundTripRawMessage(t *testing.T) {
	input := json.RawMessage(`{"name":"ok","count":4}`)
	var out sample
	if err := RoundTrip(input, &out); err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	if out.Name != "ok" || out.Count != 4 {
		t.Fatalf("RoundTrip mismatch: got %+v", out)
	}
}

func TestToMapStruct(t *testing.T) {
	input := sample{Name: "ok", Count: 5}
	m, err := ToMap(input)
	if err != nil {
		t.Fatalf("ToMap failed: %v", err)
	}
	name, ok := m["name"].(string)
	if !ok || name != "ok" {
		t.Fatalf("ToMap name mismatch: %#v", m["name"])
	}
	count, ok := m["count"].(float64)
	if !ok || count != 5 {
		t.Fatalf("ToMap count mismatch: %#v", m["count"])
	}
}

func TestToMapNil(t *testing.T) {
	m, err := ToMap(nil)
	if err != nil {
		t.Fatalf("ToMap failed: %v", err)
	}
	if m == nil || len(m) != 0 {
		t.Fatalf("ToMap nil mismatch: %#v", m)
	}
}

func TestToMapNonObject(t *testing.T) {
	_, err := ToMap([]string{"a"})
	if !errors.Is(err, ErrObjectExpected) {
		t.Fatalf("expected ErrObjectExpected, got %v", err)
	}
}

func TestCloneRawMessage(t *testing.T) {
	original := json.RawMessage(`{"ok":true}`)
	cloned := CloneRawMessage(original)
	if len(cloned) == 0 {
		t.Fatalf("expected cloned message")
	}

	cloned[0] = '['
	if original[0] != '{' {
		t.Fatalf("expected clone to avoid aliasing original")
	}
}

func TestToRawMessage(t *testing.T) {
	raw, err := ToRawMessage(sample{Name: "ok", Count: 1})
	if err != nil {
		t.Fatalf("unexpected ToRawMessage error: %v", err)
	}
	if string(raw) != `{"name":"ok","count":1}` {
		t.Fatalf("unexpected raw output: %s", string(raw))
	}

	raw, err = ToRawMessage(nil)
	if err != nil {
		t.Fatalf("unexpected ToRawMessage nil error: %v", err)
	}
	if raw != nil {
		t.Fatalf("expected nil raw message for nil input, got %s", string(raw))
	}
}
