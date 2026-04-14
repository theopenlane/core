package types //nolint:revive

import (
	"context"
	"encoding/json"
	"time"

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
	// LastRunAt is the finish time of the most recent successful run for this operation,
	// used by handlers that support incremental/delta fetches
	LastRunAt *time.Time
	// DB is the ent client for operations that need database access
	DB *generated.Client
}

// OperationHandler executes one definition operation
type OperationHandler func(ctx context.Context, request OperationRequest) (json.RawMessage, error)

// IngestHandler executes one definition operation and returns typed ingest payload sets for pipeline routing
type IngestHandler func(ctx context.Context, request OperationRequest) ([]IngestPayloadSet, error)

// DispatchForOwnerFunc resolves an integration for the given owner and dispatches
// an operation. When no customer installation exists but the definition is
// runtime-provisioned, execution falls back to the runtime path
type DispatchForOwnerFunc func(ctx context.Context, definitionID string, operationName string, ownerID string, config json.RawMessage) error

// OperationRegistration declares one executable operation for a definition
type OperationRegistration struct {
	// Name is the stable operation identifier within the definition
	Name string `json:"name"`
	// Description describes what the operation does
	Description string `json:"description,omitempty"`
	// RequiredPermissions lists what scopes or permissions are needed to retrieve data for the Operation
	RequiredPermissions []string `json:"requiredPermissions,omitempty"`
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
	// IngestHandle executes the operation and returns typed payload sets for the ingest pipeline,
	// set for operations that produce ingest data and mutually exclusive with Handle
	IngestHandle IngestHandler `json:"-"`
	// DisabledForAll indicates if the sync is not currently available for use and no config params are shown to the user
	DisabledForAll bool `json:"disabledForAll"`
	// Disabled reports whether this operation is disabled for a given installation's user input JSON;
	// when set, reconcile cycles are skipped entirely instead of running and returning empty results
	Disabled func(userInput json.RawMessage) bool `json:"-"`
	// ConfigResolver extracts the operation-specific config JSON from the installation's user input JSON;
	// when set, the resolved config is used as the operation config for reconcile runs and as the
	// source for per-operation filter expressions in the ingest pipeline
	ConfigResolver func(userInput json.RawMessage) json.RawMessage `json:"-"`
	// ReconcileSchedule overrides the default adaptive schedule for this operation's reconcile cycles;
	// useful for operations that always do a full fetch and should run less frequently
	ReconcileSchedule *gala.Schedule `json:"-"`
	// SkipDefaultLookback disables the runtime's default lookback window on initial runs;
	// when true, LastRunAt is nil on first run so the handler performs a full fetch
	SkipDefaultLookback bool `json:"-"`
}
