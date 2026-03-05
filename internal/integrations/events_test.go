package integrations

import (
	"testing"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestNewIntegrationOperationEnvelopeDefaults(t *testing.T) {
	envelope := NewIntegrationOperationEnvelope(IntegrationOperationRequestedPayload{
		RunID:     "run_123",
		OrgID:     "org_123",
		Provider:  "github",
		Operation: "health.default",
	})

	if envelope.TimeoutSeconds != 120 {
		t.Fatalf("expected timeout 120, got %d", envelope.TimeoutSeconds)
	}
	if envelope.MaxAttempts != 5 {
		t.Fatalf("expected max attempts 5, got %d", envelope.MaxAttempts)
	}

	headers := envelope.Headers()
	if headers.Queue != IntegrationQueueName {
		t.Fatalf("expected queue %q, got %q", IntegrationQueueName, headers.Queue)
	}
	if headers.MaxAttempts != 5 {
		t.Fatalf("expected header max attempts 5, got %d", headers.MaxAttempts)
	}
	if headers.IdempotencyKey != "github:health.default:org_123:run_123" {
		t.Fatalf("unexpected idempotency key %q", headers.IdempotencyKey)
	}
}

func TestNewIntegrationOperationEnvelopeWebhookPolicy(t *testing.T) {
	envelope := NewIntegrationOperationEnvelope(IntegrationOperationRequestedPayload{
		RunID:         "run_456",
		OrgID:         "org_456",
		Provider:      "github",
		Operation:     "alerts.webhook.receive",
		OperationKind: types.OperationKindCollectFindings,
		RunType:       enums.IntegrationRunTypeWebhook,
	})

	if envelope.TimeoutSeconds != 30 {
		t.Fatalf("expected timeout 30, got %d", envelope.TimeoutSeconds)
	}
	if envelope.MaxAttempts != 3 {
		t.Fatalf("expected max attempts 3, got %d", envelope.MaxAttempts)
	}
}

func TestNewIntegrationOperationEnvelopeLongRunningPolicy(t *testing.T) {
	envelope := NewIntegrationOperationEnvelope(IntegrationOperationRequestedPayload{
		RunID:         "run_789",
		OrgID:         "org_789",
		Provider:      "github",
		Operation:     "vulnerabilities.collect",
		OperationKind: types.OperationKindCollectFindings,
		RunType:       enums.IntegrationRunTypeEvent,
	})

	if envelope.TimeoutSeconds != 600 {
		t.Fatalf("expected timeout 600, got %d", envelope.TimeoutSeconds)
	}
	if envelope.MaxAttempts != 6 {
		t.Fatalf("expected max attempts 6, got %d", envelope.MaxAttempts)
	}
}
