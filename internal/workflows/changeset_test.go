package workflows

import (
	"testing"

	"github.com/stretchr/testify/require"

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

	require.Equal(t, []string{"status"}, changeSet.ChangedFields)
	require.Equal(t, map[string]any{"status": "approved"}, changeSet.ProposedChanges)

	changeSet.ChangedFields[0] = "mutated"
	changeSet.AddedIDs["controls"][0] = "mutated"
	changeSet.OldValues["status"] = "mutated"

	require.Equal(t, "status", contextData.TriggerChangedFields[0])
	require.Equal(t, "one", contextData.TriggerAddedIDs["controls"][0])
	require.Equal(t, "approved", contextData.TriggerProposedChanges["status"])
	require.Equal(t, "draft", contextData.TriggerOldValues["status"])
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

	require.Equal(t, changeSet.ChangedFields, contextData.TriggerChangedFields)
	require.Equal(t, changeSet.ChangedEdges, contextData.TriggerChangedEdges)
	require.Equal(t, changeSet.AddedIDs, contextData.TriggerAddedIDs)
	require.Equal(t, changeSet.RemovedIDs, contextData.TriggerRemovedIDs)
	require.Equal(t, map[string]any{"status": "approved"}, contextData.TriggerProposedChanges)
	require.Equal(t, changeSet.OldValues, contextData.TriggerOldValues)
}
