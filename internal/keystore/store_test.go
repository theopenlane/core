package keystore

import (
	"reflect"
	"testing"
)

func TestMergeMetadata(t *testing.T) {
	existing := map[string]any{
		"environment": "prod",
		"region":      "us-central1",
		"scopes":      []any{"repo", "user"},
	}
	updates := map[string]any{
		"environment": "dev",
		"region":      nil,
		"alias":       "primary",
	}

	merged := mergeMetadata(existing, updates, []string{"user", "admin"})

	if merged["environment"] != "dev" {
		t.Fatalf("expected environment override to dev, got %v", merged["environment"])
	}
	if _, ok := merged["region"]; ok {
		t.Fatalf("expected region to be removed when update value nil")
	}
	if merged["alias"] != "primary" {
		t.Fatalf("expected alias metadata to be set, got %v", merged["alias"])
	}
	scopes, ok := merged["scopes"].([]string)
	if !ok {
		t.Fatalf("expected scopes stored as []string, got %T", merged["scopes"])
	}
	expectedScopes := []string{"admin", "user"}
	if !reflect.DeepEqual(scopes, expectedScopes) {
		t.Fatalf("unexpected scopes: got %v want %v", scopes, expectedScopes)
	}

	// Ensure merge didn't mutate original map.
	if existing["environment"] != "prod" {
		t.Fatalf("mergeMetadata mutated existing map")
	}
}

func TestStringSliceFromAny(t *testing.T) {
	cases := []struct {
		name     string
		input    any
		expected []string
	}{
		{"nil", nil, nil},
		{"string", "alpha beta", []string{"alpha beta"}},
		{"slice", []string{"one", "two", "one"}, []string{"one", "two"}},
		{"any slice", []any{"x", 2}, []string{"2", "x"}},
		{"other", 42, []string{"42"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := stringSliceFromAny(tc.input)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("stringSliceFromAny(%v) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestUniqueStrings(t *testing.T) {
	out := uniqueStrings([]string{" beta ", "alpha", "ALPHA", ""})
	expected := []string{"alpha", "beta"}
	if !reflect.DeepEqual(out, expected) {
		t.Fatalf("uniqueStrings returned %v want %v", out, expected)
	}

	if uniqueStrings(nil) != nil {
		t.Fatalf("expected nil input to return nil slice")
	}
}
