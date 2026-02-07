package auth

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	if ParseDuration("") != 0 {
		t.Fatalf("expected zero duration for empty input")
	}
	if ParseDuration("not-a-duration") != 0 {
		t.Fatalf("expected zero duration for invalid input")
	}
	if got := ParseDuration("30m"); got != 30*time.Minute {
		t.Fatalf("expected 30m, got %v", got)
	}
}
