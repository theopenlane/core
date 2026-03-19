package types

import (
	"context"
	"encoding/json"

	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/gala"
)

// ExecutionPolicy controls synchronous execution behavior for one operation
type ExecutionPolicy struct {
	// Inline indicates the operation should execute synchronously for direct API callers
	Inline bool `json:"inline,omitempty"`
}

// IngestContract declares one ingest target emitted by an operation
type IngestContract struct {
	// Schema is the normalized target schema emitted by the operation
	Schema string `json:"schema"`
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

// IngestHandler executes one definition operation and returns typed ingest payload sets for pipeline routing
type IngestHandler func(ctx context.Context, request OperationRequest) ([]IngestPayloadSet, error)

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
	// Policy controls synchronous execution behavior for the operation
	Policy ExecutionPolicy `json:"policy,omitempty"`
	// Ingest declares the normalized schemas emitted by the operation
	Ingest []IngestContract `json:"ingest,omitempty"`
	// Handle executes the operation; set for operations that do not produce ingest payloads
	Handle OperationHandler `json:"-"`
	// IngestHandle executes the operation and returns typed payload sets for the ingest pipeline;
	// set for operations that produce ingest data — mutually exclusive with Handle
	IngestHandle IngestHandler `json:"-"`
}

const HealthDefaultOperation = "health.default"
