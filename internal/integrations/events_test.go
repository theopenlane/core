package integrations

import "testing"

func TestNewIntegrationOperationEnvelopeDefaults(t *testing.T) {
	envelope := NewIntegrationOperationEnvelope(IntegrationOperationRequestedPayload{
		RunID:     "run_123",
		OrgID:     "org_123",
		Provider:  "github",
		Operation: "vulnerabilities.collect",
	})

	if envelope.Type != IntegrationOperationTypeImport {
		t.Fatalf("expected import type, got %q", envelope.Type)
	}
	if envelope.TimeoutSeconds != 600 {
		t.Fatalf("expected timeout 600, got %d", envelope.TimeoutSeconds)
	}
	if envelope.MaxAttempts != 6 {
		t.Fatalf("expected max attempts 6, got %d", envelope.MaxAttempts)
	}

	headers := envelope.Headers()
	if headers.Queue != IntegrationQueueName {
		t.Fatalf("expected queue %q, got %q", IntegrationQueueName, headers.Queue)
	}
	if headers.MaxAttempts != 6 {
		t.Fatalf("expected header max attempts 6, got %d", headers.MaxAttempts)
	}
	if headers.IdempotencyKey != "github:vulnerabilities.collect:org_123:run_123" {
		t.Fatalf("unexpected idempotency key %q", headers.IdempotencyKey)
	}
}

func TestNewIntegrationOperationEnvelopeWebhookPolicy(t *testing.T) {
	envelope := NewIntegrationOperationEnvelope(IntegrationOperationRequestedPayload{
		RunID:     "run_456",
		OrgID:     "org_456",
		Provider:  "github",
		Operation: "alerts.webhook.receive",
	})

	if envelope.Type != IntegrationOperationTypeWebhook {
		t.Fatalf("expected webhook type, got %q", envelope.Type)
	}
	if envelope.TimeoutSeconds != 30 {
		t.Fatalf("expected timeout 30, got %d", envelope.TimeoutSeconds)
	}
	if envelope.MaxAttempts != 3 {
		t.Fatalf("expected max attempts 3, got %d", envelope.MaxAttempts)
	}
}
