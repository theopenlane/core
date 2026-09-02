//go:build test

package graphapi_test

import (
	"fmt"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

// TestCreateCampaignWithTargets tests the createCampaignWithTargets mutation through the API.
func TestCreateCampaignWithTargets(t *testing.T) {
	// Create template via builder (no FGA edge checks on template from campaign)
	template := (&th.TemplateBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// Create assessment via API with same user so EdgeViewCheck passes
	uid := ulids.New().String()
	assessmentResp, err := suite.Client.API.CreateAssessment(th.SharedTestUser1.UserCtx, testclient.CreateAssessmentInput{
		Name:       fmt.Sprintf("assessment-%s", uid),
		TemplateID: lo.ToPtr(template.ID),
		Jsonconfig: map[string]any{
			"title":       "Campaign Test Assessment",
			"description": "Assessment for campaign testing",
			"questions": []map[string]any{
				{"id": "q1", "question": "Test question?", "type": "text"},
			},
		},
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

		resp, err := suite.Client.API.CreateCampaignWithTargets(th.SharedTestUser1.UserCtx, input)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Assert(t, resp.CreateCampaignWithTargets.Campaign.ID != "")
		assert.Check(t, is.Equal(len(input.Targets), len(resp.CreateCampaignWithTargets.CampaignTargets)))
		assert.Check(t, is.Equal(len(input.Targets), int(lo.FromPtr(resp.CreateCampaignWithTargets.Campaign.RecipientCount))))

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

		resp, err := suite.Client.API.CreateCampaignWithTargets(th.SharedTestUser1.UserCtx, input)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal(int(explicitCount), int(lo.FromPtr(resp.CreateCampaignWithTargets.Campaign.RecipientCount))))

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

		_, err := suite.Client.API.CreateCampaignWithTargets(th.SharedTestUser1.UserCtx, input)
		assert.Assert(t, err != nil)
	})

	// run as a graphql request to ensure it passes graphql validation
	t.Run("succeeds when targets omit campaignID", func(t *testing.T) {
		concreteClient, ok := suite.Client.API.TestGraphClient.(*testclient.Client)
		assert.Assert(t, ok)

		testUID := ulids.New().String()
		vars := map[string]any{
			"input": map[string]any{
				"campaign": map[string]any{
					"name":                fmt.Sprintf("campaign-%s", testUID),
					"assessmentID":        assessmentID,
					"recurrenceFrequency": "YEARLY",
				},
				"targets": []map[string]any{
					{"email": fmt.Sprintf("target-%s@test.example", testUID)},
				},
			},
		}

		var resp testclient.CreateCampaignWithTargets
		err := concreteClient.Client.Post(th.SharedTestUser1.UserCtx, "CreateCampaignWithTargets", testclient.CreateCampaignWithTargetsDocument, &resp, vars)
		assert.NilError(t, err)
		assert.Assert(t, resp.CreateCampaignWithTargets.Campaign.ID != "")
		assert.Check(t, is.Equal(1, len(resp.CreateCampaignWithTargets.CampaignTargets)))

		// cleanup
		cleanupCampaignWithTargets(t, resp.CreateCampaignWithTargets.Campaign.ID, resp.CreateCampaignWithTargets.CampaignTargets)
	})

	// cleanup assessment and template
	(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: assessmentID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, ID: template.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

// cleanupCampaignWithTargets deletes campaign targets and the campaign.
func cleanupCampaignWithTargets(t *testing.T, campaignID string, targets []*testclient.CreateCampaignWithTargets_CreateCampaignWithTargets_CampaignTargets) {
	t.Helper()

	targetIDs := make([]string, 0, len(targets))
	for _, target := range targets {
		targetIDs = append(targetIDs, target.ID)
	}

	(&th.Cleanup[*generated.CampaignTargetDeleteOne]{Client: suite.Client.DB.CampaignTarget, IDs: targetIDs}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.CampaignDeleteOne]{Client: suite.Client.DB.Campaign, ID: campaignID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}
