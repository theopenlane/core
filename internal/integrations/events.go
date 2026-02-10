package integrations

import (
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/pkg/events/soiree"
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
	TopicIntegrationOperationRequested = "integration.operation.requested"
)

// IntegrationOperationRequestedTopic is emitted when an integration operation is queued
var IntegrationOperationRequestedTopic = soiree.NewTypedTopic(TopicIntegrationOperationRequested,
	soiree.WithObservability(soiree.ObservabilitySpec[IntegrationOperationRequestedPayload]{
		Operation: "handle_integration_operation_requested",
		Origin:    "listeners",
	}),
)
