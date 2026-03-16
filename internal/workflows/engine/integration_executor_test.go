package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	ent "github.com/theopenlane/core/internal/ent/generated"
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

func TestEvaluateInstallationScope(t *testing.T) {
	evaluator, err := NewIntegrationScopeEvaluator()
	if err != nil {
		t.Fatalf("failed to create scope evaluator: %v", err)
	}

	record := &ent.Integration{
		ID:             "int_123",
		DefinitionSlug: "github_app",
	}

	opName := "vulnerability.collect"

	allowed, err := evaluateInstallationScope(context.Background(), evaluator, IntegrationQueueRequest{
		OrgID:           "org_123",
		ScopeExpression: "provider == 'github_app'",
	}, record, opName, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !allowed {
		t.Fatalf("expected scope condition to allow execution")
	}

	allowed, err = evaluateInstallationScope(context.Background(), evaluator, IntegrationQueueRequest{
		OrgID:           "org_123",
		ScopeExpression: "provider == 'slack'",
	}, record, opName, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if allowed {
		t.Fatalf("expected scope condition to reject execution")
	}

	_, err = evaluateInstallationScope(context.Background(), evaluator, IntegrationQueueRequest{
		OrgID:           "org_123",
		ScopeExpression: "provider =",
	}, record, opName, nil)
	if !errors.Is(err, ErrCELCompilationFailed) {
		t.Fatalf("expected ErrCELCompilationFailed, got %v", err)
	}
}
