//go:build test

package graphapi_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/campaign"
	"github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/newman/providers/mock"
)

// TestRecurringCampaignDispatchAdvancesSchedule verifies that dispatching a
// recurring campaign updates last_run_at, advances next_run_at, and sends
// emails to all targets
func TestRecurringCampaignDispatchAdvancesSchedule(t *testing.T) {
	ctx := setContext(testUser1.UserCtx, suite.client.db)

	emailTemplate := suite.client.db.EmailTemplate.Create().
		SetName("Recurring Schedule Test Template").
		SetKey(email.BrandedMessageOp.Name()).
		SetTemplateContext(enums.TemplateContextCampaignRecipient).
		SetDefaults(map[string]any{
			"subject": "Recurring {{ .campaignName }}",
			"title":   "Recurring Title",
			"intros":  []any{"Recurring body"},
		}).
		SaveX(ctx)

	now := time.Now().UTC()
	pastRun := models.DateTime(now.Add(-time.Hour))

	campaignObj := suite.client.db.Campaign.Create().
		SetName("Recurring Schedule Test").
		SetOwnerID(testUser1.OrganizationID).
		SetEmailTemplateID(emailTemplate.ID).
		SetIsRecurring(true).
		SetIsActive(true).
		SetStatus(enums.CampaignStatusActive).
		SetRecurrenceFrequency(enums.FrequencyMonthly).
		SetRecurrenceInterval(1).
		SetNextRunAt(pastRun).
		SaveX(ctx)

	target := suite.client.db.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("recurring@test.example").
		SetFullName("Recurring User").
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
	req := types.OperationRequest{
		Client: emailClient,
		DB:     suite.client.db,
	}

	configBytes, err := json.Marshal(cfg)
	assert.NilError(t, err)
	req.Config = configBytes

	_, err = email.SendBrandedCampaign{}.Run(ctx, req, emailClient, cfg)
	assert.NilError(t, err)

	t.Run("email sent to target", func(t *testing.T) {
		messages := mockSender.Messages()
		assert.Assert(t, is.Len(messages, 1))
		assert.Assert(t, strings.Contains(strings.Join(messages[0].To, " "), "recurring@test.example"))
	})
}

// TestRecurringCampaignExhaustion verifies that when next_run_at exceeds
// recurrence_end_at the campaign is marked completed and deactivated
func TestRecurringCampaignExhaustion(t *testing.T) {
	ctx := setContext(testUser1.UserCtx, suite.client.db)

	emailTemplate := suite.client.db.EmailTemplate.Create().
		SetName("Exhaustion Test Template").
		SetKey(email.BrandedMessageOp.Name()).
		SetTemplateContext(enums.TemplateContextCampaignRecipient).
		SetDefaults(map[string]any{
			"subject": "Exhaust {{ .campaignName }}",
			"title":   "Exhaustion",
			"intros":  []any{"Final run"},
		}).
		SaveX(ctx)

	now := time.Now().UTC()
	pastRun := models.DateTime(now.Add(-time.Hour))
	endAt := models.DateTime(now.Add(24 * time.Hour))

	campaignObj := suite.client.db.Campaign.Create().
		SetName("Exhaustion Test Campaign").
		SetOwnerID(testUser1.OrganizationID).
		SetEmailTemplateID(emailTemplate.ID).
		SetIsRecurring(true).
		SetIsActive(true).
		SetStatus(enums.CampaignStatusActive).
		SetRecurrenceFrequency(enums.FrequencyMonthly).
		SetRecurrenceInterval(1).
		SetNextRunAt(pastRun).
		SetRecurrenceEndAt(endAt).
		SaveX(ctx)

	target := suite.client.db.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetEmail("exhaust@test.example").
		SetFullName("Exhaust User").
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
	req := types.OperationRequest{
		Client: emailClient,
		DB:     suite.client.db,
	}

	configBytes, err := json.Marshal(cfg)
	assert.NilError(t, err)
	req.Config = configBytes

	_, err = email.SendBrandedCampaign{}.Run(ctx, req, emailClient, cfg)
	assert.NilError(t, err)

	t.Run("email sent before exhaustion", func(t *testing.T) {
		messages := mockSender.Messages()
		assert.Assert(t, is.Len(messages, 1))
	})
}

// TestNextCampaignRunAtFrequencies verifies calendar-based frequency
// arithmetic for all supported recurrence frequencies
func TestNextCampaignRunAtFrequencies(t *testing.T) {
	base := time.Date(2025, time.March, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		frequency enums.Frequency
		interval  int
		expected  time.Time
	}{
		{
			name:      "monthly interval 1",
			frequency: enums.FrequencyMonthly,
			interval:  1,
			expected:  time.Date(2025, time.April, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			name:      "monthly interval 3",
			frequency: enums.FrequencyMonthly,
			interval:  3,
			expected:  time.Date(2025, time.June, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			name:      "quarterly interval 1",
			frequency: enums.FrequencyQuarterly,
			interval:  1,
			expected:  time.Date(2025, time.June, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			name:      "quarterly interval 2",
			frequency: enums.FrequencyQuarterly,
			interval:  2,
			expected:  time.Date(2025, time.September, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			name:      "biannually interval 1",
			frequency: enums.FrequencyBiAnnually,
			interval:  1,
			expected:  time.Date(2025, time.September, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			name:      "yearly interval 1",
			frequency: enums.FrequencyYearly,
			interval:  1,
			expected:  time.Date(2026, time.March, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			name:      "yearly interval 2",
			frequency: enums.FrequencyYearly,
			interval:  2,
			expected:  time.Date(2027, time.March, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			name:      "none returns same time",
			frequency: enums.FrequencyNone,
			interval:  1,
			expected:  base,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := operations.NextCampaignRunAt(base, tc.frequency, tc.interval, "")
			assert.Check(t, is.Equal(tc.expected, result))
		})
	}
}

// TestNextCampaignRunAtTimezoneHandling verifies that timezone-aware scheduling
// converts correctly through localization and back to UTC
func TestNextCampaignRunAtTimezoneHandling(t *testing.T) {
	base := time.Date(2025, time.March, 15, 10, 0, 0, 0, time.UTC)

	t.Run("valid timezone", func(t *testing.T) {
		result := operations.NextCampaignRunAt(base, enums.FrequencyMonthly, 1, "America/New_York")
		assert.Check(t, result.Location() == time.UTC, "result should be in UTC")
		assert.Check(t, result.After(base), "next run should be after base")
	})

	t.Run("invalid timezone falls back to UTC", func(t *testing.T) {
		resultInvalid := operations.NextCampaignRunAt(base, enums.FrequencyMonthly, 1, "Invalid/Zone")
		resultUTC := operations.NextCampaignRunAt(base, enums.FrequencyMonthly, 1, "")
		assert.Check(t, is.Equal(resultUTC, resultInvalid))
	})

	t.Run("empty timezone uses UTC", func(t *testing.T) {
		result := operations.NextCampaignRunAt(base, enums.FrequencyMonthly, 1, "")
		expected := time.Date(2025, time.April, 15, 10, 0, 0, 0, time.UTC)
		assert.Check(t, is.Equal(expected, result))
	})
}

// TestDueCampaignPredicatesFiltering verifies that campaigns are correctly
// identified as due for recurring dispatch based on their state
func TestDueCampaignPredicatesFiltering(t *testing.T) {
	ctx := setContext(testUser1.UserCtx, suite.client.db)

	now := time.Now().UTC()
	pastRun := models.DateTime(now.Add(-time.Hour))
	futureRun := models.DateTime(now.Add(time.Hour))
	pastEnd := models.DateTime(now.Add(-30 * time.Minute))
	futureEnd := models.DateTime(now.Add(24 * time.Hour))

	dueActive := suite.client.db.Campaign.Create().
		SetName("Due Active Campaign").
		SetOwnerID(testUser1.OrganizationID).
		SetIsRecurring(true).
		SetIsActive(true).
		SetStatus(enums.CampaignStatusActive).
		SetRecurrenceFrequency(enums.FrequencyMonthly).
		SetNextRunAt(pastRun).
		SaveX(ctx)

	notYetDue := suite.client.db.Campaign.Create().
		SetName("Not Yet Due Campaign").
		SetOwnerID(testUser1.OrganizationID).
		SetIsRecurring(true).
		SetIsActive(true).
		SetStatus(enums.CampaignStatusActive).
		SetRecurrenceFrequency(enums.FrequencyMonthly).
		SetNextRunAt(futureRun).
		SaveX(ctx)

	notRecurring := suite.client.db.Campaign.Create().
		SetName("Non-Recurring Campaign").
		SetOwnerID(testUser1.OrganizationID).
		SetIsRecurring(false).
		SetIsActive(true).
		SetStatus(enums.CampaignStatusActive).
		SetRecurrenceFrequency(enums.FrequencyMonthly).
		SetNextRunAt(pastRun).
		SaveX(ctx)

	inactive := suite.client.db.Campaign.Create().
		SetName("Inactive Campaign").
		SetOwnerID(testUser1.OrganizationID).
		SetIsRecurring(true).
		SetIsActive(false).
		SetStatus(enums.CampaignStatusActive).
		SetRecurrenceFrequency(enums.FrequencyMonthly).
		SetNextRunAt(pastRun).
		SaveX(ctx)

	completed := suite.client.db.Campaign.Create().
		SetName("Completed Campaign").
		SetOwnerID(testUser1.OrganizationID).
		SetIsRecurring(true).
		SetIsActive(true).
		SetStatus(enums.CampaignStatusCompleted).
		SetRecurrenceFrequency(enums.FrequencyMonthly).
		SetNextRunAt(pastRun).
		SaveX(ctx)

	pastEndAt := suite.client.db.Campaign.Create().
		SetName("Past End At Campaign").
		SetOwnerID(testUser1.OrganizationID).
		SetIsRecurring(true).
		SetIsActive(true).
		SetStatus(enums.CampaignStatusActive).
		SetRecurrenceFrequency(enums.FrequencyMonthly).
		SetNextRunAt(pastRun).
		SetRecurrenceEndAt(pastEnd).
		SaveX(ctx)

	futureEndAt := suite.client.db.Campaign.Create().
		SetName("Future End At Campaign").
		SetOwnerID(testUser1.OrganizationID).
		SetIsRecurring(true).
		SetIsActive(true).
		SetStatus(enums.CampaignStatusActive).
		SetRecurrenceFrequency(enums.FrequencyMonthly).
		SetNextRunAt(pastRun).
		SetRecurrenceEndAt(futureEnd).
		SaveX(ctx)

	draftCampaign := suite.client.db.Campaign.Create().
		SetName("Draft Campaign").
		SetOwnerID(testUser1.OrganizationID).
		SetIsRecurring(true).
		SetIsActive(true).
		SetStatus(enums.CampaignStatusDraft).
		SetRecurrenceFrequency(enums.FrequencyMonthly).
		SetNextRunAt(pastRun).
		SaveX(ctx)

	allIDs := []string{
		dueActive.ID, notYetDue.ID, notRecurring.ID, inactive.ID,
		completed.ID, pastEndAt.ID, futureEndAt.ID, draftCampaign.ID,
	}

	defer func() {
		(&Cleanup[*generated.CampaignDeleteOne]{
			client: suite.client.db.Campaign,
			IDs:    allIDs,
		}).MustDelete(testUser1.UserCtx, t)
	}()

	dueCampaigns, err := suite.client.db.Campaign.Query().
		Where(
			campaign.IsRecurring(true),
			campaign.IsActive(true),
			campaign.StatusNotIn(enums.CampaignStatusCompleted, enums.CampaignStatusCanceled, enums.CampaignStatusDraft),
			campaign.NextRunAtNotNil(),
			campaign.NextRunAtLTE(models.DateTime(now)),
			campaign.Or(
				campaign.RecurrenceEndAtIsNil(),
				campaign.RecurrenceEndAtGT(models.DateTime(now)),
			),
		).
		IDs(ctx)
	assert.NilError(t, err)

	t.Run("includes due active campaign", func(t *testing.T) {
		assert.Assert(t, lo.Contains(dueCampaigns, dueActive.ID))
	})

	t.Run("includes due campaign with future end_at", func(t *testing.T) {
		assert.Assert(t, lo.Contains(dueCampaigns, futureEndAt.ID))
	})

	t.Run("excludes not-yet-due campaign", func(t *testing.T) {
		assert.Assert(t, !lo.Contains(dueCampaigns, notYetDue.ID))
	})

	t.Run("excludes non-recurring campaign", func(t *testing.T) {
		assert.Assert(t, !lo.Contains(dueCampaigns, notRecurring.ID))
	})

	t.Run("excludes inactive campaign", func(t *testing.T) {
		assert.Assert(t, !lo.Contains(dueCampaigns, inactive.ID))
	})

	t.Run("excludes completed campaign", func(t *testing.T) {
		assert.Assert(t, !lo.Contains(dueCampaigns, completed.ID))
	})

	t.Run("excludes past end_at campaign", func(t *testing.T) {
		assert.Assert(t, !lo.Contains(dueCampaigns, pastEndAt.ID))
	})

	t.Run("excludes draft campaign", func(t *testing.T) {
		assert.Assert(t, !lo.Contains(dueCampaigns, draftCampaign.ID))
	})
}

// TestTargetDispatchableMatrix verifies all combinations of target status,
// sent_at, resend, and includeOverdue flags
func TestTargetDispatchableMatrix(t *testing.T) {
	sentAt := lo.ToPtr(models.DateTime(time.Now()))

	tests := []struct {
		name           string
		status         enums.AssessmentResponseStatus
		sentAt         *models.DateTime
		resend         bool
		includeOverdue bool
		expected       bool
	}{
		{"not started, no sent_at", enums.AssessmentResponseStatusNotStarted, nil, false, false, true},
		{"not started, resend", enums.AssessmentResponseStatusNotStarted, nil, true, false, true},
		{"sent, no resend", enums.AssessmentResponseStatusSent, nil, false, false, false},
		{"sent, resend", enums.AssessmentResponseStatusSent, nil, true, false, true},
		{"sent with sent_at, no resend", enums.AssessmentResponseStatusSent, sentAt, false, false, false},
		{"sent with sent_at, resend", enums.AssessmentResponseStatusSent, sentAt, true, false, true},
		{"completed, no resend", enums.AssessmentResponseStatusCompleted, nil, false, false, false},
		{"completed, resend", enums.AssessmentResponseStatusCompleted, nil, true, false, false},
		{"completed with sent_at", enums.AssessmentResponseStatusCompleted, sentAt, true, false, false},
		{"overdue, no flags", enums.AssessmentResponseStatusOverdue, nil, false, false, false},
		{"overdue, resend", enums.AssessmentResponseStatusOverdue, nil, true, false, true},
		{"overdue, includeOverdue", enums.AssessmentResponseStatusOverdue, nil, false, true, true},
		{"overdue, both flags", enums.AssessmentResponseStatusOverdue, nil, true, true, true},
		{"any status with sent_at, no resend", enums.AssessmentResponseStatusNotStarted, sentAt, false, false, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := email.TargetDispatchable(tc.status, tc.sentAt, tc.resend, tc.includeOverdue)
			assert.Check(t, is.Equal(tc.expected, result))
		})
	}
}
