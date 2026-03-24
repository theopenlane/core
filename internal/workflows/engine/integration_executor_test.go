package engine

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	ent "github.com/theopenlane/core/internal/ent/generated"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

func TestEvaluateInstallationScope(t *testing.T) {
	evaluator, err := NewIntegrationScopeEvaluator()
	if err != nil {
		t.Fatalf("failed to create scope evaluator: %v", err)
	}

	record := &ent.Integration{
		ID:           "int_123",
		DefinitionID: "def_01K0GHAPP000000000000000001",
	}

	opName := "vulnerability.collect"

	allowed, err := evaluateInstallationScope(context.Background(), evaluator, IntegrationQueueRequest{
		OrgID:           "org_123",
		ScopeExpression: "provider == 'def_01K0GHAPP000000000000000001'",
	}, record, opName, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !allowed {
		t.Fatalf("expected scope condition to allow execution")
	}

	allowed, err = evaluateInstallationScope(context.Background(), evaluator, IntegrationQueueRequest{
		OrgID:           "org_123",
		ScopeExpression: "provider == 'def_01K0SLACK000000000000000001'",
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

func TestEvaluateInstallationScopeUsesClientConfig(t *testing.T) {
	t.Parallel()

	evaluator, err := NewIntegrationScopeEvaluator()
	if err != nil {
		t.Fatalf("failed to create scope evaluator: %v", err)
	}

	record := &ent.Integration{
		ID:           "int_123",
		DefinitionID: "def_01K0GHAPP000000000000000001",
		Metadata: map[string]any{
			"environment": "stale",
		},
		Config: integrationtypes.IntegrationConfig{
			ClientConfig: json.RawMessage(`{"environment":"prod"}`),
		},
	}

	allowed, err := evaluateInstallationScope(context.Background(), evaluator, IntegrationQueueRequest{
		OrgID:           "org_123",
		ScopeExpression: "integration_config.environment == 'prod'",
	}, record, "vulnerability.collect", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !allowed {
		t.Fatal("expected scope condition to read integration.Config.ClientConfig")
	}
}
