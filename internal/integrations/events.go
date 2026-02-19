package integrations

import (
	"fmt"
	"strings"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/pkg/gala"
)

// IntegrationOperationType classifies integration workload behavior for queue policy.
type IntegrationOperationType string

const (
	// IntegrationOperationTypeSync is the default operation class.
	IntegrationOperationTypeSync IntegrationOperationType = "sync"
	// IntegrationOperationTypeImport captures import-style workloads.
	IntegrationOperationTypeImport IntegrationOperationType = "import"
	// IntegrationOperationTypeExport captures export-style workloads.
	IntegrationOperationTypeExport IntegrationOperationType = "export"
	// IntegrationOperationTypeWebhook captures webhook-driven workloads.
	IntegrationOperationTypeWebhook IntegrationOperationType = "webhook"
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
	// Type classifies operation behavior for policy routing.
	Type IntegrationOperationType `json:"type"`
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

var integrationOperationPolicies = map[IntegrationOperationType]integrationOperationPolicy{
	IntegrationOperationTypeSync: {
		timeoutSeconds: integrationSyncTimeoutSeconds,
		maxAttempts:    integrationSyncMaxAttempts,
	},
	IntegrationOperationTypeImport: {
		timeoutSeconds: integrationLongRunningTimeoutSeconds,
		maxAttempts:    integrationLongRunningMaxAttempts,
	},
	IntegrationOperationTypeExport: {
		timeoutSeconds: integrationLongRunningTimeoutSeconds,
		maxAttempts:    integrationLongRunningMaxAttempts,
	},
	IntegrationOperationTypeWebhook: {
		timeoutSeconds: integrationWebhookTimeoutSeconds,
		maxAttempts:    integrationWebhookMaxAttempts,
	},
}

// NewIntegrationOperationEnvelope derives queue and retry policy for a requested integration operation.
func NewIntegrationOperationEnvelope(payload IntegrationOperationRequestedPayload) IntegrationOperationEnvelope {
	operationType := classifyIntegrationOperationType(payload.Operation)
	policy, ok := integrationOperationPolicies[operationType]
	if !ok {
		policy = integrationOperationPolicies[IntegrationOperationTypeSync]
	}

	return IntegrationOperationEnvelope{
		Request:        payload,
		Type:           operationType,
		TimeoutSeconds: policy.timeoutSeconds,
		MaxAttempts:    policy.maxAttempts,
		IdempotencyKey: integrationOperationIdempotencyKey(payload),
	}
}

// Headers converts envelope policy into Gala dispatch headers.
func (e IntegrationOperationEnvelope) Headers() gala.Headers {
	return gala.Headers{
		IdempotencyKey: strings.TrimSpace(e.IdempotencyKey),
		Queue:          IntegrationQueueName,
		MaxAttempts:    e.MaxAttempts,
	}
}

func classifyIntegrationOperationType(operation string) IntegrationOperationType {
	normalized := strings.ToLower(strings.TrimSpace(operation))

	switch {
	case strings.Contains(normalized, "webhook"):
		return IntegrationOperationTypeWebhook
	case strings.Contains(normalized, "export"):
		return IntegrationOperationTypeExport
	case strings.Contains(normalized, "collect"), strings.Contains(normalized, "import"):
		return IntegrationOperationTypeImport
	default:
		return IntegrationOperationTypeSync
	}
}

func integrationOperationIdempotencyKey(payload IntegrationOperationRequestedPayload) string {
	runID := strings.TrimSpace(payload.RunID)
	orgID := strings.TrimSpace(payload.OrgID)
	provider := strings.TrimSpace(payload.Provider)
	operation := strings.TrimSpace(payload.Operation)

	if runID != "" {
		return fmt.Sprintf("%s:%s:%s:%s", provider, operation, orgID, runID)
	}

	return fmt.Sprintf("%s:%s:%s", provider, operation, orgID)
}

// IntegrationOperationRequestedTopic is emitted when an integration operation is queued.
var IntegrationOperationRequestedTopic = gala.Topic[IntegrationOperationEnvelope]{
	Name: gala.TopicName(TopicIntegrationOperationRequested),
}
