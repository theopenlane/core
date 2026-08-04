package notifications

import (
	"testing"

	"entgo.io/ent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/task"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/iam/auth"
)

func TestExtractMentionDetails(t *testing.T) {
	payload := eventqueue.MutationGalaPayload{
		MutationType: generated.TypeTask,
		Operation:    ent.OpUpdateOne.String(),
		EntityID:     "task-1",
		ProposedChanges: map[string]any{
			task.FieldTitle:   "Task One",
			task.FieldDetails: "details text",
			task.FieldOwnerID: "owner-1",
		},
		OldValues: map[string]any{
			task.FieldDetails: "old details text",
		},
	}

	details := extractMentionDetails(payload, task.FieldTitle, task.FieldDetails, task.FieldDetailsJSON, task.FieldOwnerID)

	assert.Equal(t, "task-1", details.objectID)
	assert.Equal(t, generated.TypeTask, details.objectType)
	assert.Equal(t, "Task One", details.objectName)
	assert.Equal(t, "details text", details.newDetails)
	assert.Equal(t, "old details text", details.oldDetails)
	assert.Equal(t, "owner-1", details.ownerID)
	assert.Empty(t, details.newDetailsJSON)
	assert.Empty(t, details.oldDetailsJSON)
}

func TestNoteParent(t *testing.T) {
	noteEntity := &generated.Note{ID: "note-1"}
	parentType, parentID, parentName := noteParent(noteEntity)
	assert.Equal(t, generated.TypeNote, parentType)
	assert.Equal(t, "note-1", parentID)
	assert.Equal(t, "Comment", parentName)

	noteEntity.Edges.Task = &generated.Task{ID: "task-1", Title: "Task One"}
	parentType, parentID, parentName = noteParent(noteEntity)
	assert.Equal(t, generated.TypeTask, parentType)
	assert.Equal(t, "task-1", parentID)
	assert.Equal(t, "Task One", parentName)
}

func TestIsExportNotificationAllowsSupportUser(t *testing.T) {
	assert.True(t, canProcessNotification(t.Context(), nil, auth.SupportSubjectID))
}

func TestRegisterGalaListeners(t *testing.T) {
	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := RegisterGalaListeners(runtime)
	require.NoError(t, err)
	require.Len(t, ids, 8)

	assert.True(t, runtime.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, generated.TypeTask), "create"))
	assert.True(t, runtime.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, generated.TypeInternalPolicy), "update"))
	assert.True(t, runtime.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, generated.TypeRisk), "delete"))
	assert.True(t, runtime.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, generated.TypeProcedure), "update_one"))
	assert.True(t, runtime.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, generated.TypeNote), "create"))
	assert.True(t, runtime.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, generated.TypeExport), "update"))
	assert.True(t, runtime.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, generated.TypeStandard), "update"))
	assert.True(t, runtime.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, generated.TypeProgram), ent.OpUpdate.String()))
	assert.True(t, runtime.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, generated.TypeProgram), ent.OpUpdateOne.String()))
	assert.False(t, runtime.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, generated.TypeProgram), ent.OpCreate.String()))
}
