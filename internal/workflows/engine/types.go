package engine

import "github.com/theopenlane/core/internal/mutations"

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
}

// ChangeSet returns the trigger mutation change-set from trigger input
func (input TriggerInput) ChangeSet() mutations.ChangeSet {
	return mutations.NewChangeSet(input.ChangedFields, input.ChangedEdges, input.AddedIDs, input.RemovedIDs, input.ProposedChanges)
}

// SetChangeSet applies a mutation change-set onto trigger input fields
func (input *TriggerInput) SetChangeSet(changeSet mutations.ChangeSet) {
	if input == nil {
		return
	}

	cloned := changeSet.Clone()
	input.ChangedFields = cloned.ChangedFields
	input.ChangedEdges = cloned.ChangedEdges
	input.AddedIDs = cloned.AddedIDs
	input.RemovedIDs = cloned.RemovedIDs
	input.ProposedChanges = cloned.ProposedChanges
}
