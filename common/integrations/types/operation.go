package types //nolint:revive

import (
	"context"
	"encoding/json"
)

// OperationName identifies a provider operation (health check, findings harvest, etc).
type OperationName string

// OperationKind categorizes an operation for routing/telemetry
type OperationKind string

const (
	// OperationKindHealth represents a health check operation
	OperationKindHealth OperationKind = "health_check"
	// OperationKindCollectFindings represents a findings collection operation
	OperationKindCollectFindings OperationKind = "collect_findings"
	// OperationKindScanSettings represents a settings scan operation
	OperationKindScanSettings OperationKind = "scan_settings"
)

// OperationStatus communicates the result of an operation run
type OperationStatus string

const (
	// OperationStatusUnknown represents an unknown operation status
	OperationStatusUnknown OperationStatus = "unknown"
	// OperationStatusOK represents a successful operation
	OperationStatusOK OperationStatus = "ok"
	// OperationStatusFailed represents a failed operation
	OperationStatusFailed OperationStatus = "failed"
)

// OperationDescriptor describes a provider-published operation handler
type OperationDescriptor struct {
	// Provider identifies the provider offering this operation
	Provider ProviderType
	// Name is the unique operation identifier
	Name OperationName
	// Kind categorizes the operation type
	Kind OperationKind
	// Description explains what the operation does
	Description string
	// Client specifies which client type is required
	Client ClientName
	// Run is the function that executes the operation
	Run OperationFunc
	// ConfigSchema defines the JSON schema for operation configuration
	ConfigSchema map[string]any
	// OutputSchema defines the JSON schema for operation output
	OutputSchema map[string]any
}

// OperationInput carries the runtime information supplied to operation handlers
type OperationInput struct {
	// OrgID identifies the organization executing the operation
	OrgID string
	// Provider identifies the provider for this operation
	Provider ProviderType
	// Credential contains the credential payload for authentication
	Credential CredentialPayload
	// Client is the provider-specific client instance
	Client any
	// Config contains operation-specific configuration
	Config map[string]any
}

// OperationResult reports the outcome of an operation handler
type OperationResult struct {
	// Status indicates whether the operation succeeded or failed
	Status OperationStatus
	// Summary provides a human-readable summary of the result
	Summary string
	// Details contains structured result data
	Details map[string]any
}

// AlertEnvelope wraps an alert payload emitted by integration webhooks.
type AlertEnvelope struct {
	// AlertType identifies the alert category (dependabot, code_scanning, etc).
	AlertType string `json:"alertType"`
	// Resource identifies the alert resource (repo, project, etc).
	Resource string `json:"resource,omitempty"`
	// Action indicates the webhook action (created, resolved, etc).
	Action string `json:"action,omitempty"`
	// Payload is the raw alert payload as received from the provider.
	Payload json.RawMessage `json:"payload,omitempty"`
}

// OperationFunc executes a provider operation using stored credentials and optional clients
type OperationFunc func(ctx context.Context, input OperationInput) (OperationResult, error)

// OperationRequest contains the parameters required to invoke an operation
type OperationRequest struct {
	// OrgID identifies the organization requesting the operation
	OrgID string
	// Provider identifies which provider to use
	Provider ProviderType
	// Name identifies which operation to execute
	Name OperationName
	// Config contains operation-specific configuration
	Config map[string]any
	// Force bypasses cached operation results
	Force bool
	// ClientForce forces creation of a new client instance
	ClientForce bool
}

// OperationProvider is implemented by providers that publish runtime operations
type OperationProvider interface {
	Provider
	// Operations returns the list of operations offered by the provider
	Operations() []OperationDescriptor
}
