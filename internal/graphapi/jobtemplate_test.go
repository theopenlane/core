package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/shared/enums"
	"github.com/theopenlane/utils/ulids"
)

func TestQueryJobTemplate(t *testing.T) {
	// create an jobTemplate to be queried using testUser1
	jobTemplate := (&JobTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add test cases for querying the JobTemplate
	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: jobTemplate.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: jobTemplate.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: jobTemplate.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "JobTemplate not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "JobTemplate not found, using not authorized user",
			queryID:  jobTemplate.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetJobTemplateByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.JobTemplate.ID))
		})
	}

	(&Cleanup[*generated.JobTemplateDeleteOne]{client: suite.client.db.JobTemplate, ID: jobTemplate.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryJobTemplates(t *testing.T) {
	// create multiple JobTemplates to be queried using testUser1
	jobTemplate1 := (&JobTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	jobTemplate2 := (&JobTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	jobTemplateSystem := (&JobTemplateBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 3,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 3,
		},
		{
			name:            "happy path, using api token",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: 3,
		},
		{
			name:            "happy path, using pat",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 3,
		},
		{
			name:            "another user, only system owned JobTemplates should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllJobTemplates(tc.ctx)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.JobTemplates.Edges, tc.expectedResults))
		})
	}

	(&Cleanup[*generated.JobTemplateDeleteOne]{client: suite.client.db.JobTemplate, IDs: []string{jobTemplate1.ID, jobTemplate2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.JobTemplateDeleteOne]{client: suite.client.db.JobTemplate, ID: jobTemplateSystem.ID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationCreateJobTemplate(t *testing.T) {
	testCases := []struct {
		name        string
		request     testclient.CreateJobTemplateInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateJobTemplateInput{
				Title:       "Test Job Template",
				Platform:    enums.JobPlatformTypeGo,
				DownloadURL: testScriptURL,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateJobTemplateInput{
				Title:       "Test Job Template",
				Description: lo.ToPtr("Test Description"),
				Cron:        lo.ToPtr("0 0 * * * *"),
				Platform:    enums.JobPlatformTypeGo,
				DownloadURL: testScriptURL,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateJobTemplateInput{
				Title:       "Test Job Template",
				Platform:    enums.JobPlatformTypeGo,
				DownloadURL: testScriptURL,
				OwnerID:     lo.ToPtr(testUser1.OrganizationID),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateJobTemplateInput{
				Title:       "Test Job Template",
				Platform:    enums.JobPlatformTypeGo,
				DownloadURL: testScriptURL,
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateJobTemplateInput{
				Title:       "Test Job Template",
				Platform:    enums.JobPlatformTypeGo,
				DownloadURL: testScriptURL,
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "missing required field, title",
			request: testclient.CreateJobTemplateInput{
				Platform:    enums.JobPlatformTypeGo,
				DownloadURL: testScriptURL,
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "missing required field, platform",
			request: testclient.CreateJobTemplateInput{
				Title:       "Test Job Template",
				DownloadURL: testScriptURL,
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not a valid JobTemplateJobPlatformType",
		},
		{
			name: "missing required field, download url",
			request: testclient.CreateJobTemplateInput{
				Title:    "Test Job Template",
				Platform: enums.JobPlatformTypeGo,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "invalid cron",
			request: testclient.CreateJobTemplateInput{
				Title:       "Test Job Template",
				Platform:    enums.JobPlatformTypeGo,
				DownloadURL: testScriptURL,
				Cron:        lo.ToPtr("0 0 * * * a"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "invalid cron syntax",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateJobTemplate(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// check required fields
			assert.Check(t, is.Equal(tc.request.Title, resp.CreateJobTemplate.JobTemplate.Title))
			assert.Check(t, is.Equal(tc.request.Platform, resp.CreateJobTemplate.JobTemplate.Platform))
			assert.Check(t, is.Equal(tc.request.DownloadURL, resp.CreateJobTemplate.JobTemplate.DownloadURL))

			// check optional fields with if checks if they were provided or not
			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.CreateJobTemplate.JobTemplate.Description))
			}

			if tc.request.Cron != nil {
				assert.Check(t, is.Equal(*tc.request.Cron, *resp.CreateJobTemplate.JobTemplate.Cron))
			}

			// cleanup each JobTemplate created
			(&Cleanup[*generated.JobTemplateDeleteOne]{client: suite.client.db.JobTemplate, ID: resp.CreateJobTemplate.JobTemplate.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}
}

func TestMutationUpdateJobTemplate(t *testing.T) {
	jobTemplate := (&JobTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.UpdateJobTemplateInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: testclient.UpdateJobTemplateInput{
				Description: lo.ToPtr("Test Description Updated"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: testclient.UpdateJobTemplateInput{
				DownloadURL: lo.ToPtr(testScriptURL),
				Cron:        lo.ToPtr("0 0 * * * *"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, not enough permissions",
			request: testclient.UpdateJobTemplateInput{
				Description: lo.ToPtr("Test Description Updated not allowed"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateJobTemplateInput{
				Description: lo.ToPtr("Test Description Updated not allowed"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateJobTemplate(tc.ctx, jobTemplate.ID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.UpdateJobTemplate.JobTemplate.Description))
			}

			if tc.request.DownloadURL != nil {
				assert.Check(t, is.Equal(*tc.request.DownloadURL, resp.UpdateJobTemplate.JobTemplate.DownloadURL))
			}

			if tc.request.Cron != nil {
				assert.Check(t, is.Equal(*tc.request.Cron, *resp.UpdateJobTemplate.JobTemplate.Cron))
			}
		})
	}

	(&Cleanup[*generated.JobTemplateDeleteOne]{client: suite.client.db.JobTemplate, ID: jobTemplate.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteJobTemplate(t *testing.T) {
	// create JobTemplates to be deleted
	jobTemplate1 := (&JobTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	jobTemplate2 := (&JobTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	jobTemplate3 := (&JobTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	jobTemplateSystem := (&JobTemplateBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not found, delete",
			idToDelete:  jobTemplate1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "not authorized, delete",
			idToDelete:  jobTemplate1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "not authorized, delete system owned JobTemplate",
			idToDelete:  jobTemplateSystem.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: jobTemplate1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:       "happy path, delete system owned JobTemplate",
			idToDelete: jobTemplateSystem.ID,
			client:     suite.client.api,
			ctx:        systemAdminUser.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  jobTemplate1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: jobTemplate2.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete using api token",
			idToDelete: jobTemplate3.ID,
			client:     suite.client.apiWithToken,
			ctx:        context.Background(),
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteJobTemplate(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteJobTemplate.DeletedID))
		})
	}
}
