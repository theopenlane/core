//go:build test

package graphapi_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/newman/providers/mock"
)

// TestCampaignEmailDispatch verifies that SendCampaign renders
// campaign emails with the correct branding, template variables, and
// metadata, then sends one email per target via the mock sender
func TestCampaignEmailDispatch(t *testing.T) {
	ctx := setContext(testUser1.UserCtx, suite.client.db)

	// --- fixtures ---

	emailTemplate := suite.client.db.EmailTemplate.Create().
		SetName("Campaign Dispatch Test Template").
		SetKey("campaign-dispatch-test").
		SetTemplateContext(enums.TemplateContextCampaignRecipient).
		SetSubjectTemplate("Hello {{ .recipientFirstName }}").
		SetBodyTemplate("# Welcome {{ .recipientFirstName }}\n\nCampaign: {{ .campaignName }}\n\n{{ .promoCode }}").
		SetTextTemplate("Welcome {{ .recipientFirstName }} - Campaign: {{ .campaignName }}").
		SetDefaults(map[string]any{
			"promoCode": "DEFAULT123",
		}).
		SaveX(ctx)

	emailBranding := suite.client.db.EmailBranding.Create().
		SetName("Dispatch Test Branding").
		SetBrandName("TestBrand").
		SetButtonColor("#ff5500").
		SetButtonTextColor("#ffffff").
		SetBackgroundColor("#f0f0f0").
		SetTextColor("#333333").
		SetLinkColor("#0066cc").
		SaveX(ctx)

	campaignObj := suite.client.db.Campaign.Create().
		SetName("Dispatch Integration Test Campaign").
		SetDescription("Testing email dispatch pipeline").
		SetOwnerID(testUser1.OrganizationID).
		SetEmailTemplateID(emailTemplate.ID).
		SetEmailBrandingID(emailBranding.ID).
		SetRecurrenceFrequency(enums.FrequencyNone).
		SetMetadata(map[string]any{
			"promoCode": "SUMMER2025",
		}).
		SaveX(ctx)

	targetAlice := suite.client.db.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("alice@test.example").
		SetFullName("Alice Smith").
		SetOwnerID(testUser1.OrganizationID).
		SaveX(ctx)

	targetBob := suite.client.db.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("bob@test.example").
		SetFullName("Bob Jones").
		SetOwnerID(testUser1.OrganizationID).
		SaveX(ctx)

	defer func() {
		(&Cleanup[*generated.CampaignTargetDeleteOne]{
			client: suite.client.db.CampaignTarget,
			IDs:    []string{targetAlice.ID, targetBob.ID},
		}).MustDelete(testUser1.UserCtx, t)
		(&Cleanup[*generated.CampaignDeleteOne]{
			client: suite.client.db.Campaign,
			ID:     campaignObj.ID,
		}).MustDelete(testUser1.UserCtx, t)
		(&Cleanup[*generated.EmailTemplateDeleteOne]{
			client: suite.client.db.EmailTemplate,
			ID:     emailTemplate.ID,
		}).MustDelete(testUser1.UserCtx, t)
		(&Cleanup[*generated.EmailBrandingDeleteOne]{
			client: suite.client.db.EmailBranding,
			ID:     emailBranding.ID,
		}).MustDelete(testUser1.UserCtx, t)
	}()

	// --- dispatch via SendCampaign operation ---

	mockSender, err := mock.New("")
	assert.NilError(t, err)

	emailClient := &email.EmailClient{
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

	cfg := email.SendCampaignRequest{CampaignID: campaignObj.ID}
	req := types.OperationRequest{
		Client: emailClient,
		DB:     suite.client.db,
	}

	configBytes, err := json.Marshal(cfg)
	assert.NilError(t, err)
	req.Config = configBytes

	_, err = email.SendCampaign{}.Run(ctx, req, emailClient, cfg)
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
		assert.Assert(t, strings.Contains(combined, "SUMMER2025"),
			"expected metadata promoCode to override default")
		assert.Assert(t, !strings.Contains(combined, "DEFAULT123"),
			"default promoCode should be overridden by metadata")
	})

	t.Run("branding colors applied to HTML", func(t *testing.T) {
		// FreeMarkdown templates have no .button elements, so button colors
		// are not inlined. Background and text colors are applied to wrapper
		// and paragraph elements that exist in the rendered structure
		assert.Assert(t, strings.Contains(combined, "#f0f0f0"),
			"expected background color in rendered HTML")
		assert.Assert(t, strings.Contains(combined, "#333333"),
			"expected text color in rendered HTML")
	})

	t.Run("brand name in rendered output", func(t *testing.T) {
		assert.Assert(t, strings.Contains(combined, "TestBrand"),
			"expected brand name in rendered HTML")
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
		updatedAlice, err := suite.client.db.CampaignTarget.Get(ctx, targetAlice.ID)
		assert.NilError(t, err)
		assert.Assert(t, updatedAlice.SentAt != nil, "expected sent_at to be set for alice")

		updatedBob, err := suite.client.db.CampaignTarget.Get(ctx, targetBob.ID)
		assert.NilError(t, err)
		assert.Assert(t, updatedBob.SentAt != nil, "expected sent_at to be set for bob")
	})
}

// TestCampaignEmailDispatchSkipsSentTargets verifies that targets with
// sent_at already set are not re-dispatched
func TestCampaignEmailDispatchSkipsSentTargets(t *testing.T) {
	ctx := setContext(testUser1.UserCtx, suite.client.db)

	emailTemplate := suite.client.db.EmailTemplate.Create().
		SetName("Skip Sent Test Template").
		SetKey("skip-sent-test").
		SetTemplateContext(enums.TemplateContextCampaignRecipient).
		SetSubjectTemplate("Test").
		SetBodyTemplate("<p>Test</p>").
		SaveX(ctx)

	campaignObj := suite.client.db.Campaign.Create().
		SetName("Skip Sent Test Campaign").
		SetOwnerID(testUser1.OrganizationID).
		SetEmailTemplateID(emailTemplate.ID).
		SetRecurrenceFrequency(enums.FrequencyNone).
		SaveX(ctx)

	sentTarget := suite.client.db.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("already-sent@test.example").
		SetFullName("Already Sent").
		SetOwnerID(testUser1.OrganizationID).
		SaveX(ctx)

	sentAt := models.DateTime(time.Now())
	suite.client.db.CampaignTarget.UpdateOneID(sentTarget.ID).
		SetSentAt(sentAt).
		SaveX(ctx)

	unsentTarget := suite.client.db.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("unsent@test.example").
		SetFullName("Unsent Target").
		SetOwnerID(testUser1.OrganizationID).
		SaveX(ctx)

	defer func() {
		(&Cleanup[*generated.CampaignTargetDeleteOne]{
			client: suite.client.db.CampaignTarget,
			IDs:    []string{sentTarget.ID, unsentTarget.ID},
		}).MustDelete(testUser1.UserCtx, t)
		(&Cleanup[*generated.CampaignDeleteOne]{
			client: suite.client.db.Campaign,
			ID:     campaignObj.ID,
		}).MustDelete(testUser1.UserCtx, t)
		(&Cleanup[*generated.EmailTemplateDeleteOne]{
			client: suite.client.db.EmailTemplate,
			ID:     emailTemplate.ID,
		}).MustDelete(testUser1.UserCtx, t)
	}()

	mockSender, err := mock.New("")
	assert.NilError(t, err)

	emailClient := &email.EmailClient{
		Sender: mockSender,
		Config: email.RuntimeEmailConfig{
			FromEmail:   "test@mail.example.com",
			CompanyName: "TestCorp",
			ProductURL:  "https://app.example.com",
		},
	}

	cfg := email.SendCampaignRequest{CampaignID: campaignObj.ID}
	req := types.OperationRequest{
		Client: emailClient,
		DB:     suite.client.db,
	}

	configBytes, err := json.Marshal(cfg)
	assert.NilError(t, err)
	req.Config = configBytes

	_, err = email.SendCampaign{}.Run(ctx, req, emailClient, cfg)
	assert.NilError(t, err)

	messages := mockSender.Messages()
	assert.Assert(t, is.Len(messages, 1))
	assert.Assert(t, strings.Contains(strings.Join(messages[0].To, " "), "unsent@test.example"))
}

// TestCampaignEmailDispatchNoBranding verifies dispatch works without
// an EmailBranding record attached to the campaign
func TestCampaignEmailDispatchNoBranding(t *testing.T) {
	ctx := setContext(testUser1.UserCtx, suite.client.db)

	emailTemplate := suite.client.db.EmailTemplate.Create().
		SetName("No Branding Test Template").
		SetKey("no-branding-test").
		SetTemplateContext(enums.TemplateContextCampaignRecipient).
		SetSubjectTemplate("Hello {{ .recipientFirstName }}").
		SetBodyTemplate("<p>Welcome {{ .recipientFirstName }}</p>").
		SaveX(ctx)

	campaignObj := suite.client.db.Campaign.Create().
		SetName("No Branding Campaign").
		SetOwnerID(testUser1.OrganizationID).
		SetEmailTemplateID(emailTemplate.ID).
		SetRecurrenceFrequency(enums.FrequencyNone).
		SaveX(ctx)

	target := suite.client.db.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("charlie@test.example").
		SetFullName("Charlie Brown").
		SetOwnerID(testUser1.OrganizationID).
		SaveX(ctx)

	defer func() {
		(&Cleanup[*generated.CampaignTargetDeleteOne]{
			client: suite.client.db.CampaignTarget,
			ID:     target.ID,
		}).MustDelete(testUser1.UserCtx, t)
		(&Cleanup[*generated.CampaignDeleteOne]{
			client: suite.client.db.Campaign,
			ID:     campaignObj.ID,
		}).MustDelete(testUser1.UserCtx, t)
		(&Cleanup[*generated.EmailTemplateDeleteOne]{
			client: suite.client.db.EmailTemplate,
			ID:     emailTemplate.ID,
		}).MustDelete(testUser1.UserCtx, t)
	}()

	mockSender, err := mock.New("")
	assert.NilError(t, err)

	emailClient := &email.EmailClient{
		Sender: mockSender,
		Config: email.RuntimeEmailConfig{
			FromEmail:   "noreply@test.example",
			CompanyName: "NoBrandCo",
			ProductURL:  "https://app.example.com",
		},
	}

	cfg := email.SendCampaignRequest{CampaignID: campaignObj.ID}
	req := types.OperationRequest{
		Client: emailClient,
		DB:     suite.client.db,
	}

	configBytes, err := json.Marshal(cfg)
	assert.NilError(t, err)
	req.Config = configBytes

	_, err = email.SendCampaign{}.Run(ctx, req, emailClient, cfg)
	assert.NilError(t, err)

	messages := mockSender.Messages()
	assert.Assert(t, is.Len(messages, 1))
	assert.Assert(t, strings.Contains(messages[0].Subject, "Hello Charlie"))
	assert.Assert(t, strings.Contains(strings.Join(messages[0].To, " "), "charlie@test.example"))
}

// TestCampaignEmailDispatchNoTemplate verifies dispatch is a no-op
// when no email template is linked to the campaign
func TestCampaignEmailDispatchNoTemplate(t *testing.T) {
	ctx := setContext(testUser1.UserCtx, suite.client.db)

	campaignObj := suite.client.db.Campaign.Create().
		SetName("No Template Campaign").
		SetOwnerID(testUser1.OrganizationID).
		SetRecurrenceFrequency(enums.FrequencyNone).
		SaveX(ctx)

	target := suite.client.db.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("nobody@test.example").
		SetFullName("No Body").
		SetOwnerID(testUser1.OrganizationID).
		SaveX(ctx)

	defer func() {
		(&Cleanup[*generated.CampaignTargetDeleteOne]{
			client: suite.client.db.CampaignTarget,
			ID:     target.ID,
		}).MustDelete(testUser1.UserCtx, t)
		(&Cleanup[*generated.CampaignDeleteOne]{
			client: suite.client.db.Campaign,
			ID:     campaignObj.ID,
		}).MustDelete(testUser1.UserCtx, t)
	}()

	mockSender, err := mock.New("")
	assert.NilError(t, err)

	emailClient := &email.EmailClient{
		Sender: mockSender,
		Config: email.RuntimeEmailConfig{
			FromEmail:   "noreply@test.example",
			CompanyName: "TestCorp",
			ProductURL:  "https://app.example.com",
		},
	}

	cfg := email.SendCampaignRequest{CampaignID: campaignObj.ID}
	req := types.OperationRequest{
		Client: emailClient,
		DB:     suite.client.db,
	}

	configBytes, err := json.Marshal(cfg)
	assert.NilError(t, err)
	req.Config = configBytes

	_, err = email.SendCampaign{}.Run(ctx, req, emailClient, cfg)
	assert.NilError(t, err)

	messages := mockSender.Messages()
	assert.Assert(t, is.Len(messages, 0))
}
