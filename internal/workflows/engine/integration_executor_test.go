package engine

import (
	"context"
	"testing"
	"time"
)

func TestIntegrationOperationContextWithoutTimeout(t *testing.T) {
	parent := context.Background()

	ctx, cancel := integrationOperationContext(parent, 0)
	defer cancel()

	if ctx != parent {
		t.Fatalf("expected parent context when timeout disabled")
	}

	if _, ok := ctx.Deadline(); ok {
		t.Fatalf("expected no deadline when timeout disabled")
	}
}

func TestIntegrationOperationContextWithTimeout(t *testing.T) {
	parent := context.Background()

	ctx, cancel := integrationOperationContext(parent, 30)
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatalf("expected deadline when timeout enabled")
	}

	remaining := time.Until(deadline)
	if remaining <= 0 {
		t.Fatalf("expected positive time remaining")
	}
	if remaining > 31*time.Second {
		t.Fatalf("expected timeout near 30s, got %s", remaining)
	}
}
