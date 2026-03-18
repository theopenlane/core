package jsonx

import (
	"encoding/json"
	"errors"
	"testing"
)

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
