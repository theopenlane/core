package operations

import (
	"encoding/json"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
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
	// Integration is the integration record being ingested into
	Integration *ent.Integration
}

// DispatchRequest describes one requested operation dispatch
type DispatchRequest struct {
	// IntegrationID is the target installation identifier
	IntegrationID string
	// Operation is the definition-local operation identifier
	Operation string
	// Config is the operation configuration payload
	Config json.RawMessage
	// ForceClientRebuild requests client cache bypass
	ForceClientRebuild bool
	// RunType is the integration run type recorded for the dispatch
	RunType enums.IntegrationRunType
	// Workflow carries optional workflow linkage for workflow-triggered operations
	Workflow *types.WorkflowMeta
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
	types.ExecutionMetadata
	// Payload is the raw webhook request body
	Payload json.RawMessage `json:"payload"`
	// Headers contains the inbound HTTP request headers
	Headers map[string]string `json:"headers,omitempty"`
}

// Envelope is the payload emitted to the operation topic
type Envelope struct {
	types.ExecutionMetadata
	// Config is the operation configuration payload
	Config json.RawMessage `json:"config,omitempty"`
	// ForceClientRebuild requests client cache bypass
	ForceClientRebuild bool `json:"forceClientRebuild,omitempty"`
}
