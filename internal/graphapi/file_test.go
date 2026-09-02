package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/iam/auth"
)

func TestQueryFile(t *testing.T) {
	// create an Evidence to be queried using th.SharedTestUser1
	fileUpload := th.UploadFile(t, "testdata/uploads/orgs.csv")

	// create control to be used in the Evidence
	control := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*fileUpload})

	evidence, err := suite.Client.API.CreateEvidence(th.SharedTestUser1.UserCtx, testclient.CreateEvidenceInput{
		Name:       "Test Evidence",
		ControlIDs: []string{control.ID},
	}, []*graphql.Upload{fileUpload})
	assert.NilError(t, err)

	getEvidence, err := suite.Client.API.GetEvidenceByID(th.SharedTestUser1.UserCtx, evidence.CreateEvidence.Evidence.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Len(getEvidence.Evidence.Files.Edges, 1))

	evidenceFile := getEvidence.Evidence.Files.Edges[0].Node

	fileUpload = th.UploadFile(t, th.LogoFilePath)

	th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*fileUpload})
	// update user avatar to the file
	userResp, err := suite.Client.API.UpdateUser(th.SharedTestUser1.UserCtx, th.SharedTestUser1.ID, testclient.UpdateUserInput{}, fileUpload, nil)
	assert.NilError(t, err)
	assert.Check(t, userResp.UpdateUser.User.AvatarFile != nil)

	userFileID := *userResp.UpdateUser.User.AvatarLocalFileID

	// user in another org context
	adminUserCtxAnotherOrg := auth.NewTestContextWithOrgID(th.SharedAdminUser.ID, th.SharedAdminUser.PersonalOrgID)

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
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, avatar file",
			queryID: userFileID,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, admin user",
			queryID: evidenceFile.ID,
			client:  suite.Client.API,
			ctx:     th.SharedAdminUser.UserCtx,
		},
		{
			name:    "avatar file needs to be found to display to other users",
			queryID: userFileID,
			client:  suite.Client.API,
			ctx:     th.SharedAdminUser.UserCtx,
		},
		{
			name:    "happy path, user authorized via the control to view the file",
			queryID: evidenceFile.ID,
			client:  suite.Client.API,
			ctx:     th.SharedViewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: evidenceFile.ID,
			client:  suite.Client.APIWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "admin user accessing a different org, shouldn't be able to access the evidence file from the other org",
			queryID:  evidenceFile.ID,
			client:   suite.Client.API,
			ctx:      adminUserCtxAnotherOrg,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "File not found, invalid ID",
			queryID:  "invalid",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "File not found, using not authorized user",
			queryID:  evidenceFile.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "File not found, using not authorized user to avatar file",
			queryID:  userFileID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
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

	(&th.Cleanup[*generated.FileDeleteOne]{Client: suite.Client.DB.File, IDs: []string{evidenceFile.ID, userFileID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.EvidenceDeleteOne]{Client: suite.Client.DB.Evidence, ID: evidence.CreateEvidence.Evidence.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, ID: control.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryFiles(t *testing.T) {
	// create users so we dont have conflicts with other tests
	localTestUser := suite.SeedFreshMinimalOrgUsers(t, true)

	// create users so we dont have conflicts with other tests
	localTestUser2 := suite.SeedOrgOwner(t)

	anotherTestUser := suite.UserBuilder(context.Background(), t)

	// create an evidence to be queried using testUser
	fileUpload := th.UploadFile(t, "testdata/uploads/orgs.csv")

	// create control to be used in the Evidence
	control := (&th.ControlBuilder{Client: suite.Client}).MustNew(localTestUser.Owner.UserCtx, t)

	th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*fileUpload})

	evidence, err := suite.Client.API.CreateEvidence(localTestUser.Owner.UserCtx, testclient.CreateEvidenceInput{
		Name:       "Test Evidence",
		ControlIDs: []string{control.ID},
	}, []*graphql.Upload{fileUpload})
	assert.NilError(t, err)

	getEvidence, err := suite.Client.API.GetEvidenceByID(localTestUser.Owner.UserCtx, evidence.CreateEvidence.Evidence.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Len(getEvidence.Evidence.Files.Edges, 1))

	fileUpload = th.UploadFile(t, th.LogoFilePath)

	th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*fileUpload})
	// update user avatar to the file
	userResp, err := suite.Client.API.UpdateUser(localTestUser.Owner.UserCtx, localTestUser.Owner.ID, testclient.UpdateUserInput{}, fileUpload, nil)
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
			client:          suite.Client.API,
			ctx:             localTestUser.Owner.UserCtx,
			expectedResults: 2, // 1 for evidence file, 1 for user avatar file
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.Client.API,
			ctx:             localTestUser.Member.UserCtx,
			expectedResults: 2, // 1 for evidence file, 1 for user avatar file
		},
		{
			name:            "happy path, using api token",
			client:          localTestUser.APIClient,
			ctx:             context.Background(),
			expectedResults: 1, // 1 for evidence file, as no avatar
		},
		{
			name:            "another org api token cannot access orgs files",
			client:          localTestUser2.APIClient,
			ctx:             context.Background(),
			expectedResults: 0,
		},
		{
			name:            "happy path, using pat",
			client:          localTestUser.AdminPatClient,
			ctx:             context.Background(),
			expectedResults: 2, // 1 for evidence file, 1 for user avatar file since its the same user's personal access token
		},
		{
			name:            "another user, no Files should be returned",
			client:          suite.Client.API,
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

	th.CleanupOrganizationDataWithContext(localTestUser.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(localTestUser2.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(anotherTestUser.UserCtx, t)
}
