package operations

import (
	"encoding/json"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

const QuestionnaireTransformOperationName = "questionnaire.transform.assessment"

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
