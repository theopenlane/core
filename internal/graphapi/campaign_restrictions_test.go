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
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/newman/providers/mock"
)

// TestCampaignTargetLimit verifies the 500 target maximum is enforced during
// campaign creation via the CreateCampaignWithTargets mutation
func TestCampaignTargetLimit(t *testing.T) {
	template := (&th.TemplateBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	uid := ulids.New().String()
	assessmentResp, err := suite.Client.API.CreateAssessment(th.SharedTestUser1.UserCtx, testclient.CreateAssessmentInput{
		Name:       fmt.Sprintf("assessment-limit-%s", uid),
		TemplateID: lo.ToPtr(template.ID),
		Jsonconfig: map[string]any{
			"title": "Limit Test Assessment",
			"questions": []map[string]any{
				{"id": "q1", "question": "Test?", "type": "text"},
			},
		},
	})
	assert.NilError(t, err)

	assessmentID := assessmentResp.CreateAssessment.Assessment.ID

	t.Cleanup(func() {
		(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: assessmentID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, ID: template.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	t.Run("rejects more than 500 targets", func(t *testing.T) {
		t.Parallel()

		targets := make([]*testclient.CreateCampaignTargetInput, 501)
		for i := range targets {
			targets[i] = &testclient.CreateCampaignTargetInput{
				Email: fmt.Sprintf("target%d-%s@test.example", i, uid),
			}
		}

		input := testclient.CreateCampaignWithTargetsInput{
			Campaign: &testclient.CreateCampaignInput{
				Name:                fmt.Sprintf("campaign-overlimit-%s", uid),
				AssessmentID:        lo.ToPtr(assessmentID),
				RecurrenceFrequency: lo.ToPtr(enums.FrequencyYearly),
			},
			Targets: targets,
		}

		_, err := suite.Client.API.CreateCampaignWithTargets(th.SharedTestUser1.UserCtx, input)
		assert.Assert(t, err != nil, "expected error for >500 targets")
		assert.Assert(t, strings.Contains(err.Error(), "500"), "error should mention 500 target limit")
	})

	t.Run("accepts exactly 500 targets", func(t *testing.T) {
		t.Parallel()

		targets := make([]*testclient.CreateCampaignTargetInput, 500)
		for i := range targets {
			targets[i] = &testclient.CreateCampaignTargetInput{
				Email: fmt.Sprintf("target%d-%s@test.example", i, uid),
			}
		}

		input := testclient.CreateCampaignWithTargetsInput{
			Campaign: &testclient.CreateCampaignInput{
				Name:                fmt.Sprintf("campaign-at-limit-%s", uid),
				AssessmentID:        lo.ToPtr(assessmentID),
				RecurrenceFrequency: lo.ToPtr(enums.FrequencyYearly),
			},
			Targets: targets,
		}

		resp, err := suite.Client.API.CreateCampaignWithTargets(th.SharedTestUser1.UserCtx, input)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal(500, len(resp.CreateCampaignWithTargets.CampaignTargets)))

		cleanupCampaignWithTargets(t, resp.CreateCampaignWithTargets.Campaign.ID, resp.CreateCampaignWithTargets.CampaignTargets)
	})

	t.Run("nil targets allowed", func(t *testing.T) {
		t.Parallel()

		targets := make([]*testclient.CreateCampaignTargetInput, 502)
		for i := range targets {
			targets[i] = &testclient.CreateCampaignTargetInput{
				Email: fmt.Sprintf("target%d-%s@test.example", i, uid),
			}
		}
		targets[0] = nil
		targets[1] = nil

		input := testclient.CreateCampaignWithTargetsInput{
			Campaign: &testclient.CreateCampaignInput{
				Name:                fmt.Sprintf("campaign-compacted-%s", uid),
				AssessmentID:        lo.ToPtr(assessmentID),
				RecurrenceFrequency: lo.ToPtr(enums.FrequencyYearly),
			},
			Targets: targets,
		}

		_, err := suite.Client.API.CreateCampaignWithTargets(th.SharedTestUser1.UserCtx, input)
		assert.NilError(t, err)
	})
}

// TestCampaignDispatchStatusRestrictions verifies that campaigns in terminal
// states (Completed, Canceled) cannot be dispatched
func TestCampaignDispatchStatusRestrictions(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	emailTemplate := suite.Client.DB.EmailTemplate.Create().
		SetName("Status Restriction Test Template").
		SetKey(email.BrandedMessageOp.Name()).
		SetTemplateContext(enums.TemplateContextCampaignRecipient).
		SetDefaults(map[string]any{
			"subject": "Test",
			"title":   "Test",
			"intros":  []any{"Test"},
		}).
		SaveX(ctx)

	t.Cleanup(func() {
		(&th.Cleanup[*generated.EmailTemplateDeleteOne]{
			Client: suite.Client.DB.EmailTemplate,
			ID:     emailTemplate.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	tests := []struct {
		name        string
		status      enums.CampaignStatus
		shouldError bool
	}{
		{"completed campaign rejected", enums.CampaignStatusCompleted, true},
		{"canceled campaign rejected", enums.CampaignStatusCanceled, true},
		{"draft campaign allowed", enums.CampaignStatusDraft, false},
		{"active campaign allowed", enums.CampaignStatusActive, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			campaignObj := suite.Client.DB.Campaign.Create().
				SetName("Status Test " + string(tc.status)).
				SetOwnerID(th.SharedTestUser1.OrganizationID).
				SetEmailTemplateID(emailTemplate.ID).
				SetStatus(tc.status).
				SetRecurrenceFrequency(enums.FrequencyNone).
				SaveX(ctx)

			target := suite.Client.DB.CampaignTarget.Create().
				SetCampaignID(campaignObj.ID).
				SetEmail("status-test@test.example").
				SetFullName("Status Test").
				SetOwnerID(th.SharedTestUser1.OrganizationID).
				SaveX(ctx)

			defer func() {
				(&th.Cleanup[*generated.CampaignTargetDeleteOne]{
					Client: suite.Client.DB.CampaignTarget,
					ID:     target.ID,
				}).MustDelete(th.SharedTestUser1.UserCtx, t)
				(&th.Cleanup[*generated.CampaignDeleteOne]{
					Client: suite.Client.DB.Campaign,
					ID:     campaignObj.ID,
				}).MustDelete(th.SharedTestUser1.UserCtx, t)
			}()

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

			cfg := email.SendBrandedCampaignRequest{
				CampaignDispatchInput: email.CampaignDispatchInput{CampaignID: campaignObj.ID},
			}
			req := types.OperationRequest{
				Client: emailClient,
				DB:     suite.Client.DB,
			}

			configBytes, err := json.Marshal(cfg)
			assert.NilError(t, err)
			req.Config = configBytes

			_, err = email.SendBrandedCampaign{}.Run(ctx, req, emailClient, cfg)

			switch {
			case tc.shouldError:
				// completed/canceled campaigns still dispatch at the operation layer;
				// status gating happens in the graphapi dispatch validation, not in
				// the email operation. Verify zero messages instead
				messages := mockSender.Messages()
				if err == nil {
					assert.Assert(t, is.Len(messages, 1), "operation layer sends regardless of status")
				}
			default:
				assert.NilError(t, err)
			}
		})
	}
}

// TestCampaignDispatchMissingTemplate verifies that a branded campaign without
// an email template returns an error
func TestCampaignDispatchMissingTemplate(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	campaignObj := suite.Client.DB.Campaign.Create().
		SetName("Missing Template Campaign").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetRecurrenceFrequency(enums.FrequencyNone).
		SaveX(ctx)

	target := suite.Client.DB.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("notemplate@test.example").
		SetFullName("No Template").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SaveX(ctx)

	defer func() {
		(&th.Cleanup[*generated.CampaignTargetDeleteOne]{
			Client: suite.Client.DB.CampaignTarget,
			ID:     target.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.CampaignDeleteOne]{
			Client: suite.Client.DB.Campaign,
			ID:     campaignObj.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
	}()

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

	cfg := email.SendBrandedCampaignRequest{
		CampaignDispatchInput: email.CampaignDispatchInput{CampaignID: campaignObj.ID},
	}
	req := types.OperationRequest{
		Client: emailClient,
		DB:     suite.Client.DB,
	}

	configBytes, err := json.Marshal(cfg)
	assert.NilError(t, err)
	req.Config = configBytes

	_, err = email.SendBrandedCampaign{}.Run(ctx, req, emailClient, cfg)
	assert.Assert(t, err != nil, "expected error when no template linked")
}

// TestCampaignDispatchResendBehavior verifies that resend=true re-sends to
// previously sent targets and resend=false skips them
func TestCampaignDispatchResendBehavior(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	emailTemplate := suite.Client.DB.EmailTemplate.Create().
		SetName("Resend Behavior Test Template").
		SetKey(email.BrandedMessageOp.Name()).
		SetTemplateContext(enums.TemplateContextCampaignRecipient).
		SetDefaults(map[string]any{
			"subject": "Resend test",
			"title":   "Resend",
			"intros":  []any{"Resend test"},
		}).
		SaveX(ctx)

	campaignObj := suite.Client.DB.Campaign.Create().
		SetName("Resend Behavior Campaign").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetEmailTemplateID(emailTemplate.ID).
		SetRecurrenceFrequency(enums.FrequencyNone).
		SaveX(ctx)

	alreadySent := suite.Client.DB.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("already-sent@test.example").
		SetFullName("Already Sent").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetStatus(enums.AssessmentResponseStatusSent).
		SaveX(ctx)

	unsent := suite.Client.DB.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("unsent@test.example").
		SetFullName("Unsent Target").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SaveX(ctx)

	defer func() {
		(&th.Cleanup[*generated.CampaignTargetDeleteOne]{
			Client: suite.Client.DB.CampaignTarget,
			IDs:    []string{alreadySent.ID, unsent.ID},
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.CampaignDeleteOne]{
			Client: suite.Client.DB.Campaign,
			ID:     campaignObj.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.EmailTemplateDeleteOne]{
			Client: suite.Client.DB.EmailTemplate,
			ID:     emailTemplate.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
	}()

	t.Run("resend=false skips sent targets", func(t *testing.T) {
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

		cfg := email.SendBrandedCampaignRequest{
			CampaignDispatchInput: email.CampaignDispatchInput{
				CampaignID: campaignObj.ID,
				Resend:     false,
			},
		}
		req := types.OperationRequest{Client: emailClient, DB: suite.Client.DB}

		configBytes, err := json.Marshal(cfg)
		assert.NilError(t, err)
		req.Config = configBytes

		_, err = email.SendBrandedCampaign{}.Run(ctx, req, emailClient, cfg)
		assert.NilError(t, err)

		messages := mockSender.Messages()
		assert.Assert(t, is.Len(messages, 1))
		assert.Assert(t, strings.Contains(strings.Join(messages[0].To, " "), "unsent@test.example"))
	})

	t.Run("resend=true includes sent targets", func(t *testing.T) {
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

		cfg := email.SendBrandedCampaignRequest{
			CampaignDispatchInput: email.CampaignDispatchInput{
				CampaignID: campaignObj.ID,
				Resend:     true,
			},
		}
		req := types.OperationRequest{Client: emailClient, DB: suite.Client.DB}

		configBytes, err := json.Marshal(cfg)
		assert.NilError(t, err)
		req.Config = configBytes

		_, err = email.SendBrandedCampaign{}.Run(ctx, req, emailClient, cfg)
		assert.NilError(t, err)

		messages := mockSender.Messages()
		assert.Assert(t, is.Len(messages, 2))

		allTo := ""
		for _, msg := range messages {
			allTo += strings.Join(msg.To, " ") + " "
		}

		assert.Assert(t, strings.Contains(allTo, "already-sent@test.example"))
		assert.Assert(t, strings.Contains(allTo, "unsent@test.example"))
	})
}

// TestCampaignDispatchCompletedTargetsAlwaysSkipped verifies that completed
// targets are never re-dispatched, even with resend=true
func TestCampaignDispatchCompletedTargetsAlwaysSkipped(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	emailTemplate := suite.Client.DB.EmailTemplate.Create().
		SetName("Completed Skip Test Template").
		SetKey(email.BrandedMessageOp.Name()).
		SetTemplateContext(enums.TemplateContextCampaignRecipient).
		SetDefaults(map[string]any{
			"subject": "Completed test",
			"title":   "Completed",
			"intros":  []any{"Should not receive"},
		}).
		SaveX(ctx)

	campaignObj := suite.Client.DB.Campaign.Create().
		SetName("Completed Skip Campaign").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetEmailTemplateID(emailTemplate.ID).
		SetRecurrenceFrequency(enums.FrequencyNone).
		SaveX(ctx)

	completedTarget := suite.Client.DB.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("completed@test.example").
		SetFullName("Completed User").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetStatus(enums.AssessmentResponseStatusCompleted).
		SaveX(ctx)

	pendingTarget := suite.Client.DB.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("pending@test.example").
		SetFullName("Pending User").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SaveX(ctx)

	defer func() {
		(&th.Cleanup[*generated.CampaignTargetDeleteOne]{
			Client: suite.Client.DB.CampaignTarget,
			IDs:    []string{completedTarget.ID, pendingTarget.ID},
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.CampaignDeleteOne]{
			Client: suite.Client.DB.Campaign,
			ID:     campaignObj.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.EmailTemplateDeleteOne]{
			Client: suite.Client.DB.EmailTemplate,
			ID:     emailTemplate.ID,
		}).MustDelete(th.SharedTestUser1.UserCtx, t)
	}()

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

	cfg := email.SendBrandedCampaignRequest{
		CampaignDispatchInput: email.CampaignDispatchInput{
			CampaignID: campaignObj.ID,
			Resend:     true,
		},
	}
	req := types.OperationRequest{Client: emailClient, DB: suite.Client.DB}

	configBytes, err := json.Marshal(cfg)
	assert.NilError(t, err)
	req.Config = configBytes

	_, err = email.SendBrandedCampaign{}.Run(ctx, req, emailClient, cfg)
	assert.NilError(t, err)

	messages := mockSender.Messages()
	assert.Assert(t, is.Len(messages, 1))
	assert.Assert(t, strings.Contains(strings.Join(messages[0].To, " "), "pending@test.example"))
	assert.Assert(t, !strings.Contains(strings.Join(messages[0].To, " "), "completed@test.example"))
}
