package operations

import (
	"encoding/json"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/pkg/gala"
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

// WebhookEnvelope is the durable payload emitted for one inbound integration webhook event
type WebhookEnvelope struct {
	gala.OperationContext
	// Payload is the raw webhook request body
	Payload json.RawMessage `json:"payload"`
	// Headers contains the inbound HTTP request headers
	Headers map[string]string `json:"headers,omitempty"`
}

// Envelope is the payload emitted to the operation topic
type Envelope struct {
	gala.OperationContext
	// Config is the operation configuration payload
	Config json.RawMessage `json:"config,omitempty"`
	// ForceClientRebuild requests client cache bypass
	ForceClientRebuild bool `json:"forceClientRebuild,omitempty"`
}
