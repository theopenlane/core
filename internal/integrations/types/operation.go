package types //nolint:revive

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/common/enums"
	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/gala"
)

// WorkflowMeta captures workflow linkage for a queued integration execution
type WorkflowMeta struct {
	// InstanceID identifies the workflow instance that queued the execution
	InstanceID string `json:"instanceId,omitempty"`
	// ActionKey identifies the workflow action key
	ActionKey string `json:"actionKey,omitempty"`
	// ActionIndex captures the zero-based action index
	ActionIndex int `json:"actionIndex,omitempty"`
	// ObjectID identifies the workflow object
	ObjectID string `json:"objectId,omitempty"`
	// ObjectType identifies the workflow object type
	ObjectType enums.WorkflowObjectType `json:"objectType,omitempty"`
}

// ExecutionPolicy controls synchronous execution behavior for one operation
type ExecutionPolicy struct {
	// Inline indicates the operation should execute synchronously for direct API callers
	Inline bool `json:"inline,omitempty"`
	// Reconcile indicates the operation should be dispatched on a recurring schedule
	Reconcile bool `json:"reconcile,omitempty"`
	// SkipRunRecord indicates the IntegrationRun record creation should be skipped
	SkipRunRecord bool `json:"skipRunRecord,omitempty"`
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
	// Credentials lists all resolved credential bundles for the operation by slot ref
	Credentials CredentialBindings
	// Client is the built client instance for this operation when one is registered
	Client any
	// Config is the operation-specific configuration payload
	Config json.RawMessage
	// DB is the ent client for operations that need database access
	DB *generated.Client
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
	// UISchema is optional UI layout hints for the input form; nil when absent
	UISchema json.RawMessage `json:"uiSchema,omitempty"`
	// CustomerSelectable reports whether the operation is exposed in customer-facing pickers;
	// false by default, true for operations that customers may pick from a catalog
	CustomerSelectable bool `json:"customerSelectable,omitempty"`
	// Policy controls synchronous execution behavior for the operation
	Policy ExecutionPolicy `json:"policy"`
	// Ingest declares the normalized schemas emitted by the operation
	Ingest []IngestContract `json:"ingest,omitempty"`
	// Handle executes the operation; set for operations that do not produce ingest payloads
	Handle OperationHandler `json:"-"`
	// IngestHandle executes the operation and returns typed payload sets for the ingest pipeline,
	// set for operations that produce ingest data and mutually exclusive with Handle
	IngestHandle IngestHandler `json:"-"`
}
