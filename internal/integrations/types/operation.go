package types

import (
	"context"
	"encoding/json"

	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/gala"
)

// OperationName identifies one executable operation within a definition
type OperationName string

// OperationKind classifies one operation's execution type
type OperationKind string

const (
	// OperationKindHealth identifies a health check operation
	OperationKindHealth OperationKind = "health"
	// OperationKindSync identifies a sync operation
	OperationKindSync OperationKind = "sync"
	// OperationKindCollect identifies a collection operation
	OperationKindCollect OperationKind = "collect"
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
	Schema MappingSchema `json:"schema"`
	// EnsurePayloads forces payload preservation before ingest runs
	EnsurePayloads bool `json:"ensurePayloads,omitempty"`
}

// OperationHandler executes one definition operation
type OperationHandler func(ctx context.Context, integration *generated.Integration, credential CredentialSet, client any, config json.RawMessage) (json.RawMessage, error)

// OperationRegistration declares one executable operation for a definition
type OperationRegistration struct {
	// Name is the stable operation identifier within the definition
	Name OperationName `json:"name"`
	// Kind classifies the operation execution type
	Kind OperationKind `json:"kind,omitempty"`
	// Description describes what the operation does
	Description string `json:"description,omitempty"`
	// Topic is the gala topic used to execute the operation
	Topic gala.TopicName `json:"topic"`
	// Client identifies which named client the operation uses
	Client ClientName `json:"client,omitempty"`
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
