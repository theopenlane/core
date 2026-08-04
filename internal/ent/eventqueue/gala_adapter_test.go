package eventqueue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewGalaHeadersFromMutationMetadata verifies property normalization for gala headers
func TestNewGalaHeadersFromMutationMetadata(t *testing.T) {
	t.Parallel()

	headers := NewGalaHeadersFromMutationMetadata(MutationGalaMetadata{
		Properties: map[string]string{
			"active": "true",
			"count":  "5",
			"":       "ignored",
		},
	})

	assert.Equal(t, "true", headers.Properties["active"])
	assert.Equal(t, "5", headers.Properties["count"])
	_, exists := headers.Properties[""]
	assert.False(t, exists)
}

// TestMutationGalaPayloadChangeSetRoundTrip verifies payload change-set projections preserve values and clone maps/slices
func TestMutationGalaPayloadChangeSetRoundTrip(t *testing.T) {
	t.Parallel()

	payload := MutationGalaPayload{
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
		OldValues: map[string]any{
			"status": "draft",
		},
	}

	changeSet := payload.ChangeSet()
	changeSet.ChangedFields[0] = "mutated"
	changeSet.AddedIDs["controls"][0] = "mutated"
	changeSet.ProposedChanges["status"] = "mutated"
	changeSet.OldValues["status"] = "mutated"

	assert.Equal(t, "status", payload.ChangedFields[0])
	assert.Equal(t, "one", payload.AddedIDs["controls"][0])
	assert.Equal(t, "approved", payload.ProposedChanges["status"])
	assert.Equal(t, "draft", payload.OldValues["status"])

	var roundTrip MutationGalaPayload
	roundTrip.SetChangeSet(payload.ChangeSet())
	assert.Equal(t, payload.ChangedFields, roundTrip.ChangedFields)
	assert.Equal(t, payload.ChangedEdges, roundTrip.ChangedEdges)
	assert.Equal(t, payload.AddedIDs, roundTrip.AddedIDs)
	assert.Equal(t, payload.RemovedIDs, roundTrip.RemovedIDs)
	assert.Equal(t, payload.ProposedChanges, roundTrip.ProposedChanges)
	assert.Equal(t, payload.OldValues, roundTrip.OldValues)
}
