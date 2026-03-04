package engine

import (
	"context"
	"encoding/json"
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

func TestShouldEnsurePayloads(t *testing.T) {
	if shouldEnsurePayloads(nil) {
		t.Fatalf("expected false when no contracts are present")
	}

	if shouldEnsurePayloads([]types.IngestContract{
		{Schema: types.MappingSchemaVulnerability},
		{Schema: types.MappingSchemaDirectoryAccount},
	}) {
		t.Fatalf("expected false when contracts do not require payloads")
	}

	if !shouldEnsurePayloads([]types.IngestContract{
		{Schema: types.MappingSchemaVulnerability, EnsurePayloads: true},
	}) {
		t.Fatalf("expected true when any contract requires payloads")
	}
}

func TestExtractIngestBatchesLegacyAlerts(t *testing.T) {
	details := map[string]any{
		"alerts": []map[string]any{
			{
				"alertType": "dependabot",
				"resource":  "repo",
				"payload": map[string]any{
					"id": 1,
				},
			},
		},
	}

	raw, err := json.Marshal(details)
	if err != nil {
		t.Fatalf("marshal details: %v", err)
	}

	batches, err := extractIngestBatches(raw, []types.IngestContract{
		{Schema: types.MappingSchemaVulnerability},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(batches) != 1 {
		t.Fatalf("expected one ingest batch, got %d", len(batches))
	}
	if batches[0].Schema != types.MappingSchemaVulnerability {
		t.Fatalf("expected vulnerability schema, got %q", batches[0].Schema)
	}
	if len(batches[0].Envelopes) != 1 {
		t.Fatalf("expected one envelope, got %d", len(batches[0].Envelopes))
	}
}

func TestExtractIngestBatchesStructured(t *testing.T) {
	details := map[string]any{
		"ingest_batches": []map[string]any{
			{
				"schema": "Vulnerability",
				"envelopes": []map[string]any{
					{
						"alertType": "dependabot",
						"resource":  "repo",
						"payload": map[string]any{
							"id": 1,
						},
					},
				},
			},
			{
				"schema": "DirectoryAccount",
				"envelopes": []map[string]any{
					{
						"alertType": "directory_account",
						"resource":  "user@example.com",
						"payload": map[string]any{
							"id": "u_1",
						},
					},
				},
			},
		},
	}

	raw, err := json.Marshal(details)
	if err != nil {
		t.Fatalf("marshal details: %v", err)
	}

	batches, err := extractIngestBatches(raw, []types.IngestContract{
		{Schema: types.MappingSchemaVulnerability},
		{Schema: types.MappingSchemaDirectoryAccount},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(batches) != 2 {
		t.Fatalf("expected two ingest batches, got %d", len(batches))
	}
}
