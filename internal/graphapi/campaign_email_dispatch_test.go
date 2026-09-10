//go:build test

package graphapi_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/newman/providers/mock"
)

// TestCampaignEmailDispatch verifies that SendBrandedCampaign renders
// campaign emails with the correct branding, template variables, and
// metadata, then sends one email per target via the mock sender
func TestCampaignEmailDispatch(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	// --- fixtures ---

	emailTemplate, err := suite.Client.DB.EmailTemplate.Create().
		SetName("Campaign Dispatch Test Template").
		SetKey(email.BrandedMessageOp.Name()).
		SetTemplateContext(enums.TemplateContextCampaignRecipient).
		SetDefaults(map[string]any{
			"subject":      "Hello {{ .firstName }}",
			"title":        "Default Title",
			"intros":       []any{"Campaign: {{ .campaignName }}"},
			"primaryColor": "#333333",
		}).
		Save(ctx)
	th.RequireNoError(t, err)

	campaignObj, err := suite.Client.DB.Campaign.Create().
		SetName("Dispatch Integration Test Campaign").
		SetDescription("Testing email dispatch pipeline").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetEmailTemplateID(emailTemplate.ID).
		SetRecurrenceFrequency(enums.FrequencyNone).
		SetMetadata(map[string]any{
			"title": "Campaign Override",
		}).
		Save(ctx)
	th.RequireNoError(t, err)

	targetAlice, err := suite.Client.DB.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("alice@test.example").
		SetFullName("Alice Smith").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		Save(ctx)
	th.RequireNoError(t, err)

	targetBob, err := suite.Client.DB.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("bob@test.example").
		SetFullName("Bob Jones").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*generated.CampaignTargetDeleteOne]{
			Client: suite.Client.DB.CampaignTarget,
			IDs:    []string{targetAlice.ID, targetBob.ID},
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.CampaignDeleteOne]{
			Client: suite.Client.DB.Campaign,
			ID:     campaignObj.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.EmailTemplateDeleteOne]{
			Client: suite.Client.DB.EmailTemplate,
			ID:     emailTemplate.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	// --- dispatch via SendBrandedCampaign operation ---

	mockSender, err := mock.New("")
	assert.NilError(t, err)

	emailClient := &email.Client{
		Sender: mockSender,
		Config: email.RuntimeEmailConfig{
			FromEmail:      "test@mail.example.com",
			CompanyName:    "TestCorp",
			CompanyAddress: "123 Test St",
			Corporation:    "TestCorp, Inc.",
			SupportEmail:   "support@test.example",
			LogoURL:        "https://example.com/logo.png",
			RootURL:        "https://www.example.com",
			ProductURL:     "https://app.example.com",
		},
	}

	cfg := email.SendBrandedCampaignRequest{CampaignDispatchInput: email.CampaignDispatchInput{CampaignID: campaignObj.ID}}
	req := types.OperationRequest{
		Client: emailClient,
		DB:     suite.Client.DB,
	}

	configBytes, err := json.Marshal(cfg)
	assert.NilError(t, err)
	req.Config = configBytes

	_, err = email.SendBrandedCampaign{}.Run(ctx, req, emailClient, cfg)
	assert.NilError(t, err)

	// --- verify sent messages ---

	messages := mockSender.Messages()
	assert.Assert(t, is.Len(messages, 2))

	var allHTML, allSubjects, allTo []string
	for _, msg := range messages {
		allHTML = append(allHTML, msg.HTML)
		allSubjects = append(allSubjects, msg.Subject)
		allTo = append(allTo, msg.To...)
	}

	combined := strings.Join(allHTML, "\n") + "\n" + strings.Join(allSubjects, "\n")

	t.Run("subject contains recipient first name", func(t *testing.T) {
		assert.Assert(t, strings.Contains(strings.Join(allSubjects, " "), "Hello Alice") || strings.Contains(strings.Join(allSubjects, " "), "Hello Bob"),
			"expected subject with first name")
	})

	t.Run("body contains campaign name", func(t *testing.T) {
		assert.Assert(t, strings.Contains(combined, "Dispatch Integration Test Campaign"),
			"expected campaign name in body")
	})

	t.Run("metadata overrides defaults", func(t *testing.T) {
		assert.Assert(t, strings.Contains(combined, "Campaign Override"),
			"expected metadata title to override default")
		assert.Assert(t, !strings.Contains(combined, "Default Title"),
			"default title should be overridden by metadata")
	})

	t.Run("catalog defaults rendered", func(t *testing.T) {
		assert.Assert(t, strings.Contains(combined, "Campaign: Dispatch Integration Test Campaign"),
			"expected catalog default intro with campaign name")
	})

	t.Run("each target gets its own message", func(t *testing.T) {
		allToStr := strings.Join(allTo, " ")
		assert.Assert(t, strings.Contains(allToStr, "alice@test.example"), "expected message for alice")
		assert.Assert(t, strings.Contains(allToStr, "bob@test.example"), "expected message for bob")
	})

	t.Run("campaign target tag present", func(t *testing.T) {
		found := false
		for _, msg := range messages {
			for _, tag := range msg.Tags {
				if tag.Name == email.TagCampaignTargetID {
					found = true
				}
			}
		}
		assert.Assert(t, found, "expected campaign_target_id tag")
	})

	t.Run("from address matches config", func(t *testing.T) {
		for _, msg := range messages {
			assert.Equal(t, msg.From, "test@mail.example.com")
		}
	})

	t.Run("sent_at marked on targets", func(t *testing.T) {
		updatedAlice, err := suite.Client.DB.CampaignTarget.Get(ctx, targetAlice.ID)
		assert.NilError(t, err)
		assert.Assert(t, updatedAlice.SentAt != nil, "expected sent_at to be set for alice")

		updatedBob, err := suite.Client.DB.CampaignTarget.Get(ctx, targetBob.ID)
		assert.NilError(t, err)
		assert.Assert(t, updatedBob.SentAt != nil, "expected sent_at to be set for bob")
	})
}

// TestCampaignEmailDispatchSkipsSentTargets verifies that targets with
// sent_at already set are not re-dispatched
func TestCampaignEmailDispatchSkipsSentTargets(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	emailTemplate, err := suite.Client.DB.EmailTemplate.Create().
		SetName("Skip Sent Test Template").
		SetKey(email.BrandedMessageOp.Name()).
		SetTemplateContext(enums.TemplateContextCampaignRecipient).
		SetDefaults(map[string]any{
			"subject": "Test",
			"title":   "Test",
			"intros":  []any{"Test"},
		}).
		Save(ctx)
	th.RequireNoError(t, err)

	campaignObj, err := suite.Client.DB.Campaign.Create().
		SetName("Skip Sent Test Campaign").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetEmailTemplateID(emailTemplate.ID).
		SetRecurrenceFrequency(enums.FrequencyNone).
		Save(ctx)
	th.RequireNoError(t, err)

	sentTarget, err := suite.Client.DB.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("already-sent@test.example").
		SetFullName("Already Sent").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		Save(ctx)
	th.RequireNoError(t, err)

	sentAt := models.DateTime(time.Now())
	_, err = suite.Client.DB.CampaignTarget.UpdateOneID(sentTarget.ID).
		SetSentAt(sentAt).
		Save(ctx)
	th.RequireNoError(t, err)

	unsentTarget, err := suite.Client.DB.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("unsent@test.example").
		SetFullName("Unsent Target").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*generated.CampaignTargetDeleteOne]{
			Client: suite.Client.DB.CampaignTarget,
			IDs:    []string{sentTarget.ID, unsentTarget.ID},
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.CampaignDeleteOne]{
			Client: suite.Client.DB.Campaign,
			ID:     campaignObj.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.EmailTemplateDeleteOne]{
			Client: suite.Client.DB.EmailTemplate,
			ID:     emailTemplate.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	mockSender, err := mock.New("")
	assert.NilError(t, err)

	emailClient := &email.Client{
		Sender: mockSender,
		Config: email.RuntimeEmailConfig{
			FromEmail:   "test@mail.example.com",
			CompanyName: "TestCorp",
			ProductURL:  "https://app.example.com",
		},
	}

	cfg := email.SendBrandedCampaignRequest{CampaignDispatchInput: email.CampaignDispatchInput{CampaignID: campaignObj.ID}}
	req := types.OperationRequest{
		Client: emailClient,
		DB:     suite.Client.DB,
	}

	configBytes, err := json.Marshal(cfg)
	assert.NilError(t, err)
	req.Config = configBytes

	_, err = email.SendBrandedCampaign{}.Run(ctx, req, emailClient, cfg)
	assert.NilError(t, err)

	messages := mockSender.Messages()
	assert.Assert(t, is.Len(messages, 1))
	assert.Assert(t, strings.Contains(strings.Join(messages[0].To, " "), "unsent@test.example"))
}

// TestCampaignEmailDispatchNoBranding verifies dispatch works without
// an EmailBranding record attached to the campaign
func TestCampaignEmailDispatchNoBranding(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	emailTemplate, err := suite.Client.DB.EmailTemplate.Create().
		SetName("No Branding Test Template").
		SetKey(email.BrandedMessageOp.Name()).
		SetTemplateContext(enums.TemplateContextCampaignRecipient).
		SetDefaults(map[string]any{
			"subject": "Hello {{ .firstName }}",
			"title":   "Welcome {{ .firstName }}",
			"intros":  []any{"Welcome {{ .firstName }}"},
		}).
		Save(ctx)
	th.RequireNoError(t, err)

	campaignObj, err := suite.Client.DB.Campaign.Create().
		SetName("No Branding Campaign").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetEmailTemplateID(emailTemplate.ID).
		SetRecurrenceFrequency(enums.FrequencyNone).
		Save(ctx)
	th.RequireNoError(t, err)

	target, err := suite.Client.DB.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("charlie@test.example").
		SetFullName("Charlie Brown").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*generated.CampaignTargetDeleteOne]{
			Client: suite.Client.DB.CampaignTarget,
			ID:     target.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.CampaignDeleteOne]{
			Client: suite.Client.DB.Campaign,
			ID:     campaignObj.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.EmailTemplateDeleteOne]{
			Client: suite.Client.DB.EmailTemplate,
			ID:     emailTemplate.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	mockSender, err := mock.New("")
	assert.NilError(t, err)

	emailClient := &email.Client{
		Sender: mockSender,
		Config: email.RuntimeEmailConfig{
			FromEmail:   "noreply@test.example",
			CompanyName: "NoBrandCo",
			ProductURL:  "https://app.example.com",
		},
	}

	cfg := email.SendBrandedCampaignRequest{CampaignDispatchInput: email.CampaignDispatchInput{CampaignID: campaignObj.ID}}
	req := types.OperationRequest{
		Client: emailClient,
		DB:     suite.Client.DB,
	}

	configBytes, err := json.Marshal(cfg)
	assert.NilError(t, err)
	req.Config = configBytes

	_, err = email.SendBrandedCampaign{}.Run(ctx, req, emailClient, cfg)
	assert.NilError(t, err)

	messages := mockSender.Messages()
	assert.Assert(t, is.Len(messages, 1))
	assert.Assert(t, strings.Contains(messages[0].Subject, "Hello Charlie"))
	assert.Assert(t, strings.Contains(strings.Join(messages[0].To, " "), "charlie@test.example"))
}

// TestCampaignEmailDispatchNoTemplate verifies dispatch is a no-op
// when no email template is linked to the campaign
func TestCampaignEmailDispatchNoTemplate(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	campaignObj, err := suite.Client.DB.Campaign.Create().
		SetName("No Template Campaign").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetRecurrenceFrequency(enums.FrequencyNone).
		Save(ctx)
	th.RequireNoError(t, err)

	target, err := suite.Client.DB.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("nobody@test.example").
		SetFullName("No Body").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*generated.CampaignTargetDeleteOne]{
			Client: suite.Client.DB.CampaignTarget,
			ID:     target.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.CampaignDeleteOne]{
			Client: suite.Client.DB.Campaign,
			ID:     campaignObj.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	mockSender, err := mock.New("")
	assert.NilError(t, err)

	emailClient := &email.Client{
		Sender: mockSender,
		Config: email.RuntimeEmailConfig{
			FromEmail:   "noreply@test.example",
			CompanyName: "TestCorp",
			ProductURL:  "https://app.example.com",
		},
	}

	cfg := email.SendBrandedCampaignRequest{CampaignDispatchInput: email.CampaignDispatchInput{CampaignID: campaignObj.ID}}
	req := types.OperationRequest{
		Client: emailClient,
		DB:     suite.Client.DB,
	}

	configBytes, err := json.Marshal(cfg)
	assert.NilError(t, err)
	req.Config = configBytes

	_, err = email.SendBrandedCampaign{}.Run(ctx, req, emailClient, cfg)
	assert.ErrorIs(t, err, email.ErrDispatcherNotFound)

	messages := mockSender.Messages()
	assert.Assert(t, is.Len(messages, 0))
}

// TestQuestionnaireTestEmailDispatch verifies the questionnaire test-send
// operation creates a test assessment response and sends one auth email.
func TestQuestionnaireTestEmailDispatch(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	assessmentObj := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	campaignObj, err := suite.Client.DB.Campaign.Create().
		SetName("Questionnaire Test Send Campaign").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetCampaignType(enums.CampaignTypeQuestionnaire).
		SetAssessmentID(assessmentObj.ID).
		SetRecurrenceFrequency(enums.FrequencyNone).
		Save(ctx)
	th.RequireNoError(t, err)

	var responseIDs []string
	t.Cleanup(func() {
		if len(responseIDs) > 0 {
			(&th.Cleanup[*generated.AssessmentResponseDeleteOne]{
				Client: suite.Client.DB.AssessmentResponse,
				IDs:    responseIDs,
			}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}
		(&th.Cleanup[*generated.CampaignDeleteOne]{
			Client: suite.Client.DB.Campaign,
			ID:     campaignObj.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.AssessmentDeleteOne]{
			Client: suite.Client.DB.Assessment,
			ID:     assessmentObj.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.TemplateDeleteOne]{
			Client: suite.Client.DB.Template,
			ID:     assessmentObj.TemplateID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	mockSender, err := mock.New("")
	assert.NilError(t, err)

	emailClient := &email.Client{
		Sender: mockSender,
		Config: email.RuntimeEmailConfig{
			FromEmail:   "questionnaire@test.example",
			CompanyName: "QuestionnaireCo",
			ProductURL:  "https://app.example.com",
		},
	}

	recipient := "test-recipient@test.example"
	cfg := email.SendQuestionnaireCampaignRequest{
		CampaignDispatchInput: email.CampaignDispatchInput{CampaignID: campaignObj.ID},
		TestEmail:             recipient,
	}
	req := types.OperationRequest{
		Client: emailClient,
		DB:     suite.Client.DB,
	}

	configBytes, err := json.Marshal(cfg)
	assert.NilError(t, err)
	req.Config = configBytes

	_, err = email.SendQuestionnaireCampaign{}.Run(ctx, req, emailClient, cfg)
	assert.NilError(t, err)

	messages := mockSender.Messages()
	assert.Assert(t, is.Len(messages, 1))
	assert.Assert(t, strings.Contains(strings.Join(messages[0].To, " "), recipient))

	count, err := suite.Client.DB.AssessmentResponse.Query().
		Where(
			assessmentresponse.CampaignIDEQ(campaignObj.ID),
			assessmentresponse.EmailEqualFold(recipient),
			assessmentresponse.IsTest(true),
		).
		Count(ctx)
	assert.NilError(t, err)
	assert.Equal(t, count, 1)

	responseIDs, err = suite.Client.DB.AssessmentResponse.Query().
		Where(
			assessmentresponse.CampaignIDEQ(campaignObj.ID),
			assessmentresponse.EmailEqualFold(recipient),
			assessmentresponse.IsTest(true),
		).
		IDs(ctx)
	assert.NilError(t, err)
}

// TestCustomCampaignCompletesWhenAllSent verifies a custom campaign is marked completed
// once every target has been emailed, and that other campaign types are left alone
func TestCustomCampaignCompletesWhenAllSent(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	emailTemplate := (&th.EmailTemplateBuilder{Client: suite.Client}).MustNew(ctx, t)

	t.Cleanup(func() {
		(&th.Cleanup[*generated.EmailTemplateDeleteOne]{
			Client: suite.Client.DB.EmailTemplate,
			ID:     emailTemplate.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	testCases := []struct {
		name                string
		campaignType        enums.CampaignType
		isRecurring         bool
		withAssessment      bool
		skippedTargetStatus *enums.AssessmentResponseStatus
		expectedStatus      enums.CampaignStatus
	}{
		{
			name:           "custom campaign completes once all targets are sent",
			campaignType:   enums.CampaignTypeCustom,
			expectedStatus: enums.CampaignStatusCompleted,
		},
		{
			name:                "custom campaign stays active while a target is unsent",
			campaignType:        enums.CampaignTypeCustom,
			skippedTargetStatus: &enums.AssessmentResponseStatusCompleted,
			expectedStatus:      enums.CampaignStatusActive,
		},
		{
			name:           "non custom campaign is not completed by sending",
			campaignType:   enums.CampaignTypeTraining,
			expectedStatus: enums.CampaignStatusActive,
		},
		{
			name:           "recurring custom campaign is not completed by sending",
			campaignType:   enums.CampaignTypeCustom,
			isRecurring:    true,
			expectedStatus: enums.CampaignStatusActive,
		},
		{
			name:           "custom campaign with a linked assessment is not completed by sending",
			campaignType:   enums.CampaignTypeCustom,
			withAssessment: true,
			expectedStatus: enums.CampaignStatusActive,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			create := suite.Client.DB.Campaign.Create().
				SetName(th.RandomName(t)).
				SetOwnerID(th.SharedTestUser1.OrganizationID).
				SetCampaignType(tc.campaignType).
				SetStatus(enums.CampaignStatusActive).
				SetIsActive(true).
				SetIsRecurring(tc.isRecurring).
				SetEmailTemplateID(emailTemplate.ID).
				SetRecurrenceFrequency(enums.FrequencyNone)

			if tc.withAssessment {
				assessmentObj := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(ctx, t)

				t.Cleanup(func() {
					(&th.Cleanup[*generated.AssessmentDeleteOne]{
						Client: suite.Client.DB.Assessment,
						ID:     assessmentObj.ID,
					}).MustDelete(th.SharedTestUser1.UserCtx, t)
				})

				create.SetAssessmentID(assessmentObj.ID)
			}

			campaignObj, err := create.Save(ctx)
			th.RequireNoError(t, err)

			target, err := suite.Client.DB.CampaignTarget.Create().
				SetCampaignID(campaignObj.ID).
				SetEmail("target@test.example").
				SetOwnerID(th.SharedTestUser1.OrganizationID).
				Save(ctx)
			th.RequireNoError(t, err)

			targetIDs := []string{target.ID}

			if tc.skippedTargetStatus != nil {
				skipped, err := suite.Client.DB.CampaignTarget.Create().
					SetCampaignID(campaignObj.ID).
					SetEmail("skipped@test.example").
					SetOwnerID(th.SharedTestUser1.OrganizationID).
					SetStatus(*tc.skippedTargetStatus).
					Save(ctx)
				th.RequireNoError(t, err)

				targetIDs = append(targetIDs, skipped.ID)
			}

			t.Cleanup(func() {
				(&th.Cleanup[*generated.CampaignTargetDeleteOne]{
					Client: suite.Client.DB.CampaignTarget,
					IDs:    targetIDs,
				}).MustDelete(th.SharedTestUser1.UserCtx, t)
				(&th.Cleanup[*generated.CampaignDeleteOne]{
					Client: suite.Client.DB.Campaign,
					ID:     campaignObj.ID,
				}).MustDelete(th.SharedTestUser1.UserCtx, t)
			})

			mockSender, err := mock.New("")
			assert.NilError(t, err)

			emailClient := &email.Client{
				Sender: mockSender,
				Config: email.RuntimeEmailConfig{
					FromEmail:   "test@mail.example.com",
					CompanyName: "TestCorp",
					ProductURL:  "https://app.example.com",
				},
			}

			cfg := email.SendBrandedCampaignRequest{CampaignDispatchInput: email.CampaignDispatchInput{CampaignID: campaignObj.ID}}

			_, err = email.SendBrandedCampaign{}.Run(ctx, types.OperationRequest{Client: emailClient, DB: suite.Client.DB}, emailClient, cfg)
			assert.NilError(t, err)

			completed := tc.expectedStatus == enums.CampaignStatusCompleted

			updated, err := suite.Client.DB.Campaign.Get(ctx, campaignObj.ID)
			assert.NilError(t, err)
			assert.Equal(t, updated.Status, tc.expectedStatus)
			assert.Equal(t, updated.IsActive, !completed)
			assert.Equal(t, updated.CompletedAt != nil, completed)
		})
	}
}
