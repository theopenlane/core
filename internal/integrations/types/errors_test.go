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

func TestTerminalError(t *testing.T) {
	t.Parallel()

	err := Terminal(errTestCause, "the recipient address is malformed")

	if got, want := err.Error(), "the recipient address is malformed: token exchange failed"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}

	if !errors.Is(err, errTestCause) {
		t.Fatal("expected errors.Is to find the wrapped cause")
	}
}

func TestTerminalFrom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    error
		want   bool
		reason string
	}{
		{name: "direct terminal error", err: Terminal(errTestCause, "malformed"), want: true, reason: "malformed"},
		{name: "terminal error through wrapping", err: fmt.Errorf("operation failed: %w", fmt.Errorf("send: %w", Terminal(errTestCause, "malformed"))), want: true, reason: "malformed"},
		{name: "plain error", err: errTestCause},
		{name: "unhealthy error is not terminal", err: Unhealthy(errTestCause, "needs reauthorization")},
		{name: "nil", err: nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			terminal, ok := TerminalFrom(tc.err)
			if ok != tc.want {
				t.Fatalf("TerminalFrom() ok = %v, want %v", ok, tc.want)
			}

			if ok && terminal.Reason != tc.reason {
				t.Fatalf("Reason = %q, want %q", terminal.Reason, tc.reason)
			}
		})
	}
}

func TestTerminalIsNotAHealthMarker(t *testing.T) {
	t.Parallel()

	err := Terminal(errTestCause, "malformed")

	if _, ok := UnhealthyFrom(err); ok {
		t.Fatal("expected a terminal error not to read as unhealthy")
	}

	if _, ok := DegradedFrom(err); ok {
		t.Fatal("expected a terminal error not to read as degraded")
	}
}
