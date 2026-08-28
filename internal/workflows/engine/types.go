package engine

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
)

// TriggerInput captures the trigger metadata passed to workflow execution
type TriggerInput struct {
	// EventType is the trigger event name
	EventType string
	// ChangedFields lists updated fields on the target object
	ChangedFields []string
	// ClearedFields lists fields explicitly cleared on the target object
	ClearedFields []string
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

// ChangeSet returns the trigger mutation change-set from trigger input.
func (input TriggerInput) ChangeSet() entityops.ChangeSet {
	return entityops.ChangeSet{
		ChangedFields:   input.ChangedFields,
		ClearedFields:   input.ClearedFields,
		ChangedEdges:    input.ChangedEdges,
		AddedIDs:        input.AddedIDs,
		RemovedIDs:      input.RemovedIDs,
		ProposedChanges: input.ProposedChanges,
		OldValues:       input.OldValues,
	}
}

// SetChangeSet applies a mutation change-set onto trigger input fields.
func (input *TriggerInput) SetChangeSet(changeSet entityops.ChangeSet) {
	if input == nil {
		return
	}

	cloned := changeSet.Clone()
	input.ChangedFields = cloned.ChangedFields
	input.ClearedFields = cloned.ClearedFields
	input.ChangedEdges = cloned.ChangedEdges
	input.AddedIDs = cloned.AddedIDs
	input.RemovedIDs = cloned.RemovedIDs
	input.OldValues = cloned.OldValues
	input.ProposedChanges = cloned.ProposedChanges
}
