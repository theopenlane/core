package workflows

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
)

// TestTriggerChangeSet verifies trigger change-set extraction marshals map-backed proposed
// changes into opaque JSON without aliasing the context's map values
func TestTriggerChangeSet(t *testing.T) {
	contextData := models.WorkflowInstanceContext{
		TriggerChangedFields: []string{"status"},
		TriggerChangedEdges:  []string{"controls"},
		TriggerAddedIDs: map[string][]string{
			"controls": {"one"},
		},
		TriggerRemovedIDs: map[string][]string{
			"controls": {"two"},
		},
		TriggerProposedChanges: map[string]any{
			"status": "approved",
		},
		TriggerOldValues: map[string]any{
			"status": "draft",
		},
	}

	changeSet := TriggerChangeSet(contextData)

	assert.Assert(t, is.DeepEqual([]string{"status"}, changeSet.ChangedFields))
	assert.Assert(t, is.DeepEqual(map[string]any{"status": "approved"}, changeSet.ProposedChanges))

	changeSet.ChangedFields[0] = "mutated"
	changeSet.AddedIDs["controls"][0] = "mutated"
	changeSet.OldValues["status"] = "mutated"

	assert.Assert(t, is.Equal("status", contextData.TriggerChangedFields[0]))
	assert.Assert(t, is.Equal("one", contextData.TriggerAddedIDs["controls"][0]))
	assert.Assert(t, is.Equal("approved", contextData.TriggerProposedChanges["status"]))
	assert.Assert(t, is.Equal("draft", contextData.TriggerOldValues["status"]))
}

// TestSetTriggerChangeSet verifies applying a change-set decodes proposed changes back
// into the engine's map shape on the trigger context
func TestSetTriggerChangeSet(t *testing.T) {
	changeSet := entityops.ChangeSet{
		ChangedFields: []string{"status"},
		ChangedEdges:  []string{"controls"},
		AddedIDs: map[string][]string{
			"controls": {"one"},
		},
		RemovedIDs: map[string][]string{
			"controls": {"two"},
		},
		ProposedChanges: map[string]any{"status": "approved"},
		OldValues: map[string]any{
			"status": "draft",
		},
	}

	var contextData models.WorkflowInstanceContext
	SetTriggerChangeSet(&contextData, changeSet)

	assert.Assert(t, is.DeepEqual(changeSet.ChangedFields, contextData.TriggerChangedFields))
	assert.Assert(t, is.DeepEqual(changeSet.ChangedEdges, contextData.TriggerChangedEdges))
	assert.Assert(t, is.DeepEqual(changeSet.AddedIDs, contextData.TriggerAddedIDs))
	assert.Assert(t, is.DeepEqual(changeSet.RemovedIDs, contextData.TriggerRemovedIDs))
	assert.Assert(t, is.DeepEqual(map[string]any{"status": "approved"}, contextData.TriggerProposedChanges))
	assert.Assert(t, is.DeepEqual(changeSet.OldValues, contextData.TriggerOldValues))
}
