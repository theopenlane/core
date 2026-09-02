package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestQueryAssessment(t *testing.T) {
	assessment1 := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	assessment2 := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)

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
			client:         suite.Client.API,
			ctx:            th.SharedTestUser1.UserCtx,
			expectedResult: assessment1,
		},
		{
			name:           "happy path, assessment created by admin user",
			queryID:        assessment2.ID,
			client:         suite.Client.API,
			ctx:            th.SharedTestUser1.UserCtx,
			expectedResult: assessment2,
		},
		{
			name:           "happy path using personal access token",
			queryID:        assessment1.ID,
			client:         suite.Client.APIWithPAT,
			ctx:            context.Background(),
			expectedResult: assessment1,
		},
		{
			name:     "no access, user of different org",
			queryID:  assessment1.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "no access, api token of different org",
			queryID:  assessment1.ID,
			client:   suite.Client.APIWithTokenOrg2,
			ctx:      context.Background(),
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

	(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, IDs: []string{assessment1.ID, assessment2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, IDs: []string{assessment1.TemplateID, assessment2.TemplateID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryAssessments(t *testing.T) {
	// assessments for the first organization
	assessment1 := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	assessment2 := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// assessment created by an admin user of the first organization
	assessment3 := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)

	// assessment for another organization
	anotherUser := suite.UserBuilder(context.Background(), t)
	(&th.AssessmentBuilder{Client: suite.Client}).MustNew(anotherUser.UserCtx, t)

	t.Run("Get all assessments", func(t *testing.T) {
		resp, err := suite.Client.API.GetAllAssessments(th.SharedTestUser1.UserCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		// should return at least the 3 assessments created by th.SharedTestUser1's organization
		assert.Check(t, resp.Assessments.TotalCount >= 3)
	})

	t.Run("Get assessments with where filter", func(t *testing.T) {
		whereInput := &testclient.AssessmentWhereInput{
			Name: &assessment1.Name,
		}

		resp, err := suite.Client.API.GetAssessments(th.SharedTestUser1.UserCtx, lo.ToPtr(int64(10)), nil, whereInput)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Len(resp.Assessments.Edges, 1))
		assert.Check(t, is.Equal(assessment1.ID, resp.Assessments.Edges[0].Node.ID))
	})

	t.Run("Get assessments using personal access token", func(t *testing.T) {
		resp, err := suite.Client.APIWithPAT.GetAllAssessments(context.Background())

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, resp.Assessments.TotalCount >= 3)
	})

	t.Run("Get assessments with pagination", func(t *testing.T) {
		resp, err := suite.Client.API.GetAssessments(th.SharedTestUser1.UserCtx, lo.ToPtr(int64(2)), nil, nil)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Len(resp.Assessments.Edges, 2))
		assert.Check(t, resp.Assessments.PageInfo.HasNextPage)
		assert.Assert(t, resp.Assessments.PageInfo.EndCursor != nil)
	})

	(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, IDs: []string{assessment1.ID, assessment2.ID, assessment3.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, IDs: []string{assessment1.TemplateID, assessment2.TemplateID, assessment3.TemplateID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	th.CleanupOrganizationDataWithContext(anotherUser.UserCtx, t)
}

func TestMutationCreateAssessment(t *testing.T) {
	template := (&th.TemplateBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
				OwnerID:    &th.SharedTestUser1.OrganizationID,
				Jsonconfig: jsonConfig,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, all fields",
			request: testclient.CreateAssessmentInput{
				Name:                gofakeit.Company(),
				TemplateID:          lo.ToPtr(template.ID),
				OwnerID:             &th.SharedTestUser1.OrganizationID,
				AssessmentType:      lo.ToPtr(enums.AssessmentTypeInternal),
				Tags:                []string{"tag1", "tag2"},
				ResponseDueDuration: lo.ToPtr(int64(86400)), // 1 day
				Jsonconfig:          jsonConfig,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path using personal access token",
			request: testclient.CreateAssessmentInput{
				Name:       gofakeit.Company(),
				TemplateID: lo.ToPtr(template.ID),
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "missing required field - jsonconfig",
			request: testclient.CreateAssessmentInput{
				OwnerID: &th.SharedTestUser1.OrganizationID,
				Name:    "another assessment",
			},
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: "jsonconfig is required",
		},
		{
			name: "missing required field - name",
			request: testclient.CreateAssessmentInput{
				TemplateID: lo.ToPtr(template.ID),
				OwnerID:    &th.SharedTestUser1.OrganizationID,
				Jsonconfig: jsonConfig,
			},
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: "value is less than the required length",
		},
		{
			name: "invalid template_id",
			request: testclient.CreateAssessmentInput{
				Name:       gofakeit.Company(),
				TemplateID: lo.ToPtr(ulids.New().String()),
				OwnerID:    &th.SharedTestUser1.OrganizationID,
				Jsonconfig: jsonConfig,
			},
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotAuthorizedErrorMsg,
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

			(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: resp.CreateAssessment.Assessment.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		})
	}

	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, ID: template.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationTemplateFromAssessment(t *testing.T) {
	assess := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	newName := gofakeit.Company() + "-" + ulids.New().String()
	fullName := gofakeit.Company() + "-" + ulids.New().String()

	description := gofakeit.Sentence(10)

	tags := []string{"cloned", "template"}

	cases := []struct {
		name                string
		input               testclient.CreateAssessmentTemplateInput
		expectedName        string
		expectedDescription *string
		expectedTags        []string
		errorMsg            string
	}{
		{
			name: "successfully created template from existing assessment",
			input: testclient.CreateAssessmentTemplateInput{
				AssessmentID: assess.ID,
			},
			expectedName: assess.Name,
		},
		{
			name: "creating template from assessment fails if name already exists",
			input: testclient.CreateAssessmentTemplateInput{
				AssessmentID: assess.ID,
			},
			errorMsg: "already exists",
		},
		{
			name: "creating template from existing assessment can use alternate name",
			input: testclient.CreateAssessmentTemplateInput{
				AssessmentID: assess.ID,
				Name:         lo.ToPtr(newName),
			},
			expectedName: newName,
		},
		{
			name: "creating template from existing assessment can use alternate description and tags",
			input: testclient.CreateAssessmentTemplateInput{
				AssessmentID: assess.ID,
				Name:         lo.ToPtr(fullName),
				Description:  lo.ToPtr(description),
				Tags:         tags,
			},
			expectedName:        fullName,
			expectedDescription: lo.ToPtr(description),
			expectedTags:        tags,
		},
	}

	templateIDs := []string{}

	for _, tc := range cases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := suite.Client.API.CreateAssessmentTemplate(th.SharedTestUser1.UserCtx, tc.input)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			template := resp.CreateAssessmentTemplate.Template
			templateIDs = append(templateIDs, template.ID)

			assert.Check(t, is.Equal(tc.expectedName, template.Name))
			assert.Check(t, is.Equal(enums.Document, template.TemplateType))
			assert.Assert(t, template.Kind != nil)
			assert.Check(t, is.Equal(enums.TemplateKindQuestionnaire, *template.Kind))

			if tc.expectedDescription != nil {
				assert.Assert(t, template.Description != nil)
				assert.Check(t, is.Equal(*tc.expectedDescription, *template.Description))
			}

			if len(tc.expectedTags) > 0 {
				assert.Check(t, is.DeepEqual(tc.expectedTags, template.Tags))
			}
		})
	}

	if len(templateIDs) > 0 {
		(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, IDs: templateIDs}).MustDelete(th.SharedTestUser1.UserCtx, t)
	}

	(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: assess.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, ID: assess.TemplateID}).MustDelete(th.SharedTestUser1.UserCtx, t)
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

	assessment := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
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
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, update due date",
			id:   assessment.ID,
			request: testclient.UpdateAssessmentInput{
				ResponseDueDuration: lo.ToPtr(int64(86400)), // 1 day,
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name: "happy path, update tags",
			id:   assessment.ID,
			request: testclient.UpdateAssessmentInput{
				Tags:       []string{"updated-tag1", "updated-tag2"},
				TemplateID: templateIDPtr,
				Jsonconfig: jsonConfig,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, append tags",
			id:   assessment.ID,
			request: testclient.UpdateAssessmentInput{
				AppendTags: []string{"appended-tag"},
				TemplateID: templateIDPtr,
				Jsonconfig: jsonConfig,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path using personal access token",
			id:   assessment.ID,
			request: testclient.UpdateAssessmentInput{
				Name:       lo.ToPtr(gofakeit.Company()),
				TemplateID: templateIDPtr,
				Jsonconfig: jsonConfig,
			},
			client: suite.Client.APIWithPAT,
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
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name: "no access, user of different org",
			id:   assessment.ID,
			request: testclient.UpdateAssessmentInput{
				Name:       lo.ToPtr(gofakeit.Company()),
				Jsonconfig: jsonConfig,
				TemplateID: templateIDPtr,
			},
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
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

			if tc.request.ResponseDueDuration != nil {
				assert.Check(t, is.Equal(*resp.UpdateAssessment.Assessment.ResponseDueDuration, *tc.request.ResponseDueDuration))
			}
		})
	}

	(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: assessment.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, ID: assessment.TemplateID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteAssessment(t *testing.T) {
	assessment1 := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	assessment2 := (&th.AssessmentBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "not authorized, view only user",
			idToDelete:  assessment1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: "you are not authorized to perform this action",
		},
		{
			name:       "happy path, delete assessment",
			idToDelete: assessment1.ID,
			client:     suite.Client.API,
			ctx:        th.SharedTestUser1.UserCtx,
		},
		{
			name:        "assessment already deleted, not found",
			idToDelete:  assessment1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete assessment using personal access token",
			idToDelete: assessment2.ID,
			client:     suite.Client.APIWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown assessment, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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

	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, IDs: []string{assessment1.TemplateID, assessment2.TemplateID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationCreateAssessmentWithDuplicateName(t *testing.T) {
	assessment1 := (&th.AssessmentBuilder{Client: suite.Client, Name: "Duplicate Test"}).MustNew(th.SharedTestUser1.UserCtx, t)

	t.Run("duplicate name in same org should be allowed", func(t *testing.T) {
		template := (&th.TemplateBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

		request := testclient.CreateAssessmentInput{
			Name:       "Duplicate Test",
			TemplateID: lo.ToPtr(template.ID),
			OwnerID:    &th.SharedTestUser1.OrganizationID,
		}

		resp, err := suite.Client.API.CreateAssessment(th.SharedTestUser1.UserCtx, request)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal("Duplicate Test", resp.CreateAssessment.Assessment.Name))

		(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: resp.CreateAssessment.Assessment.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, ID: template.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	// assessment with same name in different org should succeed
	t.Run("duplicate name in different org should succeed", func(t *testing.T) {
		anotherUser := suite.UserBuilder(context.Background(), t)
		template := (&th.TemplateBuilder{Client: suite.Client}).MustNew(anotherUser.UserCtx, t)

		request := testclient.CreateAssessmentInput{
			Name:       "Duplicate Test",
			TemplateID: lo.ToPtr(template.ID),
			OwnerID:    &anotherUser.OrganizationID,
		}

		resp, err := suite.Client.API.CreateAssessment(anotherUser.UserCtx, request)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal("Duplicate Test", resp.CreateAssessment.Assessment.Name))

		th.CleanupOrganizationDataWithContext(anotherUser.UserCtx, t)
	})

	(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: assessment1.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, ID: assessment1.TemplateID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}
