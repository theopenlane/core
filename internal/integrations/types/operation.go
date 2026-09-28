package types //nolint:revive

import (
	"context"
	"encoding/json"
	"time"

	"github.com/theopenlane/core/common/enums"
	generated "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// Switch is the embeddable per-installation disable toggle carried by an operation config type
type Switch struct {
	// Disable switches the operation off for the installation
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable this sync"`
}

// Disabled reports whether the toggle switches the operation off
func (s Switch) Disabled() bool {
	return s.Disable
}

// Switchable is implemented by operation config types that carry a per-installation disable toggle
type Switchable interface {
	// Disabled reports whether the config switches the operation off
	Disabled() bool
}

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

// RateLimitPolicy bounds how often one operation may run per calling organization within a rolling window, enforced identically on every execution path
type RateLimitPolicy struct {
	// Window is the rolling window duration for the operation's execution budget
	Window time.Duration
	// Limit is the number of executions allowed per window; values below one default to a single execution
	Limit int
}

// ExecutionPolicy controls synchronous execution behavior for one operation
type ExecutionPolicy struct {
	// Inline indicates the operation should execute synchronously for direct API callers
	Inline bool `json:"inline,omitempty"`
	// Reconcile indicates the operation should be dispatched on a recurring schedule per connected installation
	Reconcile bool `json:"reconcile,omitempty"`
	// Scheduled indicates the operation runs on a recurring schedule through the runtime provider path, with no installation; used for system-level sweeps
	Scheduled bool `json:"scheduled,omitempty"`
	// SkipRunRecord indicates the IntegrationRun record creation should be skipped
	SkipRunRecord bool `json:"skipRunRecord,omitempty"`
	// Snapshot marks a full-snapshot sync that owns a directory sync run and removal inference
	Snapshot bool `json:"snapshot,omitempty"`
}

// ScheduledCycleResult is the conventional response payload for scheduled runtime operations, carrying the cycle delta used for adaptive scheduling
type ScheduledCycleResult struct {
	// Processed is the number of records handled during the cycle
	Processed int `json:"processed"`
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
	// LastRunAt is the finish time of the most recent successful run for this operation
	LastRunAt *time.Time
	// DB is the ent client for operations that need database access
	DB *generated.Client
	// Dispatch enqueues other integration operations through the runtime-managed dispatcher
	Dispatch DispatchFunc
	// Services exposes the full runtime service surface
	Services RuntimeServices
}

// OperationHandler executes one definition operation
type OperationHandler func(ctx context.Context, request OperationRequest) (json.RawMessage, error)

// IngestHandler executes one definition operation and returns typed ingest payload sets for pipeline routing
type IngestHandler func(ctx context.Context, request OperationRequest) ([]IngestPayloadSet, error)

// OperationRegistration declares one executable operation for a definition
type OperationRegistration struct {
	// Name is the stable operation identifier within the definition
	Name string `json:"name"`
	// Replaces lists the retired operation names whose recorded runs and health move onto this operation
	Replaces []string `json:"-"`
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
	// UISchema is optional UI layout hints for the input form; nil when absent
	UISchema json.RawMessage `json:"uiSchema,omitempty"`
	// CustomerSelectable controls whether the operation is exposed in customer-facing surfaces
	CustomerSelectable *bool `json:"customerSelectable,omitempty"`
	// Internal marks the operation as reachable only through its own listener or saga machinery
	Internal bool `json:"-"`
	// RequiresPaymentMethod gates direct invocation through RunIntegrationOperation on the calling organization having a payment method on file
	RequiresPaymentMethod bool `json:"-"`
	// Policy controls synchronous execution behavior for the operation
	Policy ExecutionPolicy `json:"policy"`
	// RateLimit bounds how often this operation may run per organization; nil (default) means unlimited
	RateLimit *RateLimitPolicy `json:"-"`
	// Ingest declares the normalized schemas emitted by the operation
	Ingest []IngestContract `json:"ingest,omitempty"`
	// HealthCheck probes this operation's prerequisites under its own client
	HealthCheck OperationHandler `json:"-"`
	// Handle executes the operation; set for operations that do not produce ingest payloads
	Handle OperationHandler `json:"-"`
	// IngestHandle executes the operation and returns typed payload sets for the ingest pipeline
	IngestHandle IngestHandler `json:"-"`
	// DisabledForAll indicates if the sync is not currently available for use and no config params are shown to the user
	DisabledForAll bool `json:"disabledForAll"`
	// Disabled reports whether this operation is disabled for a given installation's user input JSON
	Disabled func(userInput json.RawMessage) bool `json:"-"`
	// ConfigResolver extracts the operation-specific config JSON from the installation's user input JSON
	ConfigResolver func(userInput json.RawMessage) json.RawMessage `json:"-"`
	// ConfigDisabled reports whether the operation's own config section switches it off
	ConfigDisabled func(config json.RawMessage) bool `json:"-"`
	// Schedule overrides the default adaptive schedule for this operation's recurring reconcile or scheduled cycles
	Schedule *gala.Schedule `json:"-"`
	// SkipDefaultLookback disables the runtime's default lookback window on initial runs
	SkipDefaultLookback bool `json:"-"`
}

// DisabledFor reports whether the operation is switched off for everyone or for the installation whose user input is given
func (o OperationRegistration) DisabledFor(userInput json.RawMessage) bool {
	return o.DisabledForAll || (o.Disabled != nil && o.Disabled(userInput))
}
