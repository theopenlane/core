//go:build test

package graphapi_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/samber/lo"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/internal/integrations/definitions/email"
)

// createBrandedTestCampaign creates an email template and a branded (non-questionnaire) campaign
// owned by sharedTestUser1, registering cleanup for both
func createBrandedTestCampaign(t *testing.T, name string) *generated.Campaign {
	t.Helper()

	ctx := setContext(sharedTestUser1.UserCtx, suite.client.db)

	emailTemplate := suite.client.db.EmailTemplate.Create().
		SetName(name + " Template").
		SetKey(email.BrandedMessageOp.Name()).
		SetTemplateContext(enums.TemplateContextCampaignRecipient).
		SetDefaults(map[string]any{
			"subject": "Test",
			"title":   "Test",
			"intros":  []any{"Test"},
		}).
		SaveX(ctx)

	campaignObj := suite.client.db.Campaign.Create().
		SetName(name).
		SetOwnerID(sharedTestUser1.OrganizationID).
		SetCampaignType(enums.CampaignTypeCustom).
		SetEmailTemplateID(emailTemplate.ID).
		SetRecurrenceFrequency(enums.FrequencyNone).
		SaveX(ctx)

	t.Cleanup(func() {
		(&Cleanup[*generated.CampaignDeleteOne]{client: suite.client.db.Campaign, ID: campaignObj.ID}).MustDelete(sharedTestUser1.UserCtx, t)
		(&Cleanup[*generated.EmailTemplateDeleteOne]{client: suite.client.db.EmailTemplate, ID: emailTemplate.ID}).MustDelete(sharedTestUser1.UserCtx, t)
	})

	return campaignObj
}

// sendTestEmail invokes the sendCampaignTestEmail mutation as sharedTestUser1
func sendTestEmail(campaignID string, emails []string) (*testclient.SendCampaignTestEmail, error) {
	return suite.client.api.SendCampaignTestEmail(sharedTestUser1.UserCtx, testclient.SendCampaignTestEmailInput{
		CampaignID: campaignID,
		Emails:     emails,
	})
}

// TestSendCampaignTestEmailAssessmentBackfill verifies a questionnaire campaign created with only a
// questionnaire template reference gets an assessment created from the template and linked on the
// first test email send, and that subsequent sends reuse the same assessment
func TestSendCampaignTestEmailAssessmentBackfill(t *testing.T) {
	template := (&TemplateBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)

	createResp, err := suite.client.api.CreateCampaign(sharedTestUser1.UserCtx, testclient.CreateCampaignInput{
		Name:                fmt.Sprintf("questionnaire-backfill-%s", ulids.New().String()),
		CampaignType:        lo.ToPtr(enums.CampaignTypeQuestionnaire),
		TemplateID:          lo.ToPtr(template.ID),
		RecurrenceFrequency: lo.ToPtr(enums.FrequencyNone),
	})
	assert.NilError(t, err)

	campaignID := createResp.CreateCampaign.Campaign.ID
	assert.Check(t, lo.FromPtr(createResp.CreateCampaign.Campaign.AssessmentID) == "", "campaign should start without an assessment")

	var assessmentID string

	t.Cleanup(func() {
		allowCtx := setContext(sharedTestUser1.UserCtx, suite.client.db)

		responses := suite.client.db.AssessmentResponse.Query().
			Where(assessmentresponse.CampaignIDEQ(campaignID)).
			AllX(allowCtx)
		if len(responses) > 0 {
			(&Cleanup[*generated.AssessmentResponseDeleteOne]{
				client: suite.client.db.AssessmentResponse,
				IDs:    lo.Map(responses, func(r *generated.AssessmentResponse, _ int) string { return r.ID }),
			}).MustDelete(sharedTestUser1.UserCtx, t)
		}

		(&Cleanup[*generated.CampaignDeleteOne]{client: suite.client.db.Campaign, ID: campaignID}).MustDelete(sharedTestUser1.UserCtx, t)

		if assessmentID != "" {
			(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, ID: assessmentID}).MustDelete(sharedTestUser1.UserCtx, t)
		}

		(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: template.ID}).MustDelete(sharedTestUser1.UserCtx, t)
	})

	sendResp, err := sendTestEmail(campaignID, []string{"backfill-recipient@test.example"})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(int64(1), sendResp.SendCampaignTestEmail.QueuedCount))
	assert.Check(t, is.Equal(int64(0), sendResp.SendCampaignTestEmail.SkippedCount))

	assessmentID = lo.FromPtr(sendResp.SendCampaignTestEmail.Campaign.AssessmentID)
	assert.Assert(t, assessmentID != "", "expected assessment to be backfilled from the questionnaire template")

	ctx := setContext(sharedTestUser1.UserCtx, suite.client.db)
	assessmentObj := suite.client.db.Assessment.GetX(ctx, assessmentID)
	assert.Check(t, is.Equal(template.ID, assessmentObj.TemplateID))
	assert.Check(t, is.Equal(template.Name, assessmentObj.Name))

	// compare the inherited jsonconfig through a marshal round trip to normalize value types
	wantConfig, err := json.Marshal(template.Jsonconfig)
	assert.NilError(t, err)
	gotConfig, err := json.Marshal(assessmentObj.Jsonconfig)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(string(wantConfig), string(gotConfig)))

	// a second send must reuse the backfilled assessment rather than creating another
	secondResp, err := sendTestEmail(campaignID, []string{"backfill-second@test.example"})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(assessmentID, lo.FromPtr(secondResp.SendCampaignTestEmail.Campaign.AssessmentID)))
}

// TestSendCampaignTestEmailBrandedCampaign verifies non-questionnaire campaigns with a linked email
// template can send test emails, and that recipients are trimmed and deduped case-insensitively
func TestSendCampaignTestEmailBrandedCampaign(t *testing.T) {
	campaignObj := createBrandedTestCampaign(t, "Branded Test Email Campaign")

	resp, err := sendTestEmail(campaignObj.ID, []string{"Dup@Test.Example", "dup@test.example", "   ", "unique@test.example"})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(int64(2), resp.SendCampaignTestEmail.QueuedCount))
	assert.Check(t, is.Equal(int64(2), resp.SendCampaignTestEmail.SkippedCount))
}

// TestSendCampaignTestEmailMissingEmailTemplate verifies a non-questionnaire campaign without a
// linked email template is rejected
func TestSendCampaignTestEmailMissingEmailTemplate(t *testing.T) {
	ctx := setContext(sharedTestUser1.UserCtx, suite.client.db)

	campaignObj := suite.client.db.Campaign.Create().
		SetName("Missing Email Template Test Campaign").
		SetOwnerID(sharedTestUser1.OrganizationID).
		SetCampaignType(enums.CampaignTypeCustom).
		SetRecurrenceFrequency(enums.FrequencyNone).
		SaveX(ctx)

	t.Cleanup(func() {
		(&Cleanup[*generated.CampaignDeleteOne]{client: suite.client.db.Campaign, ID: campaignObj.ID}).MustDelete(sharedTestUser1.UserCtx, t)
	})

	_, err := sendTestEmail(campaignObj.ID, []string{"missing-template@test.example"})
	assert.Assert(t, err != nil)
	assert.Assert(t, strings.Contains(err.Error(), "email template"), "expected missing email template error, got: %v", err)
}

// TestSendCampaignTestEmailRecipientCap verifies requests with more than the recipient cap are rejected
func TestSendCampaignTestEmailRecipientCap(t *testing.T) {
	campaignObj := createBrandedTestCampaign(t, "Recipient Cap Campaign")

	emails := make([]string, 6)
	for i := range emails {
		emails[i] = fmt.Sprintf("cap-%d@test.example", i)
	}

	_, err := sendTestEmail(campaignObj.ID, emails)
	assert.Assert(t, err != nil)
	assert.Assert(t, strings.Contains(err.Error(), "5 recipients"), "expected recipient cap error, got: %v", err)
}

// TestSendCampaignTestEmailRateLimit verifies the per-campaign hourly budget is enforced and scoped
// to a single campaign
func TestSendCampaignTestEmailRateLimit(t *testing.T) {
	campaignObj := createBrandedTestCampaign(t, "Rate Limit Campaign")

	batch := func(prefix string) []string {
		emails := make([]string, 5)
		for i := range emails {
			emails[i] = fmt.Sprintf("%s-%d@test.example", prefix, i)
		}

		return emails
	}

	resp, err := sendTestEmail(campaignObj.ID, batch("first"))
	assert.NilError(t, err)
	assert.Check(t, is.Equal(int64(5), resp.SendCampaignTestEmail.QueuedCount))

	resp, err = sendTestEmail(campaignObj.ID, batch("second"))
	assert.NilError(t, err)
	assert.Check(t, is.Equal(int64(5), resp.SendCampaignTestEmail.QueuedCount))

	_, err = sendTestEmail(campaignObj.ID, []string{"over-limit@test.example"})
	assert.Assert(t, err != nil)
	assert.Assert(t, strings.Contains(err.Error(), "hourly limit"), "expected rate limit error, got: %v", err)

	// a different campaign has its own hourly budget
	otherCampaign := createBrandedTestCampaign(t, "Rate Limit Campaign Two")

	otherResp, err := sendTestEmail(otherCampaign.ID, []string{"other-campaign@test.example"})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(int64(1), otherResp.SendCampaignTestEmail.QueuedCount))
}
