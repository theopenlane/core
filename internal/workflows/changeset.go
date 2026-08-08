package workflows

import (
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/pkg/jsonx"
)

// TriggerChangeSet returns the trigger mutation change-set carried by workflow instance
// context; the engine's map-shaped proposed changes marshal into the delta contract's
// opaque JSON at this boundary
func TriggerChangeSet(ctx models.WorkflowInstanceContext) entityops.ChangeSet {
	proposed, err := jsonx.ToRawMessage(ctx.TriggerProposedChanges)
	if err != nil || len(ctx.TriggerProposedChanges) == 0 {
		proposed = nil
	}

	set := entityops.ChangeSet{
		ChangedFields:   ctx.TriggerChangedFields,
		ChangedEdges:    ctx.TriggerChangedEdges,
		AddedIDs:        ctx.TriggerAddedIDs,
		RemovedIDs:      ctx.TriggerRemovedIDs,
		ProposedChanges: proposed,
		OldValues:       ctx.TriggerOldValues,
	}

	// clone so callers can mutate the change set without aliasing the persisted context
	return set.Clone()
}

// SetTriggerChangeSet applies a mutation change-set to workflow instance trigger context
// fields, decoding the opaque proposed-changes JSON into the engine's map shape
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

	proposed, err := jsonx.Decode[map[string]any](cloned.ProposedChanges)
	if err != nil {
		proposed = nil
	}

	ctx.TriggerProposedChanges = proposed
}
