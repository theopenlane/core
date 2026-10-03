package types //nolint:revive

import (
	"context"
	"encoding/json"
	"time"

	"github.com/theopenlane/core/common/enums"
	generated "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// OperationSettings holds the uniform per-installation settings every stored operation input carries; a
// stored-input config type embeds it so the settings reflect into the operation's schema beside its own fields
type OperationSettings struct {
	// Disable switches the operation off for the installation
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable this operation for the installation"`
	// FilterExpr limits ingested records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to records before ingesting"`
}

// operationSettings marks the type as carrying the uniform operation settings
func (OperationSettings) operationSettings() {}

// OperationInput is satisfied by any config type that embeds OperationSettings, which every stored-input operation requires
type OperationInput interface {
	operationSettings()
}

// operationSettingsSchema is the reflected schema of the uniform operation settings
var operationSettingsSchema = jsonx.SchemaFrom[OperationSettings]()

// OperationSettingsSchema returns a copy of the reflected uniform operation settings schema
func OperationSettingsSchema() json.RawMessage {
	return jsonx.CloneRawMessage(operationSettingsSchema)
}

// OperationSettingsFrom decodes the uniform settings from a stored operation input document
func OperationSettingsFrom(doc json.RawMessage) (OperationSettings, error) {
	var settings OperationSettings

	err := jsonx.UnmarshalIfPresent(doc, &settings)

	return settings, err
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

// RateLimitPolicy bounds how often an operation may run per organization within a rolling window
type RateLimitPolicy struct {
	// Window is the rolling window duration for the operation's execution budget
	Window time.Duration
	// Limit is executions allowed per window; below one defaults to a single execution
	Limit int
}

// ExecutionPolicy controls synchronous execution behavior for one operation
type ExecutionPolicy struct {
	// Inline indicates the operation should execute synchronously for direct API callers
	Inline bool `json:"inline,omitempty"`
	// Reconcile dispatches the operation on a recurring schedule per connected installation
	Reconcile bool `json:"reconcile,omitempty"`
	// Scheduled runs the operation on a recurring schedule with no installation, for system sweeps
	Scheduled bool `json:"scheduled,omitempty"`
	// SkipRunRecord indicates the IntegrationRun record creation should be skipped
	SkipRunRecord bool `json:"skipRunRecord,omitempty"`
	// Snapshot marks a full-snapshot sync that owns a directory sync run and removal inference
	Snapshot bool `json:"snapshot,omitempty"`
}

// ScheduledCycleResult is the response payload scheduled operations use for adaptive scheduling
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

// IngestHandler executes an operation and returns typed payload sets for the ingest pipeline
type IngestHandler func(ctx context.Context, request OperationRequest) ([]IngestPayloadSet, error)

// OperationRegistration declares one executable operation for a definition
type OperationRegistration struct {
	// Name is the stable operation identifier within the definition
	Name string `json:"name"`
	// Replaces lists retired operation names whose runs and health move onto this operation
	Replaces []string `json:"-"`
	// Description describes what the operation does
	Description string `json:"description,omitempty"`
	// RequiredPermissions lists scopes or permissions needed to retrieve data for the operation
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
	// RequiresPaymentMethod gates direct invocation on the org having a payment method on file
	RequiresPaymentMethod bool `json:"-"`
	// Policy controls synchronous execution behavior for the operation
	Policy ExecutionPolicy `json:"policy"`
	// RateLimit bounds how often this operation may run per organization; nil means unlimited
	RateLimit *RateLimitPolicy `json:"-"`
	// Ingest declares the normalized schemas emitted by the operation
	Ingest []IngestContract `json:"ingest,omitempty"`
	// HealthCheck probes this operation's prerequisites under its own client
	HealthCheck OperationHandler `json:"-"`
	// Handle executes the operation; set for operations that do not produce ingest payloads
	Handle OperationHandler `json:"-"`
	// IngestHandle executes the operation and returns typed payload sets for the ingest pipeline
	IngestHandle IngestHandler `json:"-"`
	// DisabledForAll marks the sync unavailable, hiding config params from the user
	DisabledForAll bool `json:"disabledForAll"`
	// Input describes the stored per-installation input document for the operation
	Input *InputRegistration `json:"input,omitempty"`
	// Schedule overrides the default adaptive schedule for reconcile or scheduled cycles
	Schedule *gala.Schedule `json:"-"`
	// SkipDefaultLookback disables the runtime's default lookback window on initial runs
	SkipDefaultLookback bool `json:"-"`
}

// DisabledFor reports whether the operation is switched off globally or by the stored input document
func (o OperationRegistration) DisabledFor(input json.RawMessage) bool {
	settings, err := OperationSettingsFrom(input)

	return o.DisabledForAll || (err == nil && settings.Disable)
}
