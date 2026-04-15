//go:build test

package graphapi_test

import (
	"strings"
	"testing"
	"time"

	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/theopenlane/riverboat/pkg/jobs"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/definitions/email"
)

// TestCampaignEmailDispatch verifies that SendCampaignEmails renders
// campaign emails with the correct branding, template variables, and
// metadata, then enqueues one river job per target
func TestCampaignEmailDispatch(t *testing.T) {
	ctx := setContext(testUser1.UserCtx, suite.client.db)

	// clear any pending jobs
	err := suite.client.db.Job.TruncateRiverTables(ctx)
	assert.NilError(t, err)

	// --- fixtures ---

	emailTemplate := suite.client.db.EmailTemplate.Create().
		SetName("Campaign Dispatch Test Template").
		SetKey("campaign-dispatch-test").
		SetTemplateContext(enums.TemplateContextCampaignRecipient).
		SetSubjectTemplate("Hello {{ .recipientFirstName }}").
		SetBodyTemplate("<h1>Welcome {{ .recipientFirstName }}</h1><p>Campaign: {{ .campaignName }}</p><p>{{ .promoCode }}</p>").
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

	// cleanup at end
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

	// --- dispatch ---

	emailClient := &email.EmailClient{
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

	err = email.SendCampaignEmails(ctx, suite.client.db, emailClient, campaignObj.ID)
	assert.NilError(t, err)

	// --- verify river jobs ---

	insertedJobs := rivertest.RequireManyInserted(ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()),
		[]rivertest.ExpectedJob{
			{Args: jobs.EmailArgs{}},
			{Args: jobs.EmailArgs{}},
		})
	assert.Assert(t, is.Len(insertedJobs, 2))

	// collect encoded args for content verification
	var allArgs []string
	for _, j := range insertedJobs {
		allArgs = append(allArgs, string(j.EncodedArgs))
	}

	combined := strings.Join(allArgs, "\n")

	t.Run("subject contains recipient first name", func(t *testing.T) {
		assert.Assert(t, strings.Contains(combined, "Hello Alice") || strings.Contains(combined, "Hello Bob"),
			"expected subject with first name, got: %s", combined)
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
		assert.Assert(t, strings.Contains(combined, "#ff5500"),
			"expected button color in rendered HTML")
		assert.Assert(t, strings.Contains(combined, "#f0f0f0"),
			"expected background color in rendered HTML")
		assert.Assert(t, strings.Contains(combined, "#333333"),
			"expected text color in rendered HTML")
	})

	t.Run("brand name in rendered output", func(t *testing.T) {
		assert.Assert(t, strings.Contains(combined, "TestBrand"),
			"expected brand name in rendered HTML")
	})

	t.Run("each target gets its own job", func(t *testing.T) {
		aliceFound := false
		bobFound := false

		for _, args := range allArgs {
			if strings.Contains(args, "alice@test.example") {
				aliceFound = true
			}

			if strings.Contains(args, "bob@test.example") {
				bobFound = true
			}
		}

		assert.Assert(t, aliceFound, "expected job for alice@test.example")
		assert.Assert(t, bobFound, "expected job for bob@test.example")
	})

	t.Run("campaign target tag present", func(t *testing.T) {
		assert.Assert(t, strings.Contains(combined, email.TagCampaignTargetID),
			"expected campaign_target_id tag in email args")
	})

	t.Run("from address matches config", func(t *testing.T) {
		assert.Assert(t, strings.Contains(combined, "test@mail.example.com"),
			"expected from address in email args")
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

	err := suite.client.db.Job.TruncateRiverTables(ctx)
	assert.NilError(t, err)

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

	// target already sent
	sentTarget := suite.client.db.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("already-sent@test.example").
		SetFullName("Already Sent").
		SetOwnerID(testUser1.OrganizationID).
		SaveX(ctx)

	// mark as sent
	sentAt := models.DateTime(time.Now())
	suite.client.db.CampaignTarget.UpdateOneID(sentTarget.ID).
		SetSentAt(sentAt).
		SaveX(ctx)

	// unsent target
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

	emailClient := &email.EmailClient{
		Config: email.RuntimeEmailConfig{
			FromEmail:   "test@mail.example.com",
			CompanyName: "TestCorp",
			ProductURL:  "https://app.example.com",
		},
	}

	err = email.SendCampaignEmails(ctx, suite.client.db, emailClient, campaignObj.ID)
	assert.NilError(t, err)

	insertedJobs := rivertest.RequireManyInserted(ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()),
		[]rivertest.ExpectedJob{
			{Args: jobs.EmailArgs{}},
		})
	assert.Assert(t, is.Len(insertedJobs, 1))

	assert.Assert(t, strings.Contains(string(insertedJobs[0].EncodedArgs), "unsent@test.example"),
		"expected only unsent target to receive email")
	assert.Assert(t, !strings.Contains(string(insertedJobs[0].EncodedArgs), "already-sent@test.example"),
		"already-sent target should not receive email")
}

// TestCampaignEmailDispatchNoBranding verifies dispatch works without
// an EmailBranding record attached to the campaign
func TestCampaignEmailDispatchNoBranding(t *testing.T) {
	ctx := setContext(testUser1.UserCtx, suite.client.db)

	err := suite.client.db.Job.TruncateRiverTables(ctx)
	assert.NilError(t, err)

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

	emailClient := &email.EmailClient{
		Config: email.RuntimeEmailConfig{
			FromEmail:   "noreply@test.example",
			CompanyName: "NoBrandCo",
			ProductURL:  "https://app.example.com",
		},
	}

	err = email.SendCampaignEmails(ctx, suite.client.db, emailClient, campaignObj.ID)
	assert.NilError(t, err)

	insertedJobs := rivertest.RequireManyInserted(ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()),
		[]rivertest.ExpectedJob{
			{Args: jobs.EmailArgs{}},
		})
	assert.Assert(t, is.Len(insertedJobs, 1))

	args := string(insertedJobs[0].EncodedArgs)
	assert.Assert(t, strings.Contains(args, "Hello Charlie"), "expected rendered subject")
	assert.Assert(t, strings.Contains(args, "charlie@test.example"), "expected target email")
}

// TestCampaignEmailDispatchNoTemplate verifies dispatch is a no-op
// when no email template is linked to the campaign
func TestCampaignEmailDispatchNoTemplate(t *testing.T) {
	ctx := setContext(testUser1.UserCtx, suite.client.db)

	err := suite.client.db.Job.TruncateRiverTables(ctx)
	assert.NilError(t, err)

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

	emailClient := &email.EmailClient{
		Config: email.RuntimeEmailConfig{
			FromEmail:   "noreply@test.example",
			CompanyName: "TestCorp",
			ProductURL:  "https://app.example.com",
		},
	}

	err = email.SendCampaignEmails(ctx, suite.client.db, emailClient, campaignObj.ID)
	assert.NilError(t, err)

	rivertest.RequireNotInserted(ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()), &jobs.EmailArgs{}, nil)
}
