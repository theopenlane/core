package types

import (
	"context"
	"encoding/json"

	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/gala"
)

// ExecutionPolicy controls retry and idempotency behavior for one operation
type ExecutionPolicy struct {
	// MaxRetries is the maximum number of retry attempts for a failed operation
	MaxRetries uint `json:"maxRetries,omitempty"`
	// Idempotent indicates the operation can be safely retried without side effects
	Idempotent bool `json:"idempotent,omitempty"`
}

// IngestContract declares one ingest target emitted by an operation
type IngestContract struct {
	// Schema is the normalized target schema emitted by the operation
	Schema string `json:"schema"`
	// EnsurePayloads forces payload preservation before ingest runs
	EnsurePayloads bool `json:"ensurePayloads,omitempty"`
}

// OperationRequest bundles the inputs for executing one definition operation
type OperationRequest struct {
	// Integration is the target installation record
	Integration *generated.Integration
	// Credential is the installation-scoped credential bundle
	Credential CredentialSet
	// Client is the built client instance for this operation when one is registered
	Client any
	// Config is the operation-specific configuration payload
	Config json.RawMessage
}

// OperationHandler executes one definition operation
type OperationHandler func(ctx context.Context, request OperationRequest) (json.RawMessage, error)

// OperationRegistration declares one executable operation for a definition
type OperationRegistration struct {
	// Name is the stable operation identifier within the definition
	Name string `json:"name"`
	// Description describes what the operation does
	Description string `json:"description,omitempty"`
	// Topic is the gala topic used to execute the operation
	Topic gala.TopicName `json:"topic"`
	// ClientRef identifies which registered client the operation uses
	ClientRef ClientID `json:"-"`
	// ConfigSchema is the JSON schema for operation configuration
	ConfigSchema json.RawMessage `json:"configSchema,omitempty"`
	// OutputSchema is the JSON schema for operation output
	OutputSchema json.RawMessage `json:"outputSchema,omitempty"`
	// Policy controls retry and idempotency behavior for the operation
	Policy ExecutionPolicy `json:"policy,omitempty"`
	// Ingest declares the normalized schemas emitted by the operation
	Ingest []IngestContract `json:"ingest,omitempty"`
	// Handle executes the operation
	Handle OperationHandler `json:"-"`
}
