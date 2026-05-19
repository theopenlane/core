package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/iam/auth"
)

func TestQueryFile(t *testing.T) {
	// create an Evidence to be queried using testUser1
	fileUpload := uploadFile(t, "testdata/uploads/orgs.csv")

	// create control to be used in the Evidence
	control := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)

	expectUpload(t, suite.client.mockProvider, []graphql.Upload{*fileUpload})

	evidence, err := suite.client.api.CreateEvidence(sharedTestUser1.UserCtx, testclient.CreateEvidenceInput{
		Name:       "Test Evidence",
		ControlIDs: []string{control.ID},
	}, []*graphql.Upload{fileUpload})
	assert.NilError(t, err)

	getEvidence, err := suite.client.api.GetEvidenceByID(sharedTestUser1.UserCtx, evidence.CreateEvidence.Evidence.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Len(getEvidence.Evidence.Files.Edges, 1))

	evidenceFile := getEvidence.Evidence.Files.Edges[0].Node

	fileUpload = uploadFile(t, logoFilePath)

	expectUpload(t, suite.client.mockProvider, []graphql.Upload{*fileUpload})
	// update user avatar to the file
	userResp, err := suite.client.api.UpdateUser(sharedTestUser1.UserCtx, sharedTestUser1.ID, testclient.UpdateUserInput{}, fileUpload, nil)
	assert.NilError(t, err)
	assert.Check(t, userResp.UpdateUser.User.AvatarFile != nil)

	userFileID := *userResp.UpdateUser.User.AvatarLocalFileID

	// user in another org context
	adminUserCtxAnotherOrg := auth.NewTestContextWithOrgID(sharedAdminUser.ID, sharedAdminUser.PersonalOrgID)

	// add test cases for querying the File
	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: evidenceFile.ID,
			client:  suite.client.api,
			ctx:     sharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, avatar file",
			queryID: userFileID,
			client:  suite.client.api,
			ctx:     sharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, admin user",
			queryID: evidenceFile.ID,
			client:  suite.client.api,
			ctx:     sharedAdminUser.UserCtx,
		},
		{
			name:    "avatar file needs to be found to display to other users",
			queryID: userFileID,
			client:  suite.client.api,
			ctx:     sharedAdminUser.UserCtx,
		},
		{
			name:    "happy path, user authorized via the control to view the file",
			queryID: evidenceFile.ID,
			client:  suite.client.api,
			ctx:     sharedViewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: evidenceFile.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "admin user accessing a different org, shouldn't be able to access the evidence file from the other org",
			queryID:  evidenceFile.ID,
			client:   suite.client.api,
			ctx:      adminUserCtxAnotherOrg,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "File not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      sharedTestUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "File not found, using not authorized user",
			queryID:  evidenceFile.ID,
			client:   suite.client.api,
			ctx:      sharedTestUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "File not found, using not authorized user to avatar file",
			queryID:  userFileID,
			client:   suite.client.api,
			ctx:      sharedTestUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetFileByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.File.ID))
			assert.Check(t, resp.File.StoragePath != nil)
			assert.Check(t, resp.File.StorageProvider != nil)
			assert.Check(t, resp.File.StorageRegion != nil)
		})
	}

	(&Cleanup[*generated.FileDeleteOne]{client: suite.client.db.File, IDs: []string{evidenceFile.ID, userFileID}}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.EvidenceDeleteOne]{client: suite.client.db.Evidence, ID: evidence.CreateEvidence.Evidence.ID}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(sharedTestUser1.UserCtx, t)
}

func TestQueryFiles(t *testing.T) {
	// create users so we dont have conflicts with other tests
	localTestUser := suite.seedFreshMinimalOrgUsers(t, true)

	// create users so we dont have conflicts with other tests
	localTestUser2 := suite.seedOrgOwner(t)

	anotherTestUser := suite.userBuilder(context.Background(), t)

	// create an evidence to be queried using testUser
	fileUpload := uploadFile(t, "testdata/uploads/orgs.csv")

	// create control to be used in the Evidence
	control := (&ControlBuilder{client: suite.client}).MustNew(localTestUser.owner.UserCtx, t)

	expectUpload(t, suite.client.mockProvider, []graphql.Upload{*fileUpload})

	evidence, err := suite.client.api.CreateEvidence(localTestUser.owner.UserCtx, testclient.CreateEvidenceInput{
		Name:       "Test Evidence",
		ControlIDs: []string{control.ID},
	}, []*graphql.Upload{fileUpload})
	assert.NilError(t, err)

	getEvidence, err := suite.client.api.GetEvidenceByID(localTestUser.owner.UserCtx, evidence.CreateEvidence.Evidence.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Len(getEvidence.Evidence.Files.Edges, 1))

	fileUpload = uploadFile(t, logoFilePath)

	expectUpload(t, suite.client.mockProvider, []graphql.Upload{*fileUpload})
	// update user avatar to the file
	userResp, err := suite.client.api.UpdateUser(localTestUser.owner.UserCtx, localTestUser.owner.ID, testclient.UpdateUserInput{}, fileUpload, nil)
	assert.NilError(t, err)
	assert.Check(t, userResp.UpdateUser.User.AvatarFile != nil)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             localTestUser.owner.UserCtx,
			expectedResults: 2, // 1 for evidence file, 1 for user avatar file
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.client.api,
			ctx:             localTestUser.member.UserCtx,
			expectedResults: 2, // 1 for evidence file, 1 for user avatar file
		},
		{
			name:            "happy path, using api token",
			client:          localTestUser.apiClient,
			ctx:             context.Background(),
			expectedResults: 1, // 1 for evidence file, as no avatar
		},
		{
			name:            "another org api token cannot access orgs files",
			client:          localTestUser2.apiClient,
			ctx:             context.Background(),
			expectedResults: 0,
		},
		{
			name:            "happy path, using pat",
			client:          localTestUser.adminPatClient,
			ctx:             context.Background(),
			expectedResults: 2, // 1 for evidence file, 1 for user avatar file since its the same user's personal access token
		},
		{
			name:            "another user, no Files should be returned",
			client:          suite.client.api,
			ctx:             anotherTestUser.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllFiles(tc.ctx)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Files.Edges, tc.expectedResults))
		})
	}

	cleanupOrganizationDataWithContext(localTestUser.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(localTestUser2.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(anotherTestUser.UserCtx, t)
}
