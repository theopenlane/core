package integrations

import (
	"fmt"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// IntegrationOperationRequestedPayload captures a queued integration operation request
type IntegrationOperationRequestedPayload struct {
	// RunID is the integration run identifier
	RunID string `json:"run_id"`
	// OrgID is the organization requesting the run
	OrgID string `json:"org_id"`
	// Provider is the provider identifier
	Provider string `json:"provider"`
	// Operation is the operation identifier
	Operation string `json:"operation"`
	// OperationKind is the operation kind from the provider descriptor
	OperationKind types.OperationKind `json:"operation_kind,omitempty"`
	// RunType identifies how the run was triggered
	RunType enums.IntegrationRunType `json:"run_type,omitempty"`
	// Force indicates the run should refresh credentials
	Force bool `json:"force,omitempty"`
	// ClientForce forces a client refresh
	ClientForce bool `json:"client_force,omitempty"`
	// WorkflowInstanceID links the run back to a workflow instance
	WorkflowInstanceID string `json:"workflow_instance_id,omitempty"`
	// WorkflowActionKey identifies the originating workflow action
	WorkflowActionKey string `json:"workflow_action_key,omitempty"`
	// WorkflowActionIndex identifies the action index for the workflow action
	WorkflowActionIndex int `json:"workflow_action_index,omitempty"`
	// WorkflowObjectID is the object id the workflow action targets
	WorkflowObjectID string `json:"workflow_object_id,omitempty"`
	// WorkflowObjectType is the object type the workflow action targets
	WorkflowObjectType enums.WorkflowObjectType `json:"workflow_object_type,omitempty"`
}

const (
	// TopicIntegrationOperationRequested is emitted when an integration operation is queued
	TopicIntegrationOperationRequested = "integration.command.execute"
	// IntegrationQueueName is the dedicated Gala queue for integration operation workloads.
	IntegrationQueueName = "integrations"
)

// IntegrationOperationEnvelope captures operation routing and retry policy alongside request payload.
type IntegrationOperationEnvelope struct {
	// Request is the integration operation request payload.
	Request IntegrationOperationRequestedPayload `json:"request"`
	// TimeoutSeconds is the recommended execution timeout budget.
	TimeoutSeconds int `json:"timeout_seconds"`
	// MaxAttempts is the River retry budget for this operation.
	MaxAttempts int `json:"max_attempts"`
	// IdempotencyKey enforces duplicate-safe enqueue semantics.
	IdempotencyKey string `json:"idempotency_key"`
}

type integrationOperationPolicy struct {
	timeoutSeconds int
	maxAttempts    int
}

const (
	integrationSyncTimeoutSeconds = 120
	integrationSyncMaxAttempts    = 5

	integrationLongRunningTimeoutSeconds = 600
	integrationLongRunningMaxAttempts    = 6

	integrationWebhookTimeoutSeconds = 30
	integrationWebhookMaxAttempts    = 3
)

var integrationDefaultOperationPolicy = integrationOperationPolicy{
	timeoutSeconds: integrationSyncTimeoutSeconds,
	maxAttempts:    integrationSyncMaxAttempts,
}

var integrationLongRunningOperationPolicy = integrationOperationPolicy{
	timeoutSeconds: integrationLongRunningTimeoutSeconds,
	maxAttempts:    integrationLongRunningMaxAttempts,
}

var integrationWebhookOperationPolicy = integrationOperationPolicy{
	timeoutSeconds: integrationWebhookTimeoutSeconds,
	maxAttempts:    integrationWebhookMaxAttempts,
}

// NewIntegrationOperationEnvelope derives queue and retry policy for a requested integration operation.
func NewIntegrationOperationEnvelope(payload IntegrationOperationRequestedPayload) IntegrationOperationEnvelope {
	policy := integrationOperationPolicyForPayload(payload)

	return IntegrationOperationEnvelope{
		Request:        payload,
		TimeoutSeconds: policy.timeoutSeconds,
		MaxAttempts:    policy.maxAttempts,
		IdempotencyKey: integrationOperationIdempotencyKey(payload),
	}
}

// Headers converts envelope policy into Gala dispatch headers.
func (e IntegrationOperationEnvelope) Headers() gala.Headers {
	return gala.Headers{
		IdempotencyKey: e.IdempotencyKey,
		Queue:          IntegrationQueueName,
		MaxAttempts:    e.MaxAttempts,
	}
}

// integrationOperationPolicyForPayload derives runtime retry and timeout policy from explicit request metadata
func integrationOperationPolicyForPayload(payload IntegrationOperationRequestedPayload) integrationOperationPolicy {
	if payload.RunType == enums.IntegrationRunTypeWebhook {
		return integrationWebhookOperationPolicy
	}

	switch payload.OperationKind {
	case types.OperationKindCollectFindings, types.OperationKindScanSettings:
		return integrationLongRunningOperationPolicy
	default:
		return integrationDefaultOperationPolicy
	}
}

// integrationOperationIdempotencyKey builds a deterministic idempotency key for queue dedupe
func integrationOperationIdempotencyKey(payload IntegrationOperationRequestedPayload) string {
	runID := payload.RunID
	orgID := payload.OrgID
	provider := payload.Provider
	operation := payload.Operation

	if runID != "" {
		return fmt.Sprintf("%s:%s:%s:%s", provider, operation, orgID, runID)
	}

	return fmt.Sprintf("%s:%s:%s", provider, operation, orgID)
}

// IntegrationOperationRequestedTopic is emitted when an integration operation is queued.
var IntegrationOperationRequestedTopic = gala.Topic[IntegrationOperationEnvelope]{
	Name: gala.TopicName(TopicIntegrationOperationRequested),
}
