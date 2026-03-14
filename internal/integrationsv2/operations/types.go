package operations

import (
	"encoding/json"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/integrationsv2/types"
)

// DispatchRequest describes one requested operation dispatch
type DispatchRequest struct {
	// InstallationID is the target installation identifier
	InstallationID string
	// Operation is the definition-local operation identifier
	Operation types.OperationName
	// Config is the operation configuration payload
	Config json.RawMessage
	// Force requests credential refresh when a future minting layer exists
	Force bool
	// ClientForce requests client cache bypass
	ClientForce bool
	// RunType is the integration run type recorded for the dispatch
	RunType enums.IntegrationRunType
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
	// Force requests credential refresh when a future minting layer exists
	Force bool `json:"force,omitempty"`
	// ClientForce requests client cache bypass
	ClientForce bool `json:"clientForce,omitempty"`
}
