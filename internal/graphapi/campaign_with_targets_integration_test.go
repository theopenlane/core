//go:build test

package graphapi_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/samber/lo"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

// TestCreateCampaignWithTargets tests the createCampaignWithTargets mutation through the API.
func TestCreateCampaignWithTargets(t *testing.T) {
	// Create assessment via API to ensure proper FGA tuple setup
	template := (&TemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	jsonConfig := map[string]any{
		"title":       "Campaign Test Assessment",
		"description": "Assessment for campaign testing",
		"questions": []map[string]any{
			{"id": "q1", "question": "Test question?", "type": "text"},
		},
	}

	uid := ulids.New().String()
	assessmentResp, err := suite.client.apiWithPAT.CreateAssessment(context.Background(), testclient.CreateAssessmentInput{
		Name:       fmt.Sprintf("assessment-%s", uid),
		TemplateID: lo.ToPtr(template.ID),
		Jsonconfig: jsonConfig,
	})
	assert.NilError(t, err)
	assessmentID := assessmentResp.CreateAssessment.Assessment.ID

	t.Run("successful creation with targets", func(t *testing.T) {
		testUID := ulids.New().String()
		input := testclient.CreateCampaignWithTargetsInput{
			Campaign: &testclient.CreateCampaignInput{
				Name:                fmt.Sprintf("campaign-%s", testUID),
				AssessmentID:        lo.ToPtr(assessmentID),
				RecurrenceFrequency: lo.ToPtr(enums.FrequencyYearly),
			},
			Targets: []*testclient.CreateCampaignTargetInput{
				{Email: fmt.Sprintf("target1-%s@test.example", testUID)},
				{Email: fmt.Sprintf("target2-%s@test.example", testUID)},
				{Email: fmt.Sprintf("target3-%s@test.example", testUID)},
			},
		}

		resp, err := suite.client.apiWithPAT.CreateCampaignWithTargets(context.Background(), input)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Assert(t, resp.CreateCampaignWithTargets.Campaign.ID != "")
		assert.Check(t, is.Equal(len(input.Targets), len(resp.CreateCampaignWithTargets.CampaignTargets)))
		assert.Check(t, is.Equal(len(input.Targets), resp.CreateCampaignWithTargets.Campaign.RecipientCount))

		// cleanup
		cleanupCampaignWithTargets(t, resp.CreateCampaignWithTargets.Campaign.ID, resp.CreateCampaignWithTargets.CampaignTargets)
	})

	t.Run("preserves explicit recipient count", func(t *testing.T) {
		testUID := ulids.New().String()
		explicitCount := int64(100)
		input := testclient.CreateCampaignWithTargetsInput{
			Campaign: &testclient.CreateCampaignInput{
				Name:                fmt.Sprintf("campaign-%s", testUID),
				AssessmentID:        lo.ToPtr(assessmentID),
				RecurrenceFrequency: lo.ToPtr(enums.FrequencyYearly),
				RecipientCount:      &explicitCount,
			},
			Targets: []*testclient.CreateCampaignTargetInput{
				{Email: fmt.Sprintf("target-%s@test.example", testUID)},
			},
		}

		resp, err := suite.client.apiWithPAT.CreateCampaignWithTargets(context.Background(), input)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal(int(explicitCount), resp.CreateCampaignWithTargets.Campaign.RecipientCount))

		// cleanup
		cleanupCampaignWithTargets(t, resp.CreateCampaignWithTargets.Campaign.ID, resp.CreateCampaignWithTargets.CampaignTargets)
	})

	t.Run("fails with empty targets", func(t *testing.T) {
		testUID := ulids.New().String()
		input := testclient.CreateCampaignWithTargetsInput{
			Campaign: &testclient.CreateCampaignInput{
				Name:                fmt.Sprintf("campaign-%s", testUID),
				AssessmentID:        lo.ToPtr(assessmentID),
				RecurrenceFrequency: lo.ToPtr(enums.FrequencyYearly),
			},
			Targets: []*testclient.CreateCampaignTargetInput{},
		}

		_, err := suite.client.apiWithPAT.CreateCampaignWithTargets(context.Background(), input)
		assert.Assert(t, err != nil)
	})

	// cleanup assessment and template
	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, ID: assessmentID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: template.ID}).MustDelete(testUser1.UserCtx, t)
}

// cleanupCampaignWithTargets deletes campaign targets and the campaign.
func cleanupCampaignWithTargets(t *testing.T, campaignID string, targets []*testclient.CreateCampaignWithTargets_CreateCampaignWithTargets_CampaignTargets) {
	t.Helper()

	ctx := context.Background()

	for _, target := range targets {
		err := suite.client.db.CampaignTarget.DeleteOneID(target.ID).Exec(ctx)
		if err != nil {
			t.Logf("failed to delete campaign target %s: %v", target.ID, err)
		}
	}

	err := suite.client.db.Campaign.DeleteOneID(campaignID).Exec(ctx)
	if err != nil {
		t.Logf("failed to delete campaign %s: %v", campaignID, err)
	}
}
