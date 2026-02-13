package notifications

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/internal/ent/generated/note"
)

func TestExtractNoteFromPayloadMetadata(t *testing.T) {
	fields := &noteFields{}
	payload := &events.MutationPayload{
		EntityID: "note-123",
		ProposedChanges: map[string]any{
			note.FieldText:     "hello world",
			note.FieldTextJSON: []any{map[string]any{"type": "paragraph"}},
			note.FieldOwnerID:  "owner-xyz",
		},
	}

	extractNoteFromPayload(payload, fields)

	assert.Equal(t, "note-123", fields.entityID)
	assert.Equal(t, "hello world", fields.text)
	assert.Equal(t, "owner-xyz", fields.ownerID)
	assert.NotEmpty(t, fields.textJSON)
}

func TestNoteHasParentReference(t *testing.T) {
	assert.False(t, noteHasParentReference(&noteFields{}))
	assert.True(t, noteHasParentReference(&noteFields{taskID: "task-1"}))
	assert.True(t, noteHasParentReference(&noteFields{controlID: "control-1"}))
}
