package operations

import (
	"encoding/json"

	"github.com/theopenlane/core/common/enums"
)

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
	// Force requests credential refresh when a future minting layer exists
	Force bool
	// ClientForce requests client cache bypass
	ClientForce bool
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
	// WorkflowMeta carries optional workflow linkage for workflow-triggered operations
	WorkflowMeta *WorkflowMeta `json:"workflowMeta,omitempty"`
}
