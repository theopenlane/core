package graphapi_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/internal/ent/generated/campaign"
	"github.com/theopenlane/core/internal/ent/generated/campaigntarget"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// TestQueryAssessmentResponse verifies fetching a single assessment response.
func TestQueryAssessmentResponse(t *testing.T) {
	assessment1 := (&AssessmentBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	response1 := (&AssessmentResponseBuilder{
		client:       suite.client,
		AssessmentID: assessment1.ID,
		OwnerID:      assessment1.OwnerID,
	}).MustNew(testUser1.UserCtx, t)

	assessment2 := (&AssessmentBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
	response2 := (&AssessmentResponseBuilder{
		client:       suite.client,
		AssessmentID: assessment2.ID,
		OwnerID:      assessment2.OwnerID,
	}).MustNew(adminUser.UserCtx, t)

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
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
			expectedResult: response1,
		},
		{
			name:           "happy path, response created by admin user",
			queryID:        response2.ID,
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
			expectedResult: response2,
		},
		{
			name:           "happy path using personal access token",
			queryID:        response1.ID,
			client:         suite.client.apiWithPAT,
			ctx:            context.Background(),
			expectedResult: response1,
		},
		{
			name:     "no access, user of different org",
			queryID:  response1.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "not found, invalid ID",
			queryID:  ulids.New().String(),
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
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
			assert.Check(t, is.Equal(tc.expectedResult.Email, resp.AssessmentResponse.Email))
			assert.Check(t, is.Equal(tc.expectedResult.AssessmentID, resp.AssessmentResponse.AssessmentID))
		})
	}

	(&Cleanup[*generated.AssessmentResponseDeleteOne]{client: suite.client.db.AssessmentResponse, IDs: []string{response1.ID, response2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, IDs: []string{assessment1.ID, assessment2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, IDs: []string{assessment1.TemplateID, assessment2.TemplateID}}).MustDelete(testUser1.UserCtx, t)
}

// TestQueryAssessmentResponses verifies listing assessment responses.
func TestQueryAssessmentResponses(t *testing.T) {
	assessment1 := (&AssessmentBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	response1 := (&AssessmentResponseBuilder{
		client:       suite.client,
		AssessmentID: assessment1.ID,
		OwnerID:      assessment1.OwnerID,
	}).MustNew(testUser1.UserCtx, t)

	assessment2 := (&AssessmentBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	response2 := (&AssessmentResponseBuilder{
		client:       suite.client,
		AssessmentID: assessment2.ID,
		OwnerID:      assessment2.OwnerID,
	}).MustNew(testUser1.UserCtx, t)

	anotherUser := suite.userBuilder(context.Background(), t)
	assessment3 := (&AssessmentBuilder{client: suite.client}).MustNew(anotherUser.UserCtx, t)
	response3 := (&AssessmentResponseBuilder{
		client:       suite.client,
		AssessmentID: assessment3.ID,
		OwnerID:      assessment3.OwnerID,
	}).MustNew(anotherUser.UserCtx, t)

	t.Run("Get all assessment responses", func(t *testing.T) {
		resp, err := suite.client.api.GetAllAssessmentResponses(testUser1.UserCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, resp.AssessmentResponses.TotalCount >= 2)
	})

	t.Run("Get assessment responses with where filter", func(t *testing.T) {
		email := response1.Email
		whereInput := &testclient.AssessmentResponseWhereInput{
			Email: &email,
		}

		resp, err := suite.client.api.GetAssessmentResponses(testUser1.UserCtx, nil, nil, whereInput)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Len(resp.AssessmentResponses.Edges, 1))
		assert.Check(t, is.Equal(response1.ID, resp.AssessmentResponses.Edges[0].Node.ID))
	})

	t.Run("Get assessment responses using personal access token", func(t *testing.T) {
		resp, err := suite.client.apiWithPAT.GetAllAssessmentResponses(context.Background())

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, resp.AssessmentResponses.TotalCount >= 2)
	})

	(&Cleanup[*generated.AssessmentResponseDeleteOne]{client: suite.client.db.AssessmentResponse, IDs: []string{response1.ID, response2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, IDs: []string{assessment1.ID, assessment2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, IDs: []string{assessment1.TemplateID, assessment2.TemplateID}}).MustDelete(testUser1.UserCtx, t)

	(&Cleanup[*generated.AssessmentResponseDeleteOne]{client: suite.client.db.AssessmentResponse, ID: response3.ID}).MustDelete(anotherUser.UserCtx, t)
	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, ID: assessment3.ID}).MustDelete(anotherUser.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: assessment3.TemplateID}).MustDelete(anotherUser.UserCtx, t)
}

// TestAssessmentResponseCampaignIsolation ensures responses are isolated per campaign.
func TestAssessmentResponseCampaignIsolation(t *testing.T) {
	assessment := (&AssessmentBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	ctx := setContext(testUser1.UserCtx, suite.client.db)

	campaignA, err := suite.client.db.Campaign.Create().
		SetName("Campaign A").
		SetOwnerID(testUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetRecurrenceFrequency(enums.FrequencyYearly).
		Save(ctx)
	assert.NilError(t, err)

	campaignB, err := suite.client.db.Campaign.Create().
		SetName("Campaign B").
		SetOwnerID(testUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetRecurrenceFrequency(enums.FrequencyYearly).
		Save(ctx)
	assert.NilError(t, err)

	email := gofakeit.Email()
	responseA, err := suite.client.db.AssessmentResponse.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetCampaignID(campaignA.ID).
		SetEmail(email).
		Save(ctx)
	assert.NilError(t, err)

	responseB, err := suite.client.db.AssessmentResponse.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetCampaignID(campaignB.ID).
		SetEmail(email).
		Save(ctx)
	assert.NilError(t, err)

	countA, err := suite.client.db.AssessmentResponse.Query().
		Where(
			assessmentresponse.CampaignIDEQ(campaignA.ID),
			assessmentresponse.EmailEqualFold(email),
		).
		Count(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(1, countA))

	countB, err := suite.client.db.AssessmentResponse.Query().
		Where(
			assessmentresponse.CampaignIDEQ(campaignB.ID),
			assessmentresponse.EmailEqualFold(email),
		).
		Count(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(1, countB))

	(&Cleanup[*generated.AssessmentResponseDeleteOne]{client: suite.client.db.AssessmentResponse, IDs: []string{responseA.ID, responseB.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.CampaignDeleteOne]{client: suite.client.db.Campaign, IDs: []string{campaignA.ID, campaignB.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, ID: assessment.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: assessment.TemplateID}).MustDelete(testUser1.UserCtx, t)
}

// TestAssessmentResponseUpdatesCampaignTargetsAndCompletion verifies campaign rollups on completion.
func TestAssessmentResponseUpdatesCampaignTargetsAndCompletion(t *testing.T) {
	assessment := (&AssessmentBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	ctx := setContext(testUser1.UserCtx, suite.client.db)

	campaignObj, err := suite.client.db.Campaign.Create().
		SetName("Campaign Target Sync").
		SetOwnerID(testUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetRecurrenceFrequency(enums.FrequencyYearly).
		Save(ctx)
	assert.NilError(t, err)

	emails := []string{"alpha@example.com", "beta@example.com"}
	targetIDs := make([]string, 0, len(emails))
	for _, email := range emails {
		target, err := suite.client.db.CampaignTarget.Create().
			SetOwnerID(testUser1.OrganizationID).
			SetCampaignID(campaignObj.ID).
			SetEmail(email).
			Save(ctx)
		assert.NilError(t, err)
		targetIDs = append(targetIDs, target.ID)
	}

	responseA, err := suite.client.db.AssessmentResponse.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetCampaignID(campaignObj.ID).
		SetEmail(emails[0]).
		Save(ctx)
	assert.NilError(t, err)

	responseB, err := suite.client.db.AssessmentResponse.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetCampaignID(campaignObj.ID).
		SetEmail(emails[1]).
		Save(ctx)
	assert.NilError(t, err)

	_, err = suite.client.db.AssessmentResponse.UpdateOneID(responseA.ID).
		SetStatus(enums.AssessmentResponseStatusCompleted).
		Save(ctx)
	assert.NilError(t, err)

	targetA, err := suite.client.db.CampaignTarget.Query().
		Where(
			campaigntarget.CampaignIDEQ(campaignObj.ID),
			campaigntarget.EmailEqualFold(emails[0]),
		).
		Only(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(enums.AssessmentResponseStatusCompleted, targetA.Status))

	targetB, err := suite.client.db.CampaignTarget.Query().
		Where(
			campaigntarget.CampaignIDEQ(campaignObj.ID),
			campaigntarget.EmailEqualFold(emails[1]),
		).
		Only(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(enums.AssessmentResponseStatusNotStarted, targetB.Status))

	campaignAfterFirst, err := suite.client.db.Campaign.Query().
		Where(campaign.IDEQ(campaignObj.ID)).
		Only(ctx)
	assert.NilError(t, err)
	assert.Check(t, campaignAfterFirst.Status != enums.CampaignStatusCompleted)

	_, err = suite.client.db.AssessmentResponse.UpdateOneID(responseB.ID).
		SetStatus(enums.AssessmentResponseStatusCompleted).
		Save(ctx)
	assert.NilError(t, err)

	campaignAfterSecond, err := suite.client.db.Campaign.Query().
		Where(campaign.IDEQ(campaignObj.ID)).
		Only(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(enums.CampaignStatusCompleted, campaignAfterSecond.Status))
	assert.Check(t, !campaignAfterSecond.IsActive)

	(&Cleanup[*generated.AssessmentResponseDeleteOne]{client: suite.client.db.AssessmentResponse, IDs: []string{responseA.ID, responseB.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.CampaignTargetDeleteOne]{client: suite.client.db.CampaignTarget, IDs: targetIDs}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.CampaignDeleteOne]{client: suite.client.db.Campaign, ID: campaignObj.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, ID: assessment.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: assessment.TemplateID}).MustDelete(testUser1.UserCtx, t)
}

// TestMutationCreateAssessmentResponse validates create mutation behavior.
func TestMutationCreateAssessmentResponse(t *testing.T) {
	assessment := (&AssessmentBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	assessment2 := (&AssessmentBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name    string
		request testclient.CreateAssessmentResponseInput
		client  *testclient.TestClient
		ctx     context.Context
	}{
		{
			name: "success - can create via GraphQL",
			request: testclient.CreateAssessmentResponseInput{
				Email:        gofakeit.Email(),
				AssessmentID: assessment.ID,
				OwnerID:      &testUser1.OrganizationID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "success - can create via PAT",
			request: testclient.CreateAssessmentResponseInput{
				Email:        gofakeit.Email(),
				AssessmentID: assessment.ID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "success - different org user can create",
			request: testclient.CreateAssessmentResponseInput{
				Email:        gofakeit.Email(),
				AssessmentID: assessment2.ID,
				OwnerID:      &testUser2.OrganizationID,
			},
			client: suite.client.api,
			ctx:    testUser2.UserCtx,
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

			if tc.ctx == testUser2.UserCtx {
				responseIDsOrg2 = append(responseIDsOrg2, resp.CreateAssessmentResponse.AssessmentResponse.ID)
			} else {
				responseIDsOrg1 = append(responseIDsOrg1, resp.CreateAssessmentResponse.AssessmentResponse.ID)
			}
		})
	}

	t.Run("send attempts should increment on duplicate response", func(t *testing.T) {
		email := gofakeit.Email()
		req := testclient.CreateAssessmentResponseInput{
			Email:        email,
			AssessmentID: assessment.ID,
			OwnerID:      &testUser1.OrganizationID,
		}

		resp, err := suite.client.api.CreateAssessmentResponse(testUser1.UserCtx, req)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		firstResponse := resp.CreateAssessmentResponse.AssessmentResponse
		assert.Check(t, is.Equal(int64(1), firstResponse.SendAttempts))
		responseIDsOrg1 = append(responseIDsOrg1, firstResponse.ID)

		secondResp, err := suite.client.api.CreateAssessmentResponse(testUser1.UserCtx, req)
		assert.NilError(t, err)
		assert.Assert(t, secondResp != nil)

		updatedResponse := secondResp.CreateAssessmentResponse.AssessmentResponse
		assert.Check(t, is.Equal(firstResponse.ID, updatedResponse.ID))
		assert.Check(t, is.Equal(firstResponse.SendAttempts+1, updatedResponse.SendAttempts))
	})

	t.Run("draft response should have draft status", func(t *testing.T) {
		email := gofakeit.Email()
		isDraft := true
		req := testclient.CreateAssessmentResponseInput{
			Email:        email,
			AssessmentID: assessment.ID,
			OwnerID:      &testUser1.OrganizationID,
			IsDraft:      &isDraft,
		}

		resp, err := suite.client.api.CreateAssessmentResponse(testUser1.UserCtx, req)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		response := resp.CreateAssessmentResponse.AssessmentResponse
		assert.Check(t, is.Equal(enums.AssessmentResponseStatusDraft, response.Status))
		responseIDsOrg1 = append(responseIDsOrg1, response.ID)
	})

	t.Run("draft on existing draft should succeed", func(t *testing.T) {
		email := gofakeit.Email()
		isDraft := true
		req := testclient.CreateAssessmentResponseInput{
			Email:        email,
			AssessmentID: assessment.ID,
			OwnerID:      &testUser1.OrganizationID,
			IsDraft:      &isDraft,
		}

		resp, err := suite.client.api.CreateAssessmentResponse(testUser1.UserCtx, req)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		firstResponse := resp.CreateAssessmentResponse.AssessmentResponse
		assert.Check(t, is.Equal(enums.AssessmentResponseStatusDraft, firstResponse.Status))
		responseIDsOrg1 = append(responseIDsOrg1, firstResponse.ID)

		secondResp, err := suite.client.api.CreateAssessmentResponse(testUser1.UserCtx, req)
		assert.NilError(t, err)
		assert.Assert(t, secondResp != nil)

		updatedResponse := secondResp.CreateAssessmentResponse.AssessmentResponse
		assert.Check(t, is.Equal(firstResponse.ID, updatedResponse.ID))
		assert.Check(t, is.Equal(enums.AssessmentResponseStatusDraft, updatedResponse.Status))
	})

	t.Run("draft on sent response should fail", func(t *testing.T) {
		email := gofakeit.Email()
		req := testclient.CreateAssessmentResponseInput{
			Email:        email,
			AssessmentID: assessment.ID,
			OwnerID:      &testUser1.OrganizationID,
		}

		resp, err := suite.client.api.CreateAssessmentResponse(testUser1.UserCtx, req)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		response := resp.CreateAssessmentResponse.AssessmentResponse
		responseIDsOrg1 = append(responseIDsOrg1, response.ID)

		isDraft := true
		draftReq := testclient.CreateAssessmentResponseInput{
			Email:        email,
			AssessmentID: assessment.ID,
			OwnerID:      &testUser1.OrganizationID,
			IsDraft:      &isDraft,
		}

		_, err = suite.client.api.CreateAssessmentResponse(testUser1.UserCtx, draftReq)
		assert.ErrorContains(t, err, "assessment is already in progress or completed")
	})

	t.Run("draft on completed response should fail", func(t *testing.T) {
		email := gofakeit.Email()
		req := testclient.CreateAssessmentResponseInput{
			Email:        email,
			AssessmentID: assessment.ID,
			OwnerID:      &testUser1.OrganizationID,
		}

		resp, err := suite.client.api.CreateAssessmentResponse(testUser1.UserCtx, req)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		response := resp.CreateAssessmentResponse.AssessmentResponse
		responseIDsOrg1 = append(responseIDsOrg1, response.ID)

		updateCtx := setContext(testUser1.UserCtx, suite.client.db)
		_, err = suite.client.db.AssessmentResponse.UpdateOneID(response.ID).
			SetStatus(enums.AssessmentResponseStatusCompleted).
			Save(updateCtx)
		assert.NilError(t, err)

		isDraft := true
		draftReq := testclient.CreateAssessmentResponseInput{
			Email:        email,
			AssessmentID: assessment.ID,
			OwnerID:      &testUser1.OrganizationID,
			IsDraft:      &isDraft,
		}

		_, err = suite.client.api.CreateAssessmentResponse(testUser1.UserCtx, draftReq)
		assert.ErrorContains(t, err, "assessment is already in progress or completed")
	})

	t.Run("non-draft on existing draft should send and bump attempts", func(t *testing.T) {
		email := gofakeit.Email()
		isDraft := true
		draftReq := testclient.CreateAssessmentResponseInput{
			Email:        email,
			AssessmentID: assessment.ID,
			OwnerID:      &testUser1.OrganizationID,
			IsDraft:      &isDraft,
		}

		resp, err := suite.client.api.CreateAssessmentResponse(testUser1.UserCtx, draftReq)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		draftResponse := resp.CreateAssessmentResponse.AssessmentResponse
		assert.Check(t, is.Equal(enums.AssessmentResponseStatusDraft, draftResponse.Status))
		responseIDsOrg1 = append(responseIDsOrg1, draftResponse.ID)

		sendReq := testclient.CreateAssessmentResponseInput{
			Email:        email,
			AssessmentID: assessment.ID,
			OwnerID:      &testUser1.OrganizationID,
		}

		sendResp, err := suite.client.api.CreateAssessmentResponse(testUser1.UserCtx, sendReq)
		assert.NilError(t, err)
		assert.Assert(t, sendResp != nil)

		sentResponse := sendResp.CreateAssessmentResponse.AssessmentResponse
		assert.Check(t, is.Equal(draftResponse.ID, sentResponse.ID))
		assert.Check(t, is.Equal(int64(2), sentResponse.SendAttempts))
	})

	t.Run("completed response should not be updated", func(t *testing.T) {
		email := gofakeit.Email()
		req := testclient.CreateAssessmentResponseInput{
			Email:        email,
			AssessmentID: assessment.ID,
			OwnerID:      &testUser1.OrganizationID,
		}

		resp, err := suite.client.api.CreateAssessmentResponse(testUser1.UserCtx, req)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		response := resp.CreateAssessmentResponse.AssessmentResponse
		responseIDsOrg1 = append(responseIDsOrg1, response.ID)

		updateCtx := setContext(testUser1.UserCtx, suite.client.db)
		_, err = suite.client.db.AssessmentResponse.UpdateOneID(response.ID).
			SetStatus(enums.AssessmentResponseStatusCompleted).
			Save(updateCtx)
		assert.NilError(t, err)

		_, err = suite.client.api.CreateAssessmentResponse(testUser1.UserCtx, req)
		assert.ErrorContains(t, err, "assessment is already in progress or completed")
	})

	(&Cleanup[*generated.AssessmentResponseDeleteOne]{client: suite.client.db.AssessmentResponse, IDs: responseIDsOrg1}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, ID: assessment.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: assessment.TemplateID}).MustDelete(testUser1.UserCtx, t)

	(&Cleanup[*generated.AssessmentResponseDeleteOne]{client: suite.client.db.AssessmentResponse, IDs: responseIDsOrg2}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, ID: assessment2.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: assessment2.TemplateID}).MustDelete(testUser2.UserCtx, t)
}

// TestMutationDeleteAssessmentResponse validates delete mutation behavior.
func TestMutationDeleteAssessmentResponse(t *testing.T) {
	assessment := (&AssessmentBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	response1 := (&AssessmentResponseBuilder{
		client:       suite.client,
		AssessmentID: assessment.ID,
		OwnerID:      assessment.OwnerID,
	}).MustNew(testUser1.UserCtx, t)
	response2 := (&AssessmentResponseBuilder{
		client:       suite.client,
		AssessmentID: assessment.ID,
		OwnerID:      assessment.OwnerID,
	}).MustNew(testUser1.UserCtx, t)

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
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "not authorized, delete response using view only user",
			idToDelete:  response1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: "you are not authorized to perform this action",
		},
		{
			name:       "happy path, delete response",
			idToDelete: response1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "response already deleted, not found",
			idToDelete:  response1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete response using personal access token",
			idToDelete: response2.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown response, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
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

	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, ID: assessment.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: assessment.TemplateID}).MustDelete(testUser1.UserCtx, t)
}
