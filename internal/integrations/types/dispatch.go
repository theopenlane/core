package types //nolint:revive

import (
	"context"
	"encoding/json"
	"time"

	"github.com/theopenlane/core/common/enums"
)

// DispatchRequest describes one requested operation dispatch
type DispatchRequest struct {
	// IntegrationID is the target integration identifier; required on the customer path, empty on the runtime path
	IntegrationID string
	// DefinitionID is the definition identifier carried as metadata on the runtime path
	DefinitionID string
	// OwnerID is the owning organization carried on the runtime path; derived from the DB record for customer dispatch
	OwnerID string
	// Operation is the definition-local operation identifier
	Operation string
	// Config is the operation configuration payload
	Config json.RawMessage
	// ForceClientRebuild requests client cache bypass
	ForceClientRebuild bool
	// RunType is the integration run type recorded for the dispatch
	RunType enums.IntegrationRunType
	// Workflow carries optional workflow linkage for workflow-triggered operations
	Workflow *WorkflowMeta
	// ScheduledAt defers execution until the specified time; nil means immediate
	ScheduledAt *time.Time
	// Runtime signals that this dispatch should use the runtime provider path
	Runtime bool
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

// DispatchFunc enqueues one integration operation for execution through the runtime-managed dispatcher
type DispatchFunc func(ctx context.Context, req DispatchRequest) (DispatchResult, error)
