package workflows

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/mutations"
)

// TestTriggerChangeSet verifies trigger change-set extraction clones map-backed context values
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
	}

	changeSet := TriggerChangeSet(contextData)
	changeSet.ChangedFields[0] = "mutated"
	changeSet.AddedIDs["controls"][0] = "mutated"
	changeSet.ProposedChanges["status"] = "mutated"

	require.Equal(t, "status", contextData.TriggerChangedFields[0])
	require.Equal(t, "one", contextData.TriggerAddedIDs["controls"][0])
	require.Equal(t, "approved", contextData.TriggerProposedChanges["status"])
}

// TestSetTriggerChangeSet verifies applying a change-set updates trigger context fields
func TestSetTriggerChangeSet(t *testing.T) {
	changeSet := mutations.ChangeSet{
		ChangedFields: []string{"status"},
		ChangedEdges:  []string{"controls"},
		AddedIDs: map[string][]string{
			"controls": {"one"},
		},
		RemovedIDs: map[string][]string{
			"controls": {"two"},
		},
		ProposedChanges: map[string]any{
			"status": "approved",
		},
	}

	var contextData models.WorkflowInstanceContext
	SetTriggerChangeSet(&contextData, changeSet)

	require.Equal(t, changeSet.ChangedFields, contextData.TriggerChangedFields)
	require.Equal(t, changeSet.ChangedEdges, contextData.TriggerChangedEdges)
	require.Equal(t, changeSet.AddedIDs, contextData.TriggerAddedIDs)
	require.Equal(t, changeSet.RemovedIDs, contextData.TriggerRemovedIDs)
	require.Equal(t, changeSet.ProposedChanges, contextData.TriggerProposedChanges)
}
