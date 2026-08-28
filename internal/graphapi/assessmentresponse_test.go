package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/v2/internal/ent/generated/campaign"
	"github.com/theopenlane/core/v2/internal/ent/generated/campaigntarget"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// TestQueryAssessmentResponse verifies fetching a single assessment response.
func TestQueryAssessmentResponse(t *testing.T) {
	assessment1 := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	response1 := (&th.AssessmentResponseBuilder{
		Client:       suite.Client,
		AssessmentID: assessment1.ID,
		OwnerID:      assessment1.OwnerID,
	}).MustNew(th.SharedTestUser1.UserCtx, t)

	assessment2 := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	response2 := (&th.AssessmentResponseBuilder{
		Client:       suite.Client,
		AssessmentID: assessment2.ID,
		OwnerID:      assessment2.OwnerID,
	}).MustNew(th.SharedAdminUser.UserCtx, t)

	testCases := []struct {
		name           string
		queryID        string
		client         *testclient.TestClient
		ctx            context.Context
		expectedResult *generated.AssessmentResponse
		errorMsg       string
	}{
		{
			name:           "happy path",
			queryID:        response1.ID,
			client:         suite.Client.API,
			ctx:            th.SharedTestUser1.UserCtx,
			expectedResult: response1,
		},
		{
			name:           "happy path, response created by admin user",
			queryID:        response2.ID,
			client:         suite.Client.API,
			ctx:            th.SharedTestUser1.UserCtx,
			expectedResult: response2,
		},
		{
			name:           "happy path using personal access token",
			queryID:        response1.ID,
			client:         suite.Client.APIWithPAT,
			ctx:            context.Background(),
			expectedResult: response1,
		},
		{
			name:     "no access, user of different org",
			queryID:  response1.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "not found, invalid ID",
			queryID:  ulids.New().String(),
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAssessmentResponseByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.expectedResult.ID, resp.AssessmentResponse.ID))
			assert.Assert(t, resp.AssessmentResponse.Email != nil)
			assert.Check(t, is.Equal(tc.expectedResult.Email, *resp.AssessmentResponse.Email))
			assert.Check(t, is.Equal(tc.expectedResult.AssessmentID, resp.AssessmentResponse.AssessmentID))
		})
	}

	(&th.Cleanup[*generated.AssessmentResponseDeleteOne]{Client: suite.Client.DB.AssessmentResponse, IDs: []string{response1.ID, response2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, IDs: []string{assessment1.ID, assessment2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, IDs: []string{assessment1.TemplateID, assessment2.TemplateID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

// TestQueryAssessmentResponses verifies listing assessment responses.
func TestQueryAssessmentResponses(t *testing.T) {
	assessment1 := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	response1 := (&th.AssessmentResponseBuilder{
		Client:       suite.Client,
		AssessmentID: assessment1.ID,
		OwnerID:      assessment1.OwnerID,
	}).MustNew(th.SharedTestUser1.UserCtx, t)

	assessment2 := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	response2 := (&th.AssessmentResponseBuilder{
		Client:       suite.Client,
		AssessmentID: assessment2.ID,
		OwnerID:      assessment2.OwnerID,
	}).MustNew(th.SharedTestUser1.UserCtx, t)

	anotherUser := suite.UserBuilder(context.Background(), t)
	assessment3 := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(anotherUser.UserCtx, t)
	(&th.AssessmentResponseBuilder{
		Client:       suite.Client,
		AssessmentID: assessment3.ID,
		OwnerID:      assessment3.OwnerID,
	}).MustNew(anotherUser.UserCtx, t)

	t.Run("Get all assessment responses", func(t *testing.T) {
		resp, err := suite.Client.API.GetAllAssessmentResponses(th.SharedTestUser1.UserCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, resp.AssessmentResponses.TotalCount >= 2)
	})

	t.Run("Get assessment responses with where filter", func(t *testing.T) {
		email := response1.Email
		whereInput := &testclient.AssessmentResponseWhereInput{
			Email: &email,
		}

		resp, err := suite.Client.API.GetAssessmentResponses(th.SharedTestUser1.UserCtx, nil, nil, whereInput)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Len(resp.AssessmentResponses.Edges, 1))
		assert.Check(t, is.Equal(response1.ID, resp.AssessmentResponses.Edges[0].Node.ID))
	})

	t.Run("Get assessment responses using personal access token", func(t *testing.T) {
		resp, err := suite.Client.APIWithPAT.GetAllAssessmentResponses(context.Background())

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, resp.AssessmentResponses.TotalCount >= 2)
	})

	(&th.Cleanup[*generated.AssessmentResponseDeleteOne]{Client: suite.Client.DB.AssessmentResponse, IDs: []string{response1.ID, response2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, IDs: []string{assessment1.ID, assessment2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, IDs: []string{assessment1.TemplateID, assessment2.TemplateID}}).MustDelete(th.SharedTestUser1.UserCtx, t)

	th.CleanupOrganizationDataWithContext(anotherUser.UserCtx, t)
}

// TestAssessmentResponseCampaignIsolation ensures responses are isolated per campaign.
func TestAssessmentResponseCampaignIsolation(t *testing.T) {
	assessment := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	campaignA, err := suite.Client.DB.Campaign.Create().
		SetName("Campaign A").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetRecurrenceFrequency(enums.FrequencyYearly).
		Save(ctx)
	assert.NilError(t, err)

	campaignB, err := suite.Client.DB.Campaign.Create().
		SetName("Campaign B").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetRecurrenceFrequency(enums.FrequencyYearly).
		Save(ctx)
	assert.NilError(t, err)

	email := gofakeit.Email()
	responseA, err := suite.Client.DB.AssessmentResponse.Create().
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetCampaignID(campaignA.ID).
		SetEmail(email).
		Save(ctx)
	assert.NilError(t, err)

	responseB, err := suite.Client.DB.AssessmentResponse.Create().
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetCampaignID(campaignB.ID).
		SetEmail(email).
		Save(ctx)
	assert.NilError(t, err)

	countA, err := suite.Client.DB.AssessmentResponse.Query().
		Where(
			assessmentresponse.CampaignIDEQ(campaignA.ID),
			assessmentresponse.EmailEqualFold(email),
		).
		Count(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(1, countA))

	countB, err := suite.Client.DB.AssessmentResponse.Query().
		Where(
			assessmentresponse.CampaignIDEQ(campaignB.ID),
			assessmentresponse.EmailEqualFold(email),
		).
		Count(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(1, countB))

	(&th.Cleanup[*generated.AssessmentResponseDeleteOne]{Client: suite.Client.DB.AssessmentResponse, IDs: []string{responseA.ID, responseB.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.CampaignDeleteOne]{Client: suite.Client.DB.Campaign, IDs: []string{campaignA.ID, campaignB.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: assessment.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, ID: assessment.TemplateID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

// TestAssessmentResponseUpdatesCampaignTargetsAndCompletion verifies campaign rollups on completion.
func TestAssessmentResponseUpdatesCampaignTargetsAndCompletion(t *testing.T) {
	assessment := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	campaignObj, err := suite.Client.DB.Campaign.Create().
		SetName("Campaign Target Sync").
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetRecurrenceFrequency(enums.FrequencyYearly).
		Save(ctx)
	assert.NilError(t, err)

	emails := []string{"alpha@example.com", "beta@example.com"}
	targetIDs := make([]string, 0, len(emails))
	for _, email := range emails {
		target, err := suite.Client.DB.CampaignTarget.Create().
			SetOwnerID(th.SharedTestUser1.OrganizationID).
			SetCampaignID(campaignObj.ID).
			SetEmail(email).
			Save(ctx)
		assert.NilError(t, err)
		targetIDs = append(targetIDs, target.ID)
	}

	responseA, err := suite.Client.DB.AssessmentResponse.Create().
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetCampaignID(campaignObj.ID).
		SetEmail(emails[0]).
		Save(ctx)
	assert.NilError(t, err)

	responseB, err := suite.Client.DB.AssessmentResponse.Create().
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetCampaignID(campaignObj.ID).
		SetEmail(emails[1]).
		Save(ctx)
	assert.NilError(t, err)

	_, err = suite.Client.DB.AssessmentResponse.UpdateOneID(responseA.ID).
		SetStatus(enums.AssessmentResponseStatusCompleted).
		Save(ctx)
	assert.NilError(t, err)

	targetA, err := suite.Client.DB.CampaignTarget.Query().
		Where(
			campaigntarget.CampaignIDEQ(campaignObj.ID),
			campaigntarget.EmailEqualFold(emails[0]),
		).
		Only(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(enums.AssessmentResponseStatusCompleted, targetA.Status))

	targetB, err := suite.Client.DB.CampaignTarget.Query().
		Where(
			campaigntarget.CampaignIDEQ(campaignObj.ID),
			campaigntarget.EmailEqualFold(emails[1]),
		).
		Only(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(enums.AssessmentResponseStatusNotStarted, targetB.Status))

	campaignAfterFirst, err := suite.Client.DB.Campaign.Query().
		Where(campaign.IDEQ(campaignObj.ID)).
		Only(ctx)
	assert.NilError(t, err)
	assert.Check(t, campaignAfterFirst.Status != enums.CampaignStatusCompleted)

	_, err = suite.Client.DB.AssessmentResponse.UpdateOneID(responseB.ID).
		SetStatus(enums.AssessmentResponseStatusCompleted).
		Save(ctx)
	assert.NilError(t, err)

	campaignAfterSecond, err := suite.Client.DB.Campaign.Query().
		Where(campaign.IDEQ(campaignObj.ID)).
		Only(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(enums.CampaignStatusCompleted, campaignAfterSecond.Status))
	assert.Check(t, !campaignAfterSecond.IsActive)

	(&th.Cleanup[*generated.AssessmentResponseDeleteOne]{Client: suite.Client.DB.AssessmentResponse, IDs: []string{responseA.ID, responseB.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.CampaignTargetDeleteOne]{Client: suite.Client.DB.CampaignTarget, IDs: targetIDs}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.CampaignDeleteOne]{Client: suite.Client.DB.Campaign, ID: campaignObj.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: assessment.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, ID: assessment.TemplateID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

// TestMutationCreateAssessmentResponse validates create mutation behavior.
func TestMutationCreateAssessmentResponse(t *testing.T) {
	assessment := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	assessment2 := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser2.UserCtx, t)

	testCases := []struct {
		name    string
		request testclient.CreateAssessmentResponseInput
		client  *testclient.TestClient
		ctx     context.Context
	}{
		{
			name: "success - can create via GraphQL",
			request: testclient.CreateAssessmentResponseInput{
				Email:        lo.ToPtr(gofakeit.Email()),
				AssessmentID: assessment.ID,
				OwnerID:      &th.SharedTestUser1.OrganizationID,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "success - can create via PAT",
			request: testclient.CreateAssessmentResponseInput{
				Email:        lo.ToPtr(gofakeit.Email()),
				AssessmentID: assessment.ID,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "success - different org user can create",
			request: testclient.CreateAssessmentResponseInput{
				Email:        lo.ToPtr(gofakeit.Email()),
				AssessmentID: assessment2.ID,
				OwnerID:      &th.SharedTestUser2.OrganizationID,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser2.UserCtx,
		},
	}

	var responseIDsOrg1 []string
	var responseIDsOrg2 []string
	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateAssessmentResponse(tc.ctx, tc.request)

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Assert(t, resp.CreateAssessmentResponse.AssessmentResponse.ID != "")

			if tc.ctx == th.SharedTestUser2.UserCtx {
				responseIDsOrg2 = append(responseIDsOrg2, resp.CreateAssessmentResponse.AssessmentResponse.ID)
			} else {
				responseIDsOrg1 = append(responseIDsOrg1, resp.CreateAssessmentResponse.AssessmentResponse.ID)
			}
		})
	}

	t.Run("send attempts should increment on duplicate response", func(t *testing.T) {
		req := testclient.CreateAssessmentResponseInput{
			Email:        lo.ToPtr(gofakeit.Email()),
			AssessmentID: assessment.ID,
			OwnerID:      &th.SharedTestUser1.OrganizationID,
		}

		resp, err := suite.Client.API.CreateAssessmentResponse(th.SharedTestUser1.UserCtx, req)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		firstResponse := resp.CreateAssessmentResponse.AssessmentResponse
		assert.Check(t, is.Equal(int64(1), firstResponse.SendAttempts))
		responseIDsOrg1 = append(responseIDsOrg1, firstResponse.ID)

		secondResp, err := suite.Client.API.CreateAssessmentResponse(th.SharedTestUser1.UserCtx, req)
		assert.NilError(t, err)
		assert.Assert(t, secondResp != nil)

		updatedResponse := secondResp.CreateAssessmentResponse.AssessmentResponse
		assert.Check(t, is.Equal(firstResponse.ID, updatedResponse.ID))
		assert.Check(t, is.Equal(firstResponse.SendAttempts+1, updatedResponse.SendAttempts))
	})

	t.Run("completed response should not be updated", func(t *testing.T) {
		req := testclient.CreateAssessmentResponseInput{
			Email:        lo.ToPtr(gofakeit.Email()),
			AssessmentID: assessment.ID,
			OwnerID:      &th.SharedTestUser1.OrganizationID,
		}

		resp, err := suite.Client.API.CreateAssessmentResponse(th.SharedTestUser1.UserCtx, req)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		response := resp.CreateAssessmentResponse.AssessmentResponse
		responseIDsOrg1 = append(responseIDsOrg1, response.ID)

		updateCtx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)
		_, err = suite.Client.DB.AssessmentResponse.UpdateOneID(response.ID).
			SetStatus(enums.AssessmentResponseStatusCompleted).
			Save(updateCtx)
		assert.NilError(t, err)

		_, err = suite.Client.API.CreateAssessmentResponse(th.SharedTestUser1.UserCtx, req)
		assert.ErrorContains(t, err, "assessment is already completed")
	})

	(&th.Cleanup[*generated.AssessmentResponseDeleteOne]{Client: suite.Client.DB.AssessmentResponse, IDs: responseIDsOrg1}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: assessment.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, ID: assessment.TemplateID}).MustDelete(th.SharedTestUser1.UserCtx, t)

	(&th.Cleanup[*generated.AssessmentResponseDeleteOne]{Client: suite.Client.DB.AssessmentResponse, IDs: responseIDsOrg2}).MustDelete(th.SharedTestUser2.UserCtx, t)
	(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: assessment2.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, ID: assessment2.TemplateID}).MustDelete(th.SharedTestUser2.UserCtx, t)
}

// TestMutationDeleteAssessmentResponse validates delete mutation behavior.
func TestMutationDeleteAssessmentResponse(t *testing.T) {
	assessment := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	response1 := (&th.AssessmentResponseBuilder{
		Client:       suite.Client,
		AssessmentID: assessment.ID,
		OwnerID:      assessment.OwnerID,
	}).MustNew(th.SharedTestUser1.UserCtx, t)
	response2 := (&th.AssessmentResponseBuilder{
		Client:       suite.Client,
		AssessmentID: assessment.ID,
		OwnerID:      assessment.OwnerID,
	}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, different org user",
			idToDelete:  response1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "not authorized, delete response using view only user",
			idToDelete:  response1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: "you are not authorized to perform this action",
		},
		{
			name:       "happy path, delete response",
			idToDelete: response1.ID,
			client:     suite.Client.API,
			ctx:        th.SharedTestUser1.UserCtx,
		},
		{
			name:        "response already deleted, not found",
			idToDelete:  response1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:       "happy path, delete response using personal access token",
			idToDelete: response2.ID,
			client:     suite.Client.APIWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown response, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteAssessmentResponse(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteAssessmentResponse.DeletedID))
		})
	}

	(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: assessment.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, ID: assessment.TemplateID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}
