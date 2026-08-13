package engine

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/pkg/jsonx"
)

// TriggerInput captures the trigger metadata passed to workflow execution
type TriggerInput struct {
	// EventType is the trigger event name
	EventType string
	// ChangedFields lists updated fields on the target object
	ChangedFields []string
	// ChangedEdges lists updated edges on the target object
	ChangedEdges []string
	// AddedIDs captures added edge IDs keyed by edge name
	AddedIDs map[string][]string
	// RemovedIDs captures removed edge IDs keyed by edge name
	RemovedIDs map[string][]string
	// ProposedChanges contains proposed field updates for approval workflows
	ProposedChanges map[string]any
	// OldValues contains pre-mutation values for changed fields on single-row updates
	OldValues map[string]any
}

// ChangeSet returns the trigger mutation change-set from trigger input, marshaling the
// engine's map-shaped proposed changes into the delta contract's opaque JSON
func (input TriggerInput) ChangeSet() entityops.ChangeSet {
	proposed, err := jsonx.ToRawMessage(input.ProposedChanges)
	if err != nil || len(input.ProposedChanges) == 0 {
		proposed = nil
	}

	return entityops.ChangeSet{
		ChangedFields:   input.ChangedFields,
		ChangedEdges:    input.ChangedEdges,
		AddedIDs:        input.AddedIDs,
		RemovedIDs:      input.RemovedIDs,
		ProposedChanges: proposed,
		OldValues:       input.OldValues,
	}
}

// SetChangeSet applies a mutation change-set onto trigger input fields, decoding the
// opaque proposed-changes JSON into the engine's map shape
func (input *TriggerInput) SetChangeSet(changeSet entityops.ChangeSet) {
	if input == nil {
		return
	}

	cloned := changeSet.Clone()
	input.ChangedFields = cloned.ChangedFields
	input.ChangedEdges = cloned.ChangedEdges
	input.AddedIDs = cloned.AddedIDs
	input.RemovedIDs = cloned.RemovedIDs
	input.OldValues = cloned.OldValues

	input.ProposedChanges = cloned.ProposedMap()
}
