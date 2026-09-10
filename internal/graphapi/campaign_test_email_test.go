//go:build test

package graphapi_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/email"
)

// createBrandedTestCampaign creates an email template and a branded (non-questionnaire) campaign
// owned by th.SharedTestUser1, registering cleanup for both
func createBrandedTestCampaign(t *testing.T, name string) *generated.Campaign {
	t.Helper()

	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	emailTemplate := suite.Client.DB.EmailTemplate.Create().
		SetName(name + " Template").
		SetKey(email.BrandedMessageOp.Name()).
		SetTemplateContext(enums.TemplateContextCampaignRecipient).
		SetDefaults(map[string]any{
			"subject": "Test",
			"title":   "Test",
			"intros":  []any{"Test"},
		}).
		SaveX(ctx)

	campaignObj := suite.Client.DB.Campaign.Create().
		SetName(name).
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetCampaignType(enums.CampaignTypeCustom).
		SetEmailTemplateID(emailTemplate.ID).
		SetRecurrenceFrequency(enums.FrequencyNone).
		SaveX(ctx)

	t.Cleanup(func() {
		(&th.Cleanup[*generated.CampaignDeleteOne]{Client: suite.Client.DB.Campaign, ID: campaignObj.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.EmailTemplateDeleteOne]{Client: suite.Client.DB.EmailTemplate, ID: emailTemplate.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	return campaignObj
}

// sendTestEmail invokes the sendCampaignTestEmail mutation as th.SharedTestUser1
func sendTestEmail(campaignID string, emails []string) (*testclient.SendCampaignTestEmail, error) {
	return suite.Client.API.SendCampaignTestEmail(th.SharedTestUser1.UserCtx, testclient.SendCampaignTestEmailInput{
		CampaignID: campaignID,
		Emails:     emails,
	})
}

// TestSendCampaignTestEmailAssessmentBackfill verifies a questionnaire campaign created with only a
// questionnaire template reference gets an assessment created from the template and linked on the
// first test email send, and that subsequent sends reuse the same assessment
func TestSendCampaignTestEmailAssessmentBackfill(t *testing.T) {
	template := (&th.TemplateBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	createResp, err := suite.Client.API.CreateCampaign(th.SharedTestUser1.UserCtx, testclient.CreateCampaignInput{
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
		allowCtx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

		responses := suite.Client.DB.AssessmentResponse.Query().
			Where(assessmentresponse.CampaignIDEQ(campaignID)).
			AllX(allowCtx)
		if len(responses) > 0 {
			(&th.Cleanup[*generated.AssessmentResponseDeleteOne]{
				Client: suite.Client.DB.AssessmentResponse,
				IDs:    lo.Map(responses, func(r *generated.AssessmentResponse, _ int) string { return r.ID }),
			}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*generated.CampaignDeleteOne]{Client: suite.Client.DB.Campaign, ID: campaignID}).MustDelete(th.SharedTestUser1.UserCtx, t)

		if assessmentID != "" {
			(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: assessmentID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, ID: template.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	sendResp, err := sendTestEmail(campaignID, []string{"backfill-recipient@test.example"})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(int64(1), sendResp.SendCampaignTestEmail.QueuedCount))
	assert.Check(t, is.Equal(int64(0), sendResp.SendCampaignTestEmail.SkippedCount))

	assessmentID = lo.FromPtr(sendResp.SendCampaignTestEmail.Campaign.AssessmentID)
	assert.Assert(t, assessmentID != "", "expected assessment to be backfilled from the questionnaire template")

	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)
	assessmentObj := suite.Client.DB.Assessment.GetX(ctx, assessmentID)
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
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	campaignObj := suite.Client.DB.Campaign.Create().
		SetName("Missing Email Template Test Campaign").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetCampaignType(enums.CampaignTypeCustom).
		SetRecurrenceFrequency(enums.FrequencyNone).
		SaveX(ctx)

	t.Cleanup(func() {
		(&th.Cleanup[*generated.CampaignDeleteOne]{Client: suite.Client.DB.Campaign, ID: campaignObj.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
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
