package types

import (
	"context"
)

// OperationName identifies a provider operation (health check, findings harvest, etc).
type OperationName string

// OperationKind categorizes an operation for routing/telemetry.
type OperationKind string

const (
	OperationKindHealth          OperationKind = "health_check"
	OperationKindCollectFindings OperationKind = "collect_findings"
	OperationKindScanSettings    OperationKind = "scan_settings"
)

// OperationStatus communicates the result of an operation run.
type OperationStatus string

const (
	OperationStatusUnknown OperationStatus = "unknown"
	OperationStatusOK      OperationStatus = "ok"
	OperationStatusFailed  OperationStatus = "failed"
)

// OperationDescriptor describes a provider-published operation handler.
type OperationDescriptor struct {
	Provider     ProviderType
	Name         OperationName
	Kind         OperationKind
	Description  string
	Client       ClientName
	Run          OperationFunc
	ConfigSchema map[string]any
	OutputSchema map[string]any
}

// OperationInput carries the runtime information supplied to operation handlers.
type OperationInput struct {
	OrgID      string
	Provider   ProviderType
	Credential CredentialPayload
	Client     any
	Config     map[string]any
}

// OperationResult reports the outcome of an operation handler.
type OperationResult struct {
	Status  OperationStatus
	Summary string
	Details map[string]any
}

// OperationFunc executes a provider operation using stored credentials and optional clients.
type OperationFunc func(ctx context.Context, input OperationInput) (OperationResult, error)

// OperationRequest contains the parameters required to invoke an operation.
type OperationRequest struct {
	OrgID       string
	Provider    ProviderType
	Name        OperationName
	Config      map[string]any
	Force       bool
	ClientForce bool
}

// OperationProvider is implemented by providers that publish runtime operations.
type OperationProvider interface {
	Provider
	Operations() []OperationDescriptor
}
