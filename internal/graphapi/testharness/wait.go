//go:build test

package testharness

import (
	"testing"
	"time"
)

// WaitForCondition polls condition until it holds or the deadline passes
func WaitForCondition(t *testing.T, condition func() bool, msg string) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for condition: %s", msg)
}
