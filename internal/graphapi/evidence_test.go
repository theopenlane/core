package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/ulids"
)

func (suite *GraphTestSuite) TestQueryEvidence() {
	t := suite.T()

	program := (&ProgramBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)

	(&ProgramMemberBuilder{client: suite.client, UserID: viewOnlyUser.ID, ProgramID: program.ID}).MustNew(adminUser.UserCtx, t)

	control := (&ControlBuilder{client: suite.client, ProgramID: program.ID}).MustNew(adminUser.UserCtx, t)

	// create an Evidence to be queried using adminUser
	// org owner (testUser1) should automatically have access to the Evidence
	evidence := (&EvidenceBuilder{client: suite.client, ProgramID: program.ID}).MustNew(adminUser.UserCtx, t)

	evidenceControl := (&EvidenceBuilder{client: suite.client, ControlID: control.ID}).MustNew(testUser1.UserCtx, t)

	// add test cases for querying the Evidence
	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path, creator of the evidence",
			queryID: evidence.ID,
			client:  suite.client.api,
			ctx:     adminUser.UserCtx,
		},
		{
			name:    "happy path, permissions via control",
			queryID: evidenceControl.ID,
			client:  suite.client.api,
			ctx:     adminUser.UserCtx,
		},
		{
			name:    "happy path, org owner",
			queryID: evidence.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:     "read only user in organization, authorized via program",
			queryID:  evidence.ID,
			client:   suite.client.api,
			ctx:      viewOnlyUser.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "read only user in organization, not authorized",
			queryID:  evidenceControl.ID,
			client:   suite.client.api,
			ctx:      viewOnlyUser.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:    "happy path using personal access token",
			queryID: evidence.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "Evidence not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "Evidence not found, using not authorized user",
			queryID:  evidence.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetEvidenceByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			require.NotEmpty(t, resp.Evidence)

			assert.Equal(t, tc.queryID, resp.Evidence.ID)

			assert.NotEmpty(t, resp.Evidence.Name)
			assert.NotEmpty(t, resp.Evidence.DisplayID)
			assert.NotEmpty(t, resp.Evidence.CreatedAt)
			assert.NotEmpty(t, resp.Evidence.UpdatedAt)
		})
	}
}

func (suite *GraphTestSuite) TestQueryEvidences() {
	t := suite.T()

	// create multiple objects by adminUser, org owner should have access to them as well
	(&EvidenceBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
	(&EvidenceBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using read only user of the same org, access not automatically granted",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 0,
		},
		{
			name:            "happy path, using api token, access not automatically granted",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: 0,
		},
		{
			name:            "happy path, using pat, which is for the org owner so access is granted",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no Evidences should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllEvidences(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.Evidences.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateEvidence() {
	t := suite.T()

	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	task := (&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	pngFile, err := objects.NewUploadFile("testdata/uploads/logo.png")
	require.NoError(t, err)

	csvFile, err := objects.NewUploadFile("testdata/uploads/orgs.csv")
	require.NoError(t, err)

	pdfFile, err := objects.NewUploadFile("testdata/uploads/hello.pdf")
	require.NoError(t, err)

	txtFile, err := objects.NewUploadFile("testdata/uploads/hello.txt")
	require.NoError(t, err)

	testCases := []struct {
		name        string
		request     openlaneclient.CreateEvidenceInput
		files       []*graphql.Upload
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateEvidenceInput{
				Name: "Test Evidence",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateEvidenceInput{
				Name:                "Test Evidence",
				Description:         lo.ToPtr("This is a test Evidence"),
				CollectionProcedure: lo.ToPtr("This is how we collected the Evidence"),
				Source:              lo.ToPtr("meows"),
				CreationDate:        lo.ToPtr(time.Now().Add(-time.Hour)),
				RenewalDate:         lo.ToPtr(time.Now().Add(365 * 24 * time.Hour)),
				IsAutomated:         lo.ToPtr(true),
				URL:                 lo.ToPtr("https://example.com/my-evidence.png"),
				ProgramIDs:          []string{program.ID},
				TaskIDs:             []string{task.ID},
			},
			files: []*graphql.Upload{
				{
					File:        pngFile.File,
					Filename:    pngFile.Filename,
					Size:        pngFile.Size,
					ContentType: pngFile.ContentType,
				},
				{
					File:        csvFile.File,
					Filename:    csvFile.Filename,
					Size:        csvFile.Size,
					ContentType: csvFile.ContentType,
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, admin user in org",
			request: openlaneclient.CreateEvidenceInput{
				Name:                "Test Evidence",
				Description:         lo.ToPtr("This is a test Evidence"),
				CollectionProcedure: lo.ToPtr("This is how we collected the Evidence"),
				Source:              lo.ToPtr("meows"),
				CreationDate:        lo.ToPtr(time.Now().Add(-time.Hour)),
				RenewalDate:         lo.ToPtr(time.Now().Add(365 * 24 * time.Hour)),
				IsAutomated:         lo.ToPtr(true),
				URL:                 lo.ToPtr("https://example.com/my-evidence.png"),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: openlaneclient.CreateEvidenceInput{
				Name:    "Test Evidence - TSK-123",
				TaskIDs: []string{task.ID},
				OwnerID: testUser1.OrganizationID,
			},
			files: []*graphql.Upload{
				{
					File:        pdfFile.File,
					Filename:    pdfFile.Filename,
					Size:        pdfFile.Size,
					ContentType: pdfFile.ContentType,
				},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: openlaneclient.CreateEvidenceInput{
				Name: "Test Evidence - TSK-123",
			},
			files: []*graphql.Upload{
				{
					File:        txtFile.File,
					Filename:    txtFile.Filename,
					Size:        txtFile.Size,
					ContentType: txtFile.ContentType,
				},
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: openlaneclient.CreateEvidenceInput{
				Name: "Test Evidence",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "missing required field",
			request: openlaneclient.CreateEvidenceInput{
				Description: lo.ToPtr("This is a test Evidence"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.files != nil && len(tc.files) > 0 {
				expectUploadNillable(t, suite.client.objectStore.Storage, tc.files)
			}

			resp, err := tc.client.CreateEvidence(tc.ctx, tc.request, tc.files)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// check required fields
			require.NotEmpty(t, resp.CreateEvidence.Evidence.ID)
			require.NotEmpty(t, resp.CreateEvidence.Evidence.DisplayID)
			require.NotEmpty(t, resp.CreateEvidence.Evidence.Name)

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.CreateEvidence.Evidence.Description)
			} else {
				assert.Empty(t, resp.CreateEvidence.Evidence.Description)
			}

			if tc.request.CollectionProcedure != nil {
				assert.Equal(t, *tc.request.CollectionProcedure, *resp.CreateEvidence.Evidence.CollectionProcedure)
			} else {
				assert.Empty(t, resp.CreateEvidence.Evidence.CollectionProcedure)
			}

			if tc.request.Source != nil {
				assert.Equal(t, *tc.request.Source, *resp.CreateEvidence.Evidence.Source)
			} else {
				assert.Empty(t, resp.CreateEvidence.Evidence.Source)
			}

			if tc.request.CreationDate != nil {
				assert.WithinDuration(t, *tc.request.CreationDate, resp.CreateEvidence.Evidence.CreationDate, time.Minute)
			} else {
				assert.WithinDuration(t, time.Now(), resp.CreateEvidence.Evidence.CreationDate, time.Minute)
			}

			if tc.request.RenewalDate != nil {
				assert.WithinDuration(t, *tc.request.RenewalDate, *resp.CreateEvidence.Evidence.RenewalDate, time.Minute)
			} else {
				assert.WithinDuration(t, time.Now().Add(365*24*time.Hour), *resp.CreateEvidence.Evidence.RenewalDate, time.Minute)
			}

			if tc.request.IsAutomated != nil {
				assert.Equal(t, *tc.request.IsAutomated, *resp.CreateEvidence.Evidence.IsAutomated)
			} else {
				assert.False(t, *resp.CreateEvidence.Evidence.IsAutomated)
			}

			if tc.request.URL != nil {
				assert.Equal(t, *tc.request.URL, *resp.CreateEvidence.Evidence.URL)
			} else {
				assert.Empty(t, resp.CreateEvidence.Evidence.URL)
			}

			if tc.request.ProgramIDs != nil {
				assert.Len(t, resp.CreateEvidence.Evidence.Programs, len(tc.request.ProgramIDs))
			} else {
				assert.Empty(t, resp.CreateEvidence.Evidence.Programs)
			}

			if tc.request.TaskIDs != nil {
				assert.Len(t, resp.CreateEvidence.Evidence.Tasks, len(tc.request.TaskIDs))
			} else {
				assert.Empty(t, resp.CreateEvidence.Evidence.Tasks)
			}

			if tc.files != nil && len(tc.files) > 0 {
				assert.Len(t, resp.CreateEvidence.Evidence.Files, len(tc.files))
			} else {
				assert.Empty(t, resp.CreateEvidence.Evidence.Files)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateEvidence() {
	t := suite.T()

	evidence := (&EvidenceBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)

	pdfFile, err := objects.NewUploadFile("testdata/uploads/hello.pdf")
	require.NoError(t, err)

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateEvidenceInput
		files       []*graphql.Upload
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: openlaneclient.UpdateEvidenceInput{
				CollectionProcedure: lo.ToPtr("This is how we collected the updated Evidence"),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: openlaneclient.UpdateEvidenceInput{
				Name:                lo.ToPtr("Updated Evidence"),
				Description:         lo.ToPtr("This is an updated Evidence"),
				CollectionProcedure: lo.ToPtr("This is how we collected the updated Evidence"),
				Source:              lo.ToPtr("meows"),
			},
			files: []*graphql.Upload{
				{
					File:        pdfFile.File,
					Filename:    pdfFile.Filename,
					Size:        pdfFile.Size,
					ContentType: pdfFile.ContentType,
				},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, no permissions",
			request: openlaneclient.UpdateEvidenceInput{
				Description: lo.ToPtr("This is an updated description of evidence"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: openlaneclient.UpdateEvidenceInput{
				Source: lo.ToPtr("woofs"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			if tc.files != nil && len(tc.files) > 0 {
				expectUploadNillable(t, suite.client.objectStore.Storage, tc.files)
			}

			resp, err := tc.client.UpdateEvidence(tc.ctx, evidence.ID, tc.request, tc.files)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// add checks for the updated fields if they were set in the request
			if tc.request.Name != nil {
				assert.Equal(t, *tc.request.Name, resp.UpdateEvidence.Evidence.Name)
			}

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.UpdateEvidence.Evidence.Description)
			}

			if tc.request.CollectionProcedure != nil {
				assert.Equal(t, *tc.request.CollectionProcedure, *resp.UpdateEvidence.Evidence.CollectionProcedure)
			}

			if tc.request.Source != nil {
				assert.Equal(t, *tc.request.Source, *resp.UpdateEvidence.Evidence.Source)
			}

			if tc.files != nil && len(tc.files) > 0 {
				assert.Len(t, resp.UpdateEvidence.Evidence.Files, len(tc.files))
			}

		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteEvidence() {
	t := suite.T()

	// create objects to be deleted
	evidence1 := (&EvidenceBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
	evidence2 := (&EvidenceBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  evidence1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: evidence1.ID,
			client:     suite.client.api,
			ctx:        adminUser.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  evidence1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: evidence2.ID,
			client:     suite.client.apiWithPAT,
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
			resp, err := tc.client.DeleteEvidence(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteEvidence.DeletedID)
		})
	}
}
