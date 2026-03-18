package auth

import "testing"

func TestNormalizeServiceAccountKey(t *testing.T) {
	if got := NormalizeServiceAccountKey(" "); got != "" {
		t.Fatalf("expected empty output, got %q", got)
	}

	if got := NormalizeServiceAccountKey("\"foo\""); got != "foo" {
		t.Fatalf("expected decoded value, got %q", got)
	}

	if got := NormalizeServiceAccountKey("  bar  "); got != "bar" {
		t.Fatalf("expected trimmed value, got %q", got)
	}
}
