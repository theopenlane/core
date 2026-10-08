package notifications

import (
	"context"
	"testing"

	"entgo.io/ent"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/task"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/iam/auth"
)

func TestMentionTexts(t *testing.T) {
	spec, ok := entityops.MentionSpecFor(generated.TypeTask)
	require.True(t, ok)
	assert.Equal(t, task.FieldTitle, spec.NameField)
	assert.Equal(t, task.FieldDetails, spec.DetailsField)
	assert.Equal(t, task.FieldDetailsJSON, spec.DetailsJSONField)
	assert.Equal(t, task.FieldOwnerID, spec.OwnerField)

	payload := entityops.MutationPayload{
		MutationType: generated.TypeTask,
		Operation:    ent.OpUpdateOne.String(),
		EntityID:     "task-1",
		ChangeSet: entityops.ChangeSet{
			ProposedChanges: map[string]any{
				task.FieldTitle:   "Task One",
				task.FieldDetails: "details text",
				task.FieldOwnerID: "owner-1",
			},
			OldValues: map[string]any{
				task.FieldDetails: "old details text",
			},
		},
	}

	newText, oldText := entityops.MentionTexts(payload, spec)
	assert.Equal(t, "details text", newText)
	assert.Equal(t, "old details text", oldText)
}

func TestConsoleObjectPaths(t *testing.T) {
	assert.Equal(t, "automation/tasks?id=task-1", entityops.ConsoleObjectPath(generated.TypeTask, "task-1"))
	assert.Equal(t, "policies/pol-1/view", entityops.ConsoleObjectPath(generated.TypeInternalPolicy, "pol-1"))
	assert.Equal(t, "exposure/risks/risk-1", entityops.ConsoleObjectPath(generated.TypeRisk, "risk-1"))
	assert.Equal(t, "evidence?id=ev-1", entityops.ConsoleObjectPath(generated.TypeEvidence, "ev-1"))
	assert.Equal(t, "exposure/vulnerabilities?id=vul-1", entityops.ConsoleObjectPath(generated.TypeVulnerability, "vul-1"))
	assert.Equal(t, "exposure/findings?id=find-1", entityops.ConsoleObjectPath(generated.TypeFinding, "find-1"))
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

	noteEntity.Edges.Task = nil
	noteEntity.Edges.Vulnerability = &generated.Vulnerability{ID: "vul-1", DisplayID: "VUL-000001", CveID: "CVE-2026-0001"}
	parentType, parentID, parentName = noteParent(noteEntity)
	assert.Equal(t, generated.TypeVulnerability, parentType)
	assert.Equal(t, "vul-1", parentID)
	assert.Equal(t, "CVE-2026-0001", parentName)

	noteEntity.Edges.Vulnerability = nil
	noteEntity.Edges.Finding = &generated.Finding{ID: "find-1", DisplayID: "FIND-000001", DisplayName: "Open S3 bucket"}
	parentType, parentID, parentName = noteParent(noteEntity)
	assert.Equal(t, generated.TypeFinding, parentType)
	assert.Equal(t, "find-1", parentID)
	assert.Equal(t, "Open S3 bucket", parentName)

	noteEntity.Edges.Finding = &generated.Finding{ID: "find-2", DisplayID: "FIND-000002"}
	_, _, parentName = noteParent(noteEntity)
	assert.Equal(t, "FIND-000002", parentName)
}

func TestIsExportNotificationAllowsSupportUser(t *testing.T) {
	assert.True(t, canProcessNotification(t.Context(), nil, auth.SupportSubjectID))
}

func TestRegisterGalaListeners(t *testing.T) {
	runtime, err := gala.NewGala(context.Background(), gala.Config{DispatchMode: gala.DispatchModeInMemory, WorkerCount: 3})
	require.NoError(t, err)

	ids, err := gala.Register(runtime, Listeners()...)
	require.NoError(t, err)

	mentionable := lo.CountBy(entityops.AllSchemas(), func(s *entityops.Schema) bool {
		return s.MentionSpec != nil && s != entityops.SchemaNote
	})

	approvals := lo.CountBy(entityops.AllSchemas(), func(s *entityops.Schema) bool {
		return s.ApprovalSpec != nil
	})

	// mention and approval fan-outs plus five explicit listeners
	require.Len(t, ids, mentionable+approvals+5)

	assert.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeTask), entityops.OpCreate))
	assert.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeInternalPolicy), entityops.OpUpdate))
	assert.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeRisk), entityops.OpDelete))
	assert.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeProcedure), entityops.OpUpdateOne))
	assert.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeNote), entityops.OpCreate))
	assert.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeExport), entityops.OpUpdate))
	assert.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeStandard), ent.OpUpdate.String()))
	assert.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeProgram), ent.OpUpdate.String()))
	assert.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeProgram), ent.OpUpdateOne.String()))
	assert.False(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeProgram), ent.OpCreate.String()))
}
