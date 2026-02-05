package soiree

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
}

const (
	// TopicIntegrationOperationRequested is emitted when an integration operation is queued
	TopicIntegrationOperationRequested = "integration.operation.requested"
)

// IntegrationOperationRequestedTopic is emitted when an integration operation is queued
var IntegrationOperationRequestedTopic = NewTypedTopic(TopicIntegrationOperationRequested,
	WithObservability(ObservabilitySpec[IntegrationOperationRequestedPayload]{
		Operation: "handle_integration_operation_requested",
		Origin:    "listeners",
	}),
)
