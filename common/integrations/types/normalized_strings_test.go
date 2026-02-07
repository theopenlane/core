package types

import "testing"

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
