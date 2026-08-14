//go:build test

package graphapi_test

import (
	"context"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/entity"
	"github.com/theopenlane/core/internal/ent/generated/note"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/graphapi"
)

func TestQuestionnaireTransformListener(t *testing.T) {
	setup, err := graphapi.SetupListenerRuntime(context.Background(), suite.client.db, suite.tf.URI, hooks.QuestionnaireTransformListeners())
	assert.NilError(t, err)
	defer setup.Teardown()

	user := suite.userBuilder(context.Background(), t)
	orgID := user.OrganizationID
	allowCtx := privacy.DecisionContext(setContext(user.UserCtx, suite.client.db), privacy.Allow)

	template := (&TemplateBuilder{client: suite.client}).MustNew(user.UserCtx, t)
	assert.NilError(t, suite.client.db.Template.UpdateOneID(template.ID).SetTransformConfiguration(models.TemplateProjectionConfig{
		Enabled: true,
		Target:  enums.TemplateProjectionTargetEntity,
		Mappings: []models.TemplateProjectionFieldMapping{
			{From: "vendorName", To: "name"},
			{From: "vendorNotes", To: "notes"},
		},
	}).Exec(allowCtx))

	assessment := (&AssessmentBuilder{client: suite.client, TemplateID: template.ID}).MustNew(user.UserCtx, t)

	vendorName := "acme-vendor-" + ulids.New().String()
	noteText := "vendor context from the questionnaire submission"

	doc, err := suite.client.db.DocumentData.Create().
		SetOwnerID(orgID).
		SetTemplateID(template.ID).
		SetData(map[string]any{"vendorName": vendorName, "vendorNotes": noteText}).
		Save(allowCtx)
	assert.NilError(t, err)

	response := (&AssessmentResponseBuilder{client: suite.client, AssessmentID: assessment.ID, OwnerID: orgID}).MustNew(user.UserCtx, t)
	noteRef := "questionnaire_transform:" + response.ID

	var entityID string

	t.Run("completed response transforms into entity", func(t *testing.T) {
		assert.NilError(t, suite.client.db.AssessmentResponse.UpdateOneID(response.ID).
			SetDocumentDataID(doc.ID).
			SetStatus(enums.AssessmentResponseStatusCompleted).
			Exec(allowCtx))

		waitForCondition(t, func() bool {
			updated, err := suite.client.db.AssessmentResponse.Get(allowCtx, response.ID)
			return err == nil && updated.EntityID != ""
		}, "assessment response should link the transformed entity")

		setup.Runtime.WaitIdle()

		record, err := suite.client.db.Entity.Query().
			Where(entity.ExternalIDEQ(vendorName), entity.OwnerIDEQ(orgID)).
			Only(allowCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(vendorName, record.Name))

		meta, ok := record.VendorMetadata["questionnaire_transform"].(map[string]any)
		assert.Assert(t, ok)
		assert.Check(t, is.Equal(response.ID, meta["assessment_response_id"]))
		assert.Check(t, is.Equal(assessment.ID, meta["assessment_id"]))
		assert.Check(t, is.Equal(template.ID, meta["template_id"]))
		assert.Check(t, is.Equal(doc.ID, meta["document_data_id"]))

		updated, err := suite.client.db.AssessmentResponse.Get(allowCtx, response.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(record.ID, updated.EntityID))
		assert.Check(t, is.Equal(vendorName, updated.DisplayName))

		notes, err := suite.client.db.Note.Query().
			Where(note.NoteRefEQ(noteRef), note.OwnerIDEQ(orgID)).
			All(allowCtx)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(notes, 1))
		assert.Check(t, is.Equal(noteText, notes[0].Text))

		linkedNotes, err := suite.client.db.Entity.QueryNotes(record).All(allowCtx)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(linkedNotes, 1))
		assert.Check(t, is.Equal(notes[0].ID, linkedNotes[0].ID))

		linkedEntities, err := suite.client.db.DocumentData.QueryEntities(doc).All(allowCtx)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(linkedEntities, 1))
		assert.Check(t, is.Equal(record.ID, linkedEntities[0].ID))

		entityID = record.ID
	})

	t.Run("redelivery after entity link is a no-op", func(t *testing.T) {
		// completed is a terminal status, so the second qualifying update touches completed_at
		assert.NilError(t, suite.client.db.AssessmentResponse.UpdateOneID(response.ID).
			SetCompletedAt(time.Now()).
			Exec(allowCtx))

		setup.Runtime.WaitIdle()

		entityCount, err := suite.client.db.Entity.Query().
			Where(entity.ExternalIDEQ(vendorName), entity.OwnerIDEQ(orgID)).
			Count(allowCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(1, entityCount))

		noteCount, err := suite.client.db.Note.Query().
			Where(note.NoteRefEQ(noteRef), note.OwnerIDEQ(orgID)).
			Count(allowCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(1, noteCount))

		updated, err := suite.client.db.AssessmentResponse.Get(allowCtx, response.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(entityID, updated.EntityID))
	})

	vendorName2 := "acme-vendor-" + ulids.New().String()

	doc2, err := suite.client.db.DocumentData.Create().
		SetOwnerID(orgID).
		SetTemplateID(template.ID).
		SetData(map[string]any{"vendorName": vendorName2}).
		Save(allowCtx)
	assert.NilError(t, err)

	response2 := (&AssessmentResponseBuilder{client: suite.client, AssessmentID: assessment.ID, OwnerID: orgID}).MustNew(user.UserCtx, t)

	t.Run("update without gate fields does not transform", func(t *testing.T) {
		assert.NilError(t, suite.client.db.AssessmentResponse.UpdateOneID(response2.ID).
			SetSendAttempts(2).
			Exec(allowCtx))

		setup.Runtime.WaitIdle()

		updated, err := suite.client.db.AssessmentResponse.Get(allowCtx, response2.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal("", updated.EntityID))

		entityCount, err := suite.client.db.Entity.Query().
			Where(entity.ExternalIDEQ(vendorName2), entity.OwnerIDEQ(orgID)).
			Count(allowCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(0, entityCount))
	})

	t.Run("uncompleted response does not transform", func(t *testing.T) {
		assert.NilError(t, suite.client.db.AssessmentResponse.UpdateOneID(response2.ID).
			SetDocumentDataID(doc2.ID).
			Exec(allowCtx))

		setup.Runtime.WaitIdle()

		updated, err := suite.client.db.AssessmentResponse.Get(allowCtx, response2.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal("", updated.EntityID))

		entityCount, err := suite.client.db.Entity.Query().
			Where(entity.ExternalIDEQ(vendorName2), entity.OwnerIDEQ(orgID)).
			Count(allowCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(0, entityCount))
	})

	(&Cleanup[*generated.AssessmentResponseDeleteOne]{client: suite.client.db.AssessmentResponse, IDs: []string{response.ID, response2.ID}}).MustDelete(user.UserCtx, t)
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: entityID}).MustDelete(user.UserCtx, t)
	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, ID: assessment.ID}).MustDelete(user.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: template.ID}).MustDelete(user.UserCtx, t)
}
