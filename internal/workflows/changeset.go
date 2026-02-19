package workflows

import (
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/mutations"
)

// TriggerChangeSet returns the trigger mutation change-set carried by workflow instance context
func TriggerChangeSet(ctx models.WorkflowInstanceContext) mutations.ChangeSet {
	return mutations.NewChangeSet(ctx.TriggerChangedFields, ctx.TriggerChangedEdges, ctx.TriggerAddedIDs, ctx.TriggerRemovedIDs, ctx.TriggerProposedChanges)
}

// SetTriggerChangeSet applies a mutation change-set to workflow instance trigger context fields
func SetTriggerChangeSet(ctx *models.WorkflowInstanceContext, changeSet mutations.ChangeSet) {
	if ctx == nil {
		return
	}

	cloned := changeSet.Clone()
	ctx.TriggerChangedFields = cloned.ChangedFields
	ctx.TriggerChangedEdges = cloned.ChangedEdges
	ctx.TriggerAddedIDs = cloned.AddedIDs
	ctx.TriggerRemovedIDs = cloned.RemovedIDs
	ctx.TriggerProposedChanges = cloned.ProposedChanges
}
