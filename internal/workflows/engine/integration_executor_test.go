package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/integrations/types"
	ent "github.com/theopenlane/core/internal/ent/generated"
	integrationscope "github.com/theopenlane/core/internal/integrations/scope"
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

func TestIntegrationRunOperationKind(t *testing.T) {
	if got := integrationRunOperationKind(enums.IntegrationRunTypeEvent, types.OperationKindNotify); got != enums.IntegrationOperationKindPush {
		t.Fatalf("expected notify to map to push, got %q", got)
	}

	if got := integrationRunOperationKind(enums.IntegrationRunTypeEvent, types.OperationKindCollectFindings); got != enums.IntegrationOperationKindPull {
		t.Fatalf("expected collect findings to map to pull, got %q", got)
	}

	if got := integrationRunOperationKind(enums.IntegrationRunTypeWebhook, types.OperationKindCollectFindings); got != enums.IntegrationOperationKindWebhook {
		t.Fatalf("expected webhook run type to map to webhook, got %q", got)
	}
}

func TestEvaluateIntegrationScope(t *testing.T) {
	evaluator, err := integrationscope.NewEvaluator(integrationscope.DefaultEvaluatorConfig())
	if err != nil {
		t.Fatalf("failed to create scope evaluator: %v", err)
	}

	integrationRecord := &ent.Integration{ID: "int_123"}

	allowed, err := evaluateIntegrationScope(context.Background(), evaluator, IntegrationQueueRequest{
		OrgID:           "org_123",
		ScopeExpression: "provider == 'githubapp'",
	}, integrationRecord, types.ProviderType("githubapp"), types.OperationVulnerabilitiesCollect, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !allowed {
		t.Fatalf("expected scope condition to allow execution")
	}

	allowed, err = evaluateIntegrationScope(context.Background(), evaluator, IntegrationQueueRequest{
		OrgID:           "org_123",
		ScopeExpression: "provider == 'slack'",
	}, integrationRecord, types.ProviderType("githubapp"), types.OperationVulnerabilitiesCollect, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if allowed {
		t.Fatalf("expected scope condition to reject execution")
	}

	_, err = evaluateIntegrationScope(context.Background(), evaluator, IntegrationQueueRequest{
		OrgID:           "org_123",
		ScopeExpression: "provider =",
	}, integrationRecord, types.ProviderType("githubapp"), types.OperationVulnerabilitiesCollect, nil, nil)
	if !errors.Is(err, integrationscope.ErrScopeCompilationFailed) {
		t.Fatalf("expected ErrScopeCompilationFailed, got %v", err)
	}
}
