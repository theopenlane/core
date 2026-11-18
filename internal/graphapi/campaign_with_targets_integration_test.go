//go:build test

package graphapi_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
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
		TemplateID: new(template.ID),
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
				AssessmentID:        new(assessmentID),
				RecurrenceFrequency: new(enums.FrequencyYearly),
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
				AssessmentID:        new(assessmentID),
				RecurrenceFrequency: new(enums.FrequencyYearly),
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
				AssessmentID:        new(assessmentID),
				RecurrenceFrequency: new(enums.FrequencyYearly),
			},
			Targets: []*testclient.CreateCampaignTargetInput{},
		}

		_, err := suite.Client.API.CreateCampaignWithTargets(th.SharedTestUser1.UserCtx, input)
		assert.Assert(t, err != nil)
	})

	t.Run("fails with an assessment from another organization", func(t *testing.T) {
		otherOrgAssessment := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser2.UserCtx, t)

		testUID := ulids.New().String()
		input := testclient.CreateCampaignWithTargetsInput{
			Campaign: &testclient.CreateCampaignInput{
				Name:                fmt.Sprintf("campaign-%s", testUID),
				AssessmentID:        new(otherOrgAssessment.ID),
				RecurrenceFrequency: new(enums.FrequencyYearly),
			},
			Targets: []*testclient.CreateCampaignTargetInput{
				{Email: fmt.Sprintf("target-%s@test.example", testUID)},
			},
		}

		_, err := suite.Client.API.CreateCampaignWithTargets(th.SharedTestUser1.UserCtx, input)
		assert.Assert(t, err != nil)

		(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: otherOrgAssessment.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
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

func TestMutationUpdateCampaign(t *testing.T) {
	assessment := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	otherOrgAssessment := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser2.UserCtx, t)

	createResp, err := suite.Client.API.CreateCampaign(th.SharedTestUser1.UserCtx, testclient.CreateCampaignInput{
		Name: fmt.Sprintf("campaign-%s", ulids.New().String()),
	})
	assert.NilError(t, err)

	campaignID := createResp.CreateCampaign.Campaign.ID
	dueDate := models.DateTime(time.Now().Add(7 * 24 * time.Hour))

	testCases := []struct {
		name     string
		id       string
		request  testclient.UpdateCampaignInput
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name: "happy path, update name",
			id:   campaignID,
			request: testclient.UpdateCampaignInput{
				Name: new(fmt.Sprintf("updated-campaign-%s", ulids.New().String())),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, update description",
			id:   campaignID,
			request: testclient.UpdateCampaignInput{
				Description: new("updated description"),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, clear description",
			id:   campaignID,
			request: testclient.UpdateCampaignInput{
				ClearDescription: new(true),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, update tags",
			id:   campaignID,
			request: testclient.UpdateCampaignInput{
				Tags: []string{"updated-tag1", "updated-tag2"},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, append tags",
			id:   campaignID,
			request: testclient.UpdateCampaignInput{
				AppendTags: []string{"appended-tag"},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, update due date",
			id:   campaignID,
			request: testclient.UpdateCampaignInput{
				DueDate: &dueDate,
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name: "happy path, update recurrence",
			id:   campaignID,
			request: testclient.UpdateCampaignInput{
				RecurrenceFrequency: new(enums.FrequencyMonthly),
				RecurrenceInterval:  new(int64(2)),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, link assessment in the same organization",
			id:   campaignID,
			request: testclient.UpdateCampaignInput{
				AssessmentID: new(assessment.ID),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, clear assessment",
			id:   campaignID,
			request: testclient.UpdateCampaignInput{
				ClearAssessment: new(true),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path using personal access token",
			id:   campaignID,
			request: testclient.UpdateCampaignInput{
				Name: new(fmt.Sprintf("pat-campaign-%s", ulids.New().String())),
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "not authorized, assessment from another organization",
			id:   campaignID,
			request: testclient.UpdateCampaignInput{
				AssessmentID: new(otherOrgAssessment.ID),
			},
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotAuthorizedErrorMsg,
		},
		{
			name: "not found, view only user",
			id:   campaignID,
			request: testclient.UpdateCampaignInput{
				Name: new("view-only-update"),
			},
			client:   suite.Client.API,
			ctx:      th.SharedViewOnlyUser.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name: "not found, invalid ID",
			id:   ulids.New().String(),
			request: testclient.UpdateCampaignInput{
				Name: new("invalid-id-update"),
			},
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name: "no access, user of different org",
			id:   campaignID,
			request: testclient.UpdateCampaignInput{
				Name: new("other-org-update"),
			},
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateCampaign(tc.ctx, tc.id, tc.request)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			campaign := resp.UpdateCampaign.Campaign
			assert.Check(t, is.Equal(tc.id, campaign.ID))

			if tc.request.Name != nil {
				assert.Check(t, is.Equal(*tc.request.Name, campaign.Name))
			}

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, lo.FromPtr(campaign.Description)))
			}

			if tc.request.ClearDescription != nil {
				assert.Check(t, is.Equal("", lo.FromPtr(campaign.Description)))
			}

			if len(tc.request.Tags) > 0 {
				assert.Check(t, is.DeepEqual(tc.request.Tags, campaign.Tags))
			}

			if len(tc.request.AppendTags) > 0 {
				assert.Check(t, is.Contains(campaign.Tags, tc.request.AppendTags[0]))
			}

			if tc.request.DueDate != nil {
				assert.Check(t, campaign.DueDate != nil)
			}

			if tc.request.RecurrenceFrequency != nil {
				assert.Check(t, is.Equal(*tc.request.RecurrenceFrequency, lo.FromPtr(campaign.RecurrenceFrequency)))
				assert.Check(t, is.Equal(*tc.request.RecurrenceInterval, lo.FromPtr(campaign.RecurrenceInterval)))
			}

			if tc.request.AssessmentID != nil {
				assert.Check(t, is.Equal(*tc.request.AssessmentID, lo.FromPtr(campaign.AssessmentID)))
			}

			if tc.request.ClearAssessment != nil {
				assert.Check(t, is.Equal("", lo.FromPtr(campaign.AssessmentID)))
			}
		})
	}

	(&th.Cleanup[*generated.CampaignDeleteOne]{Client: suite.Client.DB.Campaign, ID: campaignID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: assessment.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, ID: assessment.TemplateID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: otherOrgAssessment.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, ID: otherOrgAssessment.TemplateID}).MustDelete(th.SharedTestUser2.UserCtx, t)
}
