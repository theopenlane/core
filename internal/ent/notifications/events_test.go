package notifications

import (
	"context"
	"testing"

	"entgo.io/ent"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/task"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/iam/auth"
)

func TestMentionTexts(t *testing.T) {
	spec, ok := entityops.MentionSpecFor(generated.TypeTask)
	assert.Assert(t, ok)
	assert.Check(t, is.Equal(task.FieldTitle, spec.NameField))
	assert.Check(t, is.Equal(task.FieldDetails, spec.DetailsField))
	assert.Check(t, is.Equal(task.FieldDetailsJSON, spec.DetailsJSONField))
	assert.Check(t, is.Equal(task.FieldOwnerID, spec.OwnerField))

	payload := entityops.MutationPayload{
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

	newText, oldText := entityops.MentionTexts(payload, spec)
	assert.Check(t, is.Equal("details text", newText))
	assert.Check(t, is.Equal("old details text", oldText))
}

func TestConsoleObjectPaths(t *testing.T) {
	assert.Check(t, is.Equal("automation/tasks?id=task-1", entityops.ConsoleObjectPath(generated.TypeTask, "task-1")))
	assert.Check(t, is.Equal("policies/pol-1/view", entityops.ConsoleObjectPath(generated.TypeInternalPolicy, "pol-1")))
	assert.Check(t, is.Equal("exposure/risks/risk-1", entityops.ConsoleObjectPath(generated.TypeRisk, "risk-1")))
	assert.Check(t, is.Equal("evidence?id=ev-1", entityops.ConsoleObjectPath(generated.TypeEvidence, "ev-1")))
	assert.Check(t, is.Equal("exposure/vulnerabilities?id=vul-1", entityops.ConsoleObjectPath(generated.TypeVulnerability, "vul-1")))
	assert.Check(t, is.Equal("exposure/findings?id=find-1", entityops.ConsoleObjectPath(generated.TypeFinding, "find-1")))
	assert.Check(t, is.Equal("trust-center/NDAs", entityops.ConsoleLanding(generated.TypeTrustCenterNDARequest)))
	assert.Check(t, is.Equal("", entityops.ConsoleObjectPath(generated.TypeNote, "note-1")))
}

func TestNoteParent(t *testing.T) {
	noteEntity := &generated.Note{ID: "note-1"}
	parentType, parentID, parentName := noteParent(noteEntity)
	assert.Check(t, is.Equal(generated.TypeNote, parentType))
	assert.Check(t, is.Equal("note-1", parentID))
	assert.Check(t, is.Equal("Comment", parentName))

	noteEntity.Edges.Task = &generated.Task{ID: "task-1", Title: "Task One"}
	parentType, parentID, parentName = noteParent(noteEntity)
	assert.Check(t, is.Equal(generated.TypeTask, parentType))
	assert.Check(t, is.Equal("task-1", parentID))
	assert.Check(t, is.Equal("Task One", parentName))

	noteEntity.Edges.Task = nil
	noteEntity.Edges.Vulnerability = &generated.Vulnerability{ID: "vul-1", DisplayID: "VUL-000001", CveID: "CVE-2026-0001"}
	parentType, parentID, parentName = noteParent(noteEntity)
	assert.Check(t, is.Equal(generated.TypeVulnerability, parentType))
	assert.Check(t, is.Equal("vul-1", parentID))
	assert.Check(t, is.Equal("CVE-2026-0001", parentName))

	noteEntity.Edges.Vulnerability = nil
	noteEntity.Edges.Finding = &generated.Finding{ID: "find-1", DisplayID: "FIND-000001", DisplayName: "Open S3 bucket"}
	parentType, parentID, parentName = noteParent(noteEntity)
	assert.Check(t, is.Equal(generated.TypeFinding, parentType))
	assert.Check(t, is.Equal("find-1", parentID))
	assert.Check(t, is.Equal("Open S3 bucket", parentName))

	noteEntity.Edges.Finding = &generated.Finding{ID: "find-2", DisplayID: "FIND-000002"}
	_, _, parentName = noteParent(noteEntity)
	assert.Check(t, is.Equal("FIND-000002", parentName))
}

func TestIsExportNotificationAllowsSupportUser(t *testing.T) {
	assert.Check(t, canProcessNotification(t.Context(), nil, auth.SupportSubjectID))
}

func TestRegisterGalaListeners(t *testing.T) {
	runtime, err := gala.NewGala(context.Background(), gala.Config{DispatchMode: gala.DispatchModeInMemory, WorkerCount: 3})
	assert.NilError(t, err)

	ids, err := gala.Register(runtime, Listeners()...)
	assert.NilError(t, err)

	mentionable := lo.CountBy(entityops.AllSchemas(), func(s *entityops.Schema) bool {
		return s.MentionSpec != nil && s != entityops.SchemaNote
	})

	approvals := lo.CountBy(entityops.AllSchemas(), func(s *entityops.Schema) bool {
		return s.ApprovalSpec != nil
	})

	// mention and approval fan-outs plus six explicit listeners
	assert.Assert(t, is.Len(ids, mentionable+approvals+6))

	assert.Check(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeTask), entityops.OpCreate))
	assert.Check(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeInternalPolicy), entityops.OpUpdate))
	assert.Check(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeRisk), entityops.OpDelete))
	assert.Check(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeProcedure), entityops.OpUpdateOne))
	assert.Check(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeNote), entityops.OpCreate))
	assert.Check(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeExport), entityops.OpUpdate))
	assert.Check(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeStandard), ent.OpUpdate.String()))
	assert.Check(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeProgram), ent.OpUpdate.String()))
	assert.Check(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeProgram), ent.OpUpdateOne.String()))
	assert.Check(t, !(runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeProgram), ent.OpCreate.String())))
	assert.Check(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeReportMembership), entityops.OpCreate))
	assert.Check(t, !(runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeReportMembership), entityops.OpUpdate)))
	assert.Check(t, !(runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeReportMembership), entityops.OpUpdateOne)))
	assert.Check(t, !(runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernNotification, generated.TypeReportMembership), entityops.OpDelete)))
}
