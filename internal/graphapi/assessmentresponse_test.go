package graphapi_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

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

func TestMutationCreateAssessmentResponse(t *testing.T) {
	assessment := (&AssessmentBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	t.Run("Create happy path - with questionnaire context", func(t *testing.T) {
		// simulate the questionnaire flow where assessment responses are created
		// by using the QuestionnaireContextKey
		email := gofakeit.Email()
		response := (&AssessmentResponseBuilder{
			client:       suite.client,
			AssessmentID: assessment.ID,
			Email:        email,
			OwnerID:      testUser1.OrganizationID,
		}).MustNew(testUser1.UserCtx, t)

		assert.Assert(t, response != nil)
		assert.Check(t, is.Equal(email, response.Email))
		assert.Check(t, is.Equal(assessment.ID, response.AssessmentID))
		assert.Check(t, is.Equal(testUser1.OrganizationID, response.OwnerID))

		(&Cleanup[*generated.AssessmentResponseDeleteOne]{client: suite.client.db.AssessmentResponse, ID: response.ID}).MustDelete(testUser1.UserCtx, t)
	})

	// no questionnaire context
	// denials are returned as "not found"
	testCases := []struct {
		name     string
		request  testclient.CreateAssessmentResponseInput
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name: "denied - cannot create via GraphQL without questionnaire context",
			request: testclient.CreateAssessmentResponseInput{
				Email:        gofakeit.Email(),
				AssessmentID: assessment.ID,
				OwnerID:      &testUser1.OrganizationID,
			},
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: "assessmentresponse not found",
		},
		{
			name: "denied - cannot create via PAT without questionnaire context",
			request: testclient.CreateAssessmentResponseInput{
				Email:        gofakeit.Email(),
				AssessmentID: assessment.ID,
			},
			client:   suite.client.apiWithPAT,
			ctx:      context.Background(),
			errorMsg: "assessmentresponse not found",
		},
		{
			name: "denied - different org user also cannot create",
			request: testclient.CreateAssessmentResponseInput{
				Email:        gofakeit.Email(),
				AssessmentID: assessment.ID,
				OwnerID:      &testUser1.OrganizationID,
			},
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: "assessmentresponse not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateAssessmentResponse(tc.ctx, tc.request)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
		})
	}

	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, ID: assessment.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: assessment.TemplateID}).MustDelete(testUser1.UserCtx, t)
}

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
