package operations

import (
	"encoding/json"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/pkg/gala"
)

// IngestContext holds the stable per-integration dependencies shared across all ingest call paths
type IngestContext struct {
	// Registry is the integration definition registry used to resolve mappings and definitions
	Registry *registry.Registry
	// DB is the ent client used for persistence
	DB *ent.Client
	// Runtime is the Gala instance used for async emit; nil on the synchronous persist path
	Runtime *gala.Gala
	// Installation is the integration record being ingested into
	Installation *ent.Integration
}

// WorkflowMeta carries optional workflow linkage for workflow-triggered operations
type WorkflowMeta struct {
	// InstanceID identifies the workflow instance that triggered the operation
	InstanceID string `json:"instanceId"`
	// ActionKey identifies the workflow action key
	ActionKey string `json:"actionKey"`
	// ActionIndex captures the workflow action index
	ActionIndex int `json:"actionIndex"`
	// ObjectID identifies the workflow object
	ObjectID string `json:"objectId"`
	// ObjectType identifies the workflow object type
	ObjectType enums.WorkflowObjectType `json:"objectType,omitempty"`
}

// DispatchRequest describes one requested operation dispatch
type DispatchRequest struct {
	// InstallationID is the target installation identifier
	InstallationID string
	// Operation is the definition-local operation identifier
	Operation string
	// Config is the operation configuration payload
	Config json.RawMessage
	// ForceClientRebuild requests client cache bypass
	ForceClientRebuild bool
	// RunType is the integration run type recorded for the dispatch
	RunType enums.IntegrationRunType
	// WorkflowMeta carries optional workflow linkage for workflow-triggered operations
	WorkflowMeta *WorkflowMeta
}

// DispatchResult captures the queued run metadata
type DispatchResult struct {
	// RunID is the persisted run identifier
	RunID string
	// EventID is the emitted Gala event identifier
	EventID string
	// Status is the run status at dispatch time
	Status enums.IntegrationRunStatus
}

// WebhookEnvelope is the durable payload emitted for one inbound integration webhook event
type WebhookEnvelope struct {
	// IntegrationID is the installation identifier that received the webhook
	IntegrationID string `json:"integrationId"`
	// DefinitionID is the integration definition identifier
	DefinitionID string `json:"definitionId"`
	// Webhook is the webhook name within the definition
	Webhook string `json:"webhook"`
	// Event is the normalized event name within the webhook
	Event string `json:"event"`
	// DeliveryID is the provider-assigned delivery identifier for deduplication
	DeliveryID string `json:"deliveryId,omitempty"`
	// Payload is the raw webhook request body
	Payload json.RawMessage `json:"payload"`
	// Headers contains the inbound HTTP request headers
	Headers map[string]string `json:"headers,omitempty"`
}

// Envelope is the payload emitted to the operation topic
type Envelope struct {
	// RunID is the persisted run identifier
	RunID string `json:"runId"`
	// InstallationID is the target installation identifier
	InstallationID string `json:"installationId"`
	// DefinitionID is the target definition identifier
	DefinitionID string `json:"definitionId"`
	// Operation is the definition-local operation identifier
	Operation string `json:"operation"`
	// Config is the operation configuration payload
	Config json.RawMessage `json:"config,omitempty"`
	// ForceClientRebuild requests client cache bypass
	ForceClientRebuild bool `json:"forceClientRebuild,omitempty"`
	// WorkflowMeta carries optional workflow linkage for workflow-triggered operations
	WorkflowMeta *WorkflowMeta `json:"workflowMeta,omitempty"`
}
