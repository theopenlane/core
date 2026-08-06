package notifications

import (
	"testing"

	"entgo.io/ent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/task"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/iam/auth"
)

func TestExtractMentionDetails(t *testing.T) {
	spec, ok := entityops.MentionSpecFor(generated.TypeTask)
	require.True(t, ok)
	assert.Equal(t, task.FieldTitle, spec.NameField)
	assert.Equal(t, task.FieldDetails, spec.DetailsField)
	assert.Equal(t, task.FieldDetailsJSON, spec.DetailsJSONField)
	assert.Equal(t, task.FieldOwnerID, spec.OwnerField)

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

	details := extractMentionDetails(payload, spec)

	assert.Equal(t, "task-1", details.objectID)
	assert.Equal(t, generated.TypeTask, details.objectType)
	assert.Equal(t, "Task One", details.objectName)
	assert.Equal(t, "details text", details.newDetails)
	assert.Equal(t, "old details text", details.oldDetails)
	assert.Equal(t, "owner-1", details.ownerID)
	assert.Empty(t, details.newDetailsJSON)
	assert.Empty(t, details.oldDetailsJSON)
}

func TestConsoleObjectPaths(t *testing.T) {
	assert.Equal(t, "automation/tasks?id=task-1", entityops.ConsoleObjectPath(generated.TypeTask, "task-1"))
	assert.Equal(t, "policies/pol-1/view", entityops.ConsoleObjectPath(generated.TypeInternalPolicy, "pol-1"))
	assert.Equal(t, "exposure/risks/risk-1", entityops.ConsoleObjectPath(generated.TypeRisk, "risk-1"))
	assert.Equal(t, "evidence?id=ev-1", entityops.ConsoleObjectPath(generated.TypeEvidence, "ev-1"))
	assert.Equal(t, "trust-center/NDAs", entityops.ConsoleLanding(generated.TypeTrustCenterNDARequest))
	assert.Empty(t, entityops.ConsoleObjectPath(generated.TypeNote, "note-1"))
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
