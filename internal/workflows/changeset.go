package workflows

import (
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/entityops"
)

// TriggerChangeSet returns the trigger mutation change-set carried by workflow instance
// context.
func TriggerChangeSet(ctx models.WorkflowInstanceContext) entityops.ChangeSet {
	set := entityops.ChangeSet{
		ChangedFields:   ctx.TriggerChangedFields,
		ChangedEdges:    ctx.TriggerChangedEdges,
		AddedIDs:        ctx.TriggerAddedIDs,
		RemovedIDs:      ctx.TriggerRemovedIDs,
		ProposedChanges: ctx.TriggerProposedChanges,
		OldValues:       ctx.TriggerOldValues,
	}

	// clone so callers can mutate the change set without aliasing the persisted context
	return set.Clone()
}

// SetTriggerChangeSet applies a mutation change-set to workflow instance trigger context
// fields.
func SetTriggerChangeSet(ctx *models.WorkflowInstanceContext, changeSet entityops.ChangeSet) {
	if ctx == nil {
		return
	}

	cloned := changeSet.Clone()
	ctx.TriggerChangedFields = cloned.ChangedFields
	ctx.TriggerChangedEdges = cloned.ChangedEdges
	ctx.TriggerAddedIDs = cloned.AddedIDs
	ctx.TriggerRemovedIDs = cloned.RemovedIDs
	ctx.TriggerOldValues = cloned.OldValues

	ctx.TriggerProposedChanges = cloned.ProposedMap()
}
