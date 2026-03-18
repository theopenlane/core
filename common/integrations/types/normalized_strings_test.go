package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeStringSlice(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		expect []string
	}{
		{"nil input", nil, nil},
		{"empty input", []string{}, nil},
		{"all empty strings", []string{"", " ", "  "}, nil},
		{"trims and deduplicates", []string{" a ", "b", " a"}, []string{"a", "b"}},
		{"preserves order", []string{"c", "b", "a"}, []string{"c", "b", "a"}},
		{"single value", []string{"x"}, []string{"x"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := NormalizeStringSlice(tc.input)
			assert.Equal(t, tc.expect, result)
		})
	}
}

func TestNormalizedStrings(t *testing.T) {
	var trimmed TrimmedString
	if err := trimmed.UnmarshalText([]byte("  value  ")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trimmed.String() != "value" {
		t.Fatalf("expected trimmed value, got %q", trimmed.String())
	}

	var lower LowerString
	_ = lower.UnmarshalText([]byte("  FooBar "))
	if lower.String() != "foobar" {
		t.Fatalf("expected lower value, got %q", lower.String())
	}

	var upper UpperString
	_ = upper.UnmarshalText([]byte("  FooBar "))
	if upper.String() != "FOOBAR" {
		t.Fatalf("expected upper value, got %q", upper.String())
	}
}
