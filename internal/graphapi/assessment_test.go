package graphapi_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/shared/enums"
	"github.com/theopenlane/utils/ulids"
)

func TestQueryAssessment(t *testing.T) {
	assessment1 := (&AssessmentBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	assessment2 := (&AssessmentBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)

	testCases := []struct {
		name           string
		queryID        string
		client         *testclient.TestClient
		ctx            context.Context
		expectedResult *generated.Assessment
		errorMsg       string
	}{
		{
			name:           "happy path",
			queryID:        assessment1.ID,
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
			expectedResult: assessment1,
		},
		{
			name:           "happy path, assessment created by admin user",
			queryID:        assessment2.ID,
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
			expectedResult: assessment2,
		},
		{
			name:           "happy path using personal access token",
			queryID:        assessment1.ID,
			client:         suite.client.apiWithPAT,
			ctx:            context.Background(),
			expectedResult: assessment1,
		},
		{
			name:     "no access, user of different org",
			queryID:  assessment1.ID,
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
			resp, err := tc.client.GetAssessmentByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.expectedResult.ID, resp.Assessment.ID))
			assert.Check(t, is.Equal(tc.expectedResult.Name, resp.Assessment.Name))
			assert.Check(t, is.Equal(string(tc.expectedResult.AssessmentType), string(resp.Assessment.AssessmentType)))
		})
	}

	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, IDs: []string{assessment1.ID, assessment2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, IDs: []string{assessment1.TemplateID, assessment2.TemplateID}}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryAssessments(t *testing.T) {
	// assessments for the first organization
	assessment1 := (&AssessmentBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	assessment2 := (&AssessmentBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// assessment created by an admin user of the first organization
	assessment3 := (&AssessmentBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)

	// assessment for another organization
	anotherUser := suite.userBuilder(context.Background(), t)
	assessment4 := (&AssessmentBuilder{client: suite.client}).MustNew(anotherUser.UserCtx, t)

	t.Run("Get all assessments", func(t *testing.T) {
		resp, err := suite.client.api.GetAllAssessments(testUser1.UserCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		// should return at least the 3 assessments created by testUser1's organization
		assert.Check(t, resp.Assessments.TotalCount >= 3)
	})

	t.Run("Get assessments with where filter", func(t *testing.T) {
		whereInput := &testclient.AssessmentWhereInput{
			Name: &assessment1.Name,
		}

		resp, err := suite.client.api.GetAssessments(testUser1.UserCtx, lo.ToPtr(int64(10)), nil, whereInput)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Len(resp.Assessments.Edges, 1))
		assert.Check(t, is.Equal(assessment1.ID, resp.Assessments.Edges[0].Node.ID))
	})

	t.Run("Get assessments using personal access token", func(t *testing.T) {
		resp, err := suite.client.apiWithPAT.GetAllAssessments(context.Background())

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, resp.Assessments.TotalCount >= 3)
	})

	t.Run("Get assessments with pagination", func(t *testing.T) {
		resp, err := suite.client.api.GetAssessments(testUser1.UserCtx, lo.ToPtr(int64(2)), nil, nil)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Len(resp.Assessments.Edges, 2))
		assert.Check(t, resp.Assessments.PageInfo.HasNextPage)
		assert.Assert(t, resp.Assessments.PageInfo.EndCursor != nil)
	})

	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, IDs: []string{assessment1.ID, assessment2.ID, assessment3.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, IDs: []string{assessment4.ID}}).MustDelete(anotherUser.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, IDs: []string{assessment1.TemplateID, assessment2.TemplateID, assessment3.TemplateID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, IDs: []string{assessment4.TemplateID}}).MustDelete(anotherUser.UserCtx, t)
}

func TestMutationCreateAssessment(t *testing.T) {
	template := (&TemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	jsonConfig := map[string]any{
		"title":       "Test Assessment Template Missing",
		"description": "A test questionnaire template that will be deleted",
		"questions": []map[string]any{
			{
				"id":       "q1",
				"question": "What is your name?",
				"type":     "text",
			},
		},
	}

	testCases := []struct {
		name     string
		request  testclient.CreateAssessmentInput
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name: "happy path, minimal fields",
			request: testclient.CreateAssessmentInput{
				Name:       gofakeit.Company(),
				TemplateID: lo.ToPtr(template.ID),
				OwnerID:    &testUser1.OrganizationID,
				Jsonconfig: jsonConfig,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all fields",
			request: testclient.CreateAssessmentInput{
				Name:                gofakeit.Company(),
				TemplateID:          lo.ToPtr(template.ID),
				OwnerID:             &testUser1.OrganizationID,
				AssessmentType:      lo.ToPtr(enums.AssessmentTypeInternal),
				Tags:                []string{"tag1", "tag2"},
				ResponseDueDuration: lo.ToPtr(int64(86400)), // 1 day
				Jsonconfig:          jsonConfig,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path using personal access token",
			request: testclient.CreateAssessmentInput{
				Name:       gofakeit.Company(),
				TemplateID: lo.ToPtr(template.ID),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "missing required field - jsonconfig",
			request: testclient.CreateAssessmentInput{
				OwnerID: &testUser1.OrganizationID,
				Name:    "another assessment",
			},
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: "jsonconfig is required",
		},
		{
			name: "missing required field - name",
			request: testclient.CreateAssessmentInput{
				TemplateID: lo.ToPtr(template.ID),
				OwnerID:    &testUser1.OrganizationID,
				Jsonconfig: jsonConfig,
			},
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: "value is less than the required length",
		},
		{
			name: "invalid template_id",
			request: testclient.CreateAssessmentInput{
				Name:       gofakeit.Company(),
				TemplateID: lo.ToPtr(ulids.New().String()),
				OwnerID:    &testUser1.OrganizationID,
				Jsonconfig: jsonConfig,
			},
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: "do not have permission to perform this action",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateAssessment(tc.ctx, tc.request)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.request.Name, resp.CreateAssessment.Assessment.Name))

			if tc.request.AssessmentType != nil {
				assert.Check(t, is.Equal(string(*tc.request.AssessmentType), string(resp.CreateAssessment.Assessment.AssessmentType)))
			}

			if len(tc.request.Tags) > 0 {
				assert.Check(t, is.Len(resp.CreateAssessment.Assessment.Tags, len(tc.request.Tags)))
			}

			(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, ID: resp.CreateAssessment.Assessment.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}

	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: template.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateAssessment(t *testing.T) {

	jsonConfig := map[string]any{
		"title":       "Test Assessment Template Missing",
		"description": "A test questionnaire template that will be deleted",
		"questions": []map[string]any{
			{
				"id":       "q1",
				"question": "What is your name?",
				"type":     "text",
			},
		},
	}

	assessment := (&AssessmentBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	templateIDPtr := lo.ToPtr(assessment.TemplateID)

	testCases := []struct {
		name     string
		id       string
		request  testclient.UpdateAssessmentInput
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name: "happy path, update name",
			id:   assessment.ID,
			request: testclient.UpdateAssessmentInput{
				Name:       lo.ToPtr(gofakeit.Company()),
				TemplateID: templateIDPtr,
				Jsonconfig: jsonConfig,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update tags",
			id:   assessment.ID,
			request: testclient.UpdateAssessmentInput{
				Tags:       []string{"updated-tag1", "updated-tag2"},
				TemplateID: templateIDPtr,
				Jsonconfig: jsonConfig,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, append tags",
			id:   assessment.ID,
			request: testclient.UpdateAssessmentInput{
				AppendTags: []string{"appended-tag"},
				TemplateID: templateIDPtr,
				Jsonconfig: jsonConfig,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path using personal access token",
			id:   assessment.ID,
			request: testclient.UpdateAssessmentInput{
				Name:       lo.ToPtr(gofakeit.Company()),
				TemplateID: templateIDPtr,
				Jsonconfig: jsonConfig,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "not found, invalid ID",
			id:   ulids.New().String(),
			request: testclient.UpdateAssessmentInput{
				Name:       lo.ToPtr(gofakeit.Company()),
				TemplateID: templateIDPtr,
				Jsonconfig: jsonConfig,
			},
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name: "no access, user of different org",
			id:   assessment.ID,
			request: testclient.UpdateAssessmentInput{
				Name:       lo.ToPtr(gofakeit.Company()),
				Jsonconfig: jsonConfig,
				TemplateID: templateIDPtr,
			},
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateAssessment(tc.ctx, tc.id, tc.request)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.id, resp.UpdateAssessment.Assessment.ID))

			if tc.request.Name != nil {
				assert.Check(t, is.Equal(*tc.request.Name, resp.UpdateAssessment.Assessment.Name))
			}

			if len(tc.request.Tags) > 0 {
				assert.Check(t, is.Len(resp.UpdateAssessment.Assessment.Tags, len(tc.request.Tags)))
			}

			if len(tc.request.AppendTags) > 0 {
				assert.Check(t, len(resp.UpdateAssessment.Assessment.Tags) >= len(tc.request.AppendTags))
			}
		})
	}

	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, ID: assessment.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: assessment.TemplateID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteAssessment(t *testing.T) {
	assessment1 := (&AssessmentBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	assessment2 := (&AssessmentBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete assessment",
			idToDelete:  assessment1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "not authorized, view only user",
			idToDelete:  assessment1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: "you are not authorized to perform this action",
		},
		{
			name:       "happy path, delete assessment",
			idToDelete: assessment1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "assessment already deleted, not found",
			idToDelete:  assessment1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete assessment using personal access token",
			idToDelete: assessment2.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown assessment, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteAssessment(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteAssessment.DeletedID))
		})
	}

	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, IDs: []string{assessment1.TemplateID, assessment2.TemplateID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateAssessmentWithDuplicateName(t *testing.T) {

	assessment1 := (&AssessmentBuilder{client: suite.client, Name: "Duplicate Test"}).MustNew(testUser1.UserCtx, t)

	t.Run("duplicate name in same org should fail", func(t *testing.T) {
		template := (&TemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

		request := testclient.CreateAssessmentInput{
			Name:       "Duplicate Test",
			TemplateID: lo.ToPtr(template.ID),
			OwnerID:    &testUser1.OrganizationID,
		}

		_, err := suite.client.api.CreateAssessment(testUser1.UserCtx, request)
		assert.ErrorContains(t, err, "assessment already exists")

		(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: template.ID}).MustDelete(testUser1.UserCtx, t)
	})

	// assessment with same name in different org should succeed
	t.Run("duplicate name in different org should succeed", func(t *testing.T) {
		anotherUser := suite.userBuilder(context.Background(), t)
		template := (&TemplateBuilder{client: suite.client}).MustNew(anotherUser.UserCtx, t)

		request := testclient.CreateAssessmentInput{
			Name:       "Duplicate Test",
			TemplateID: lo.ToPtr(template.ID),
			OwnerID:    &anotherUser.OrganizationID,
		}

		resp, err := suite.client.api.CreateAssessment(anotherUser.UserCtx, request)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal("Duplicate Test", resp.CreateAssessment.Assessment.Name))

		(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, ID: resp.CreateAssessment.Assessment.ID}).MustDelete(anotherUser.UserCtx, t)
		(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: template.ID}).MustDelete(anotherUser.UserCtx, t)
	})

	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, ID: assessment1.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: assessment1.TemplateID}).MustDelete(testUser1.UserCtx, t)
}
