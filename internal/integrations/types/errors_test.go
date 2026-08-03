package types //nolint:revive

import (
	"errors"
	"fmt"
	"testing"
)

var errTestCause = errors.New("token exchange failed")

func TestUnhealthyError(t *testing.T) {
	t.Parallel()

	err := Unhealthy(errTestCause, "the connection needs to be reauthorized")

	if got, want := err.Error(), "the connection needs to be reauthorized: token exchange failed"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}

	if !errors.Is(err, errTestCause) {
		t.Fatal("expected errors.Is to find the wrapped cause")
	}
}

func TestUnhealthyFrom(t *testing.T) {
	t.Parallel()

	unhealthy, ok := UnhealthyFrom(Unhealthy(errTestCause, "needs reauthorization"))
	if !ok {
		t.Fatal("expected UnhealthyFrom to find a direct UnhealthyError")
	}

	if unhealthy.Reason != "needs reauthorization" {
		t.Fatalf("Reason = %q, want %q", unhealthy.Reason, "needs reauthorization")
	}
}

func TestUnhealthyFromWrappedChain(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("operation failed: %w", fmt.Errorf("ingest: %w", Unhealthy(errTestCause, "needs reauthorization")))

	unhealthy, ok := UnhealthyFrom(wrapped)
	if !ok {
		t.Fatal("expected UnhealthyFrom to find an UnhealthyError through wrapping")
	}

	if unhealthy.Reason != "needs reauthorization" {
		t.Fatalf("Reason = %q, want %q", unhealthy.Reason, "needs reauthorization")
	}
}

func TestUnhealthyFromPlainError(t *testing.T) {
	t.Parallel()

	if _, ok := UnhealthyFrom(errTestCause); ok {
		t.Fatal("expected UnhealthyFrom to reject a plain error")
	}

	if _, ok := UnhealthyFrom(nil); ok {
		t.Fatal("expected UnhealthyFrom to reject nil")
	}
}
