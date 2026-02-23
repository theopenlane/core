package graphapi

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
)

// TestCampaignTargetDispatchableOverdueResend ensures overdue targets can be resent.
func TestCampaignTargetDispatchableOverdueResend(t *testing.T) {
	t.Parallel()

	dispatchable := campaignTargetDispatchable(enums.AssessmentResponseStatusOverdue, true, false)
	assert.Check(t, dispatchable)
}

// TestResolveCampaignScheduleAt validates schedule resolution rules.
func TestResolveCampaignScheduleAt(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, time.January, 2, 15, 4, 5, 0, time.UTC)
	past := models.DateTime(now.Add(-time.Hour))
	future := models.DateTime(now.Add(time.Hour))

	_, err := resolveCampaignScheduleAt(now, nil, campaignDispatchActionLaunch, &past)
	assert.Assert(t, err != nil)

	scheduled, err := resolveCampaignScheduleAt(now, nil, campaignDispatchActionLaunch, &future)
	assert.NilError(t, err)
	assert.Assert(t, scheduled != nil)
	assert.Check(t, cmp.Equal(*scheduled, time.Time(future)))

	campaign := &generated.Campaign{ScheduledAt: &future}
	scheduled, err = resolveCampaignScheduleAt(now, campaign, campaignDispatchActionLaunch, nil)
	assert.NilError(t, err)
	assert.Assert(t, scheduled != nil)
	assert.Check(t, cmp.Equal(*scheduled, time.Time(future)))

	scheduled, err = resolveCampaignScheduleAt(now, campaign, campaignDispatchActionResend, nil)
	assert.NilError(t, err)
	assert.Check(t, scheduled == nil)
}

// TestShouldSetCampaignDueDate ensures due-date behavior on resend.
func TestShouldSetCampaignDueDate(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, time.January, 2, 15, 4, 5, 0, time.UTC)
	past := models.DateTime(now.Add(-time.Hour))
	future := models.DateTime(now.Add(time.Hour))

	campaignPast := &generated.Campaign{DueDate: &past}
	campaignFuture := &generated.Campaign{DueDate: &future}

	assert.Check(t, !shouldSetCampaignDueDate(campaignPast, true, now))
	assert.Check(t, shouldSetCampaignDueDate(campaignFuture, true, now))
	assert.Check(t, shouldSetCampaignDueDate(campaignFuture, false, now))
}
