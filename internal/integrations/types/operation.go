package types

import (
	"context"
	"encoding/json"
)

// OperationName identifies a provider operation (health check, findings harvest, etc.)
type OperationName string

const (
	// OperationHealthDefault identifies the default provider health check operation
	OperationHealthDefault OperationName = "health.default"
	// OperationVulnerabilitiesCollect identifies the vulnerabilities collection operation
	OperationVulnerabilitiesCollect OperationName = "vulnerabilities.collect"
	// OperationDirectorySync identifies the directory synchronization operation
	OperationDirectorySync OperationName = "directory.sync"
)

// OperationKind categorizes an operation for routing and telemetry
type OperationKind string

const (
	// OperationKindHealth represents a health check operation
	OperationKindHealth OperationKind = "health_check"
	// OperationKindCollectFindings represents a findings collection operation
	OperationKindCollectFindings OperationKind = "collect_findings"
	// OperationKindScanSettings represents a settings scan operation
	OperationKindScanSettings OperationKind = "scan_settings"
	// OperationKindNotify represents a notification operation
	OperationKindNotify OperationKind = "notify"
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

// OperationFunc executes a provider operation using stored credentials and optional clients
type OperationFunc func(ctx context.Context, input OperationInput) (OperationResult, error)

// IngestContract declares one ingest target schema emitted by an operation;
// the Fn field carries the provider's ingest implementation for this schema —
// there is no centralized ingest package, each provider registers its own
// IngestFn per contract
type IngestContract struct {
	// Schema identifies the normalized ingest schema (e.g. Vulnerability)
	Schema MappingSchema
	// EnsurePayloads forces include_payloads=true prior to operation execution
	EnsurePayloads bool
	// Fn is the provider's ingest implementation for this schema
	Fn IngestFunc
}

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
	ConfigSchema json.RawMessage
	// OutputSchema defines the JSON schema for operation output
	OutputSchema json.RawMessage
	// Ingest declares one or more optional ingest contracts emitted by this operation
	Ingest []IngestContract
}

// OperationInput carries the runtime information supplied to operation handlers
type OperationInput struct {
	// OrgID identifies the organization executing the operation
	OrgID string
	// Provider identifies the provider for this operation
	Provider ProviderType
	// Credential contains the credential fields for authentication
	Credential CredentialSet
	// Client is the provider-specific client instance wrapper
	Client ClientInstance
	// Config contains operation-specific configuration as a JSON object document
	Config json.RawMessage
}

// OperationResult reports the outcome of an operation handler
type OperationResult struct {
	// Status indicates whether the operation succeeded or failed
	Status OperationStatus
	// Summary provides a human-readable summary of the result
	Summary string
	// Details contains structured result data as a JSON object document
	Details json.RawMessage
}

// OperationRequest contains the parameters required to invoke an operation
type OperationRequest struct {
	// OrgID identifies the organization requesting the operation
	OrgID string
	// IntegrationID identifies a specific installed integration to use for credential lookup
	IntegrationID string
	// Provider identifies which provider to use
	Provider ProviderType
	// Name identifies which operation to execute
	Name OperationName
	// Config contains operation-specific configuration as a JSON object document
	Config json.RawMessage
	// Force bypasses cached credentials and forces a credential refresh
	Force bool
	// ClientForce forces creation of a new client instance bypassing the client pool cache
	ClientForce bool
}

// AlertEnvelope wraps an alert payload emitted by integration webhooks
type AlertEnvelope struct {
	// AlertType identifies the alert category (dependabot, code_scanning, etc.)
	AlertType string `json:"alertType"`
	// Resource identifies the alert resource (repo, project, etc.)
	Resource string `json:"resource,omitempty"`
	// Action indicates the webhook action (created, resolved, etc.)
	Action string `json:"action,omitempty"`
	// Payload is the raw alert payload as received from the provider
	Payload json.RawMessage `json:"payload,omitempty"`
}

// OperationTemplate is a unified replacement for both common/openapi.IntegrationOperationTemplate
// and any operations-level OperationTemplate type; it captures persisted configuration for an operation
type OperationTemplate struct {
	// Config holds the operation configuration
	Config json.RawMessage `json:"config,omitempty"`
	// AllowOverrides lists which config fields can be overridden at runtime
	AllowOverrides []string `json:"allowOverrides,omitempty"`
}
