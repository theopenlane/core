//go:build test

package eventstest_test

import (
	"context"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	"github.com/theopenlane/core/v2/internal/ent/generated/note"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi"
)

func TestQuestionnaireTransformListener(t *testing.T) {
	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, hooks.QuestionnaireTransformListeners())
	assert.NilError(t, err)
	defer setup.Teardown()

	user := suite.UserBuilder(context.Background(), t)
	orgID := user.OrganizationID
	internalCtx := th.SetInternalContext(user.UserCtx, suite.Client.DB)

	template := (&th.TemplateBuilder{Client: suite.Client}).MustNew(user.UserCtx, t)
	assert.NilError(t, suite.Client.DB.Template.UpdateOneID(template.ID).SetTransformConfiguration(models.TemplateProjectionConfig{
		Enabled: true,
		Mappings: []models.TemplateProjectionFieldMapping{
			{From: "vendorName", To: "name"},
			{From: "vendorNotes", To: "notes"},
		},
	}).Exec(internalCtx))

	assessment := (&th.AssessmentBuilder{Client: suite.Client, TemplateID: template.ID}).MustNew(user.UserCtx, t)

	vendorName := "acme-vendor-" + ulids.New().String()
	noteText := "vendor context from the questionnaire submission"

	doc, err := suite.Client.DB.DocumentData.Create().
		SetOwnerID(orgID).
		SetTemplateID(template.ID).
		SetData(map[string]any{"vendorName": vendorName, "vendorNotes": noteText}).
		Save(internalCtx)
	assert.NilError(t, err)

	response := (&th.AssessmentResponseBuilder{Client: suite.Client, AssessmentID: assessment.ID, OwnerID: orgID}).MustNew(user.UserCtx, t)
	noteRef := "questionnaire_transform:" + response.ID

	var entityID string

	t.Run("completed response transforms into entity", func(t *testing.T) {
		assert.NilError(t, suite.Client.DB.AssessmentResponse.UpdateOneID(response.ID).
			SetDocumentDataID(doc.ID).
			SetStatus(enums.AssessmentResponseStatusCompleted).
			Exec(internalCtx))

		waitForCondition(t, func() bool {
			updated, err := suite.Client.DB.AssessmentResponse.Get(internalCtx, response.ID)
			return err == nil && updated.EntityID != ""
		}, "assessment response should link the transformed entity")

		waitForGala(t, setup.Runtime)

		record, err := suite.Client.DB.Entity.Query().
			Where(entity.ExternalIDEQ(vendorName), entity.OwnerIDEQ(orgID)).
			Only(internalCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(vendorName, record.Name))

		meta, ok := record.VendorMetadata["questionnaire_transform"].(map[string]any)
		assert.Assert(t, ok)
		assert.Check(t, is.Equal(response.ID, meta["assessment_response_id"]))
		assert.Check(t, is.Equal(assessment.ID, meta["assessment_id"]))
		assert.Check(t, is.Equal(template.ID, meta["template_id"]))
		assert.Check(t, is.Equal(doc.ID, meta["document_data_id"]))

		updated, err := suite.Client.DB.AssessmentResponse.Get(internalCtx, response.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(record.ID, updated.EntityID))
		assert.Check(t, is.Equal(vendorName, updated.DisplayName))

		notes, err := suite.Client.DB.Note.Query().
			Where(note.NoteRefEQ(noteRef), note.OwnerIDEQ(orgID)).
			All(internalCtx)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(notes, 1))
		assert.Check(t, is.Equal(noteText, notes[0].Text))

		linkedNotes, err := suite.Client.DB.Entity.QueryNotes(record).All(internalCtx)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(linkedNotes, 1))
		assert.Check(t, is.Equal(notes[0].ID, linkedNotes[0].ID))

		linkedEntities, err := suite.Client.DB.DocumentData.QueryEntities(doc).All(internalCtx)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(linkedEntities, 1))
		assert.Check(t, is.Equal(record.ID, linkedEntities[0].ID))

		entityID = record.ID
	})

	t.Run("redelivery after entity link is a no-op", func(t *testing.T) {
		// completed is a terminal status, so the second qualifying update touches completed_at
		assert.NilError(t, suite.Client.DB.AssessmentResponse.UpdateOneID(response.ID).
			SetCompletedAt(time.Now()).
			Exec(internalCtx))

		waitForGala(t, setup.Runtime)

		entityCount, err := suite.Client.DB.Entity.Query().
			Where(entity.ExternalIDEQ(vendorName), entity.OwnerIDEQ(orgID)).
			Count(internalCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(1, entityCount))

		noteCount, err := suite.Client.DB.Note.Query().
			Where(note.NoteRefEQ(noteRef), note.OwnerIDEQ(orgID)).
			Count(internalCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(1, noteCount))

		updated, err := suite.Client.DB.AssessmentResponse.Get(internalCtx, response.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(entityID, updated.EntityID))
	})

	t.Run("response completed by anonymous respondent transforms into entity", func(t *testing.T) {
		anonVendorName := "acme-vendor-" + ulids.New().String()

		anonDoc, err := suite.Client.DB.DocumentData.Create().
			SetOwnerID(orgID).
			SetTemplateID(template.ID).
			SetData(map[string]any{"vendorName": anonVendorName}).
			Save(internalCtx)
		assert.NilError(t, err)

		anonResponse := (&th.AssessmentResponseBuilder{Client: suite.Client, AssessmentID: assessment.ID, OwnerID: orgID}).MustNew(user.UserCtx, t)

		// covers the listener under the anonymous respondent caller which is with the internal operation is here; the handler path is covered by TestSubmitQuestionnaire
		respondent := auth.NewQuestionnaireCaller(orgID, ulids.New().String(), "Anonymous Respondent", "")
		respondentCtx := auth.WithInternalOperationContext(auth.WithCaller(context.Background(), respondent))

		assert.NilError(t, suite.Client.DB.AssessmentResponse.UpdateOneID(anonResponse.ID).
			SetDocumentDataID(anonDoc.ID).
			SetStatus(enums.AssessmentResponseStatusCompleted).
			Exec(respondentCtx))

		waitForCondition(t, func() bool {
			updated, err := suite.Client.DB.AssessmentResponse.Get(internalCtx, anonResponse.ID)
			return err == nil && updated.EntityID != ""
		}, "assessment response completed anonymously should link the transformed entity")

		record, err := suite.Client.DB.Entity.Query().
			Where(entity.ExternalIDEQ(anonVendorName), entity.OwnerIDEQ(orgID)).
			Only(internalCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(anonVendorName, record.Name))

		t.Cleanup(func() {
			(&th.Cleanup[*generated.AssessmentResponseDeleteOne]{Client: suite.Client.DB.AssessmentResponse, ID: anonResponse.ID}).MustDelete(user.UserCtx, t)
			(&th.Cleanup[*generated.EntityDeleteOne]{Client: suite.Client.DB.Entity, ID: record.ID}).MustDelete(user.UserCtx, t)
		})
	})

	vendorName2 := "acme-vendor-" + ulids.New().String()

	doc2, err := suite.Client.DB.DocumentData.Create().
		SetOwnerID(orgID).
		SetTemplateID(template.ID).
		SetData(map[string]any{"vendorName": vendorName2}).
		Save(internalCtx)
	assert.NilError(t, err)

	response2 := (&th.AssessmentResponseBuilder{Client: suite.Client, AssessmentID: assessment.ID, OwnerID: orgID}).MustNew(user.UserCtx, t)

	t.Run("update without gate fields does not transform", func(t *testing.T) {
		assert.NilError(t, suite.Client.DB.AssessmentResponse.UpdateOneID(response2.ID).
			SetSendAttempts(2).
			Exec(internalCtx))

		waitForGala(t, setup.Runtime)

		updated, err := suite.Client.DB.AssessmentResponse.Get(internalCtx, response2.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal("", updated.EntityID))

		entityCount, err := suite.Client.DB.Entity.Query().
			Where(entity.ExternalIDEQ(vendorName2), entity.OwnerIDEQ(orgID)).
			Count(internalCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(0, entityCount))
	})

	t.Run("uncompleted response does not transform", func(t *testing.T) {
		assert.NilError(t, suite.Client.DB.AssessmentResponse.UpdateOneID(response2.ID).
			SetDocumentDataID(doc2.ID).
			Exec(internalCtx))

		waitForGala(t, setup.Runtime)

		updated, err := suite.Client.DB.AssessmentResponse.Get(internalCtx, response2.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal("", updated.EntityID))

		entityCount, err := suite.Client.DB.Entity.Query().
			Where(entity.ExternalIDEQ(vendorName2), entity.OwnerIDEQ(orgID)).
			Count(internalCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(0, entityCount))
	})

	dateTemplate := (&th.TemplateBuilder{Client: suite.Client}).MustNew(user.UserCtx, t)
	assert.NilError(t, suite.Client.DB.Template.UpdateOneID(dateTemplate.ID).SetTransformConfiguration(models.TemplateProjectionConfig{
		Enabled: true,
		Mappings: []models.TemplateProjectionFieldMapping{
			{From: "vendorName", To: "name"},
			{From: "contractStart", To: "contract_start_date"},
		},
	}).Exec(internalCtx))

	dateAssessment := (&th.AssessmentBuilder{Client: suite.Client, TemplateID: dateTemplate.ID}).MustNew(user.UserCtx, t)
	vendorName3 := "acme-vendor-" + ulids.New().String()

	dateDoc, err := suite.Client.DB.DocumentData.Create().
		SetOwnerID(orgID).
		SetTemplateID(dateTemplate.ID).
		SetData(map[string]any{"vendorName": vendorName3, "contractStart": "not-a-date"}).
		Save(internalCtx)
	assert.NilError(t, err)

	response3 := (&th.AssessmentResponseBuilder{Client: suite.Client, AssessmentID: dateAssessment.ID, OwnerID: orgID}).MustNew(user.UserCtx, t)

	t.Run("invalid date value fails validation and acks without retry", func(t *testing.T) {
		assert.NilError(t, suite.Client.DB.AssessmentResponse.UpdateOneID(response3.ID).
			SetDocumentDataID(dateDoc.ID).
			SetStatus(enums.AssessmentResponseStatusCompleted).
			Exec(internalCtx))

		// WaitIdle counts retrying jobs as active, so returning proves the ack
		waitForGala(t, setup.Runtime)

		updated, err := suite.Client.DB.AssessmentResponse.Get(internalCtx, response3.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal("", updated.EntityID))

		entityCount, err := suite.Client.DB.Entity.Query().
			Where(entity.ExternalIDEQ(vendorName3), entity.OwnerIDEQ(orgID)).
			Count(internalCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(0, entityCount))
	})

	(&th.Cleanup[*generated.AssessmentResponseDeleteOne]{Client: suite.Client.DB.AssessmentResponse, IDs: []string{response.ID, response2.ID, response3.ID}}).MustDelete(user.UserCtx, t)
	(&th.Cleanup[*generated.EntityDeleteOne]{Client: suite.Client.DB.Entity, ID: entityID}).MustDelete(user.UserCtx, t)
	(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, IDs: []string{assessment.ID, dateAssessment.ID}}).MustDelete(user.UserCtx, t)
	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, IDs: []string{template.ID, dateTemplate.ID}}).MustDelete(user.UserCtx, t)
}
