package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/iam/auth"
)

func TestQueryFile(t *testing.T) {
	// create an Evidence to be queried using testUser1
	uf, err := objects.NewUploadFile("testdata/uploads/orgs.csv")
	assert.NilError(t, err)

	uploadFile := &graphql.Upload{
		File:        uf.File,
		Filename:    uf.Filename,
		Size:        uf.Size,
		ContentType: uf.ContentType,
	}

	// create control to be used in the Evidence
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	expectUpload(t, suite.client.objectStore.Storage, []graphql.Upload{*uploadFile})

	evidence, err := suite.client.api.CreateEvidence(testUser1.UserCtx, testclient.CreateEvidenceInput{
		Name:       "Test Evidence",
		ControlIDs: []string{control.ID},
	}, []*graphql.Upload{uploadFile})
	assert.NilError(t, err)

	getEvidence, err := suite.client.api.GetEvidenceByID(testUser1.UserCtx, evidence.CreateEvidence.Evidence.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Len(getEvidence.Evidence.Files.Edges, 1))

	evidenceFile := getEvidence.Evidence.Files.Edges[0].Node

	avatarFile, err := objects.NewUploadFile("testdata/uploads/logo.png")
	assert.NilError(t, err)

	uploadFile = &graphql.Upload{
		File:        avatarFile.File,
		Filename:    avatarFile.Filename,
		Size:        avatarFile.Size,
		ContentType: avatarFile.ContentType,
	}

	expectUpload(t, suite.client.objectStore.Storage, []graphql.Upload{*uploadFile})
	// update user avatar to the file
	userResp, err := suite.client.api.UpdateUser(testUser1.UserCtx, testUser1.ID, testclient.UpdateUserInput{}, uploadFile)
	assert.NilError(t, err)
	assert.Check(t, userResp.UpdateUser.User.AvatarFile != nil)

	userFileID := *userResp.UpdateUser.User.AvatarLocalFileID

	// user in another org context
	adminUserCtxAnotherOrg := auth.NewTestContextWithOrgID(adminUser.ID, adminUser.PersonalOrgID)

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
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, avatar file",
			queryID: userFileID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, admin user",
			queryID: evidenceFile.ID,
			client:  suite.client.api,
			ctx:     adminUser.UserCtx,
		},
		{
			name:    "avatar file needs to be found to display to other users",
			queryID: userFileID,
			client:  suite.client.api,
			ctx:     adminUser.UserCtx,
		},
		{
			name:    "happy path, user authorized via the control to view the file",
			queryID: evidenceFile.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
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
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "File not found, using not authorized user",
			queryID:  evidenceFile.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "File not found, using not authorized user to avatar file",
			queryID:  userFileID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetFileByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.File.ID))
		})
	}

	(&Cleanup[*generated.FileDeleteOne]{client: suite.client.db.File, IDs: []string{evidenceFile.ID, userFileID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.EvidenceDeleteOne]{client: suite.client.db.Evidence, ID: evidence.CreateEvidence.Evidence.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryFiles(t *testing.T) {
	t.Parallel()

	// create users so we dont have conflicts with other tests
	testUser := suite.userBuilder(context.Background(), t)
	patClient := suite.setupPatClient(testUser, t)
	tokenClient := suite.setupAPITokenClient(testUser.UserCtx, t)

	orgMember := (&OrgMemberBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	orgMemberCtx := auth.NewTestContextWithOrgID(orgMember.UserID, orgMember.OrganizationID)

	// create an Evidence to be queried using testUser
	uf, err := objects.NewUploadFile("testdata/uploads/orgs.csv")
	assert.NilError(t, err)

	uploadFile := &graphql.Upload{
		File:        uf.File,
		Filename:    uf.Filename,
		Size:        uf.Size,
		ContentType: uf.ContentType,
	}

	// create control to be used in the Evidence
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	expectUpload(t, suite.client.objectStore.Storage, []graphql.Upload{*uploadFile})

	evidence, err := suite.client.api.CreateEvidence(testUser.UserCtx, testclient.CreateEvidenceInput{
		Name:       "Test Evidence",
		ControlIDs: []string{control.ID},
	}, []*graphql.Upload{uploadFile})
	assert.NilError(t, err)

	getEvidence, err := suite.client.api.GetEvidenceByID(testUser.UserCtx, evidence.CreateEvidence.Evidence.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Len(getEvidence.Evidence.Files.Edges, 1))

	evidenceFile := getEvidence.Evidence.Files.Edges[0].Node

	avatarFile, err := objects.NewUploadFile("testdata/uploads/logo.png")
	assert.NilError(t, err)

	uploadFile = &graphql.Upload{
		File:        avatarFile.File,
		Filename:    avatarFile.Filename,
		Size:        avatarFile.Size,
		ContentType: avatarFile.ContentType,
	}

	expectUpload(t, suite.client.objectStore.Storage, []graphql.Upload{*uploadFile})
	// update user avatar to the file
	userResp, err := suite.client.api.UpdateUser(testUser.UserCtx, testUser.ID, testclient.UpdateUserInput{}, uploadFile)
	assert.NilError(t, err)
	assert.Check(t, userResp.UpdateUser.User.AvatarFile != nil)

	userFileID := *userResp.UpdateUser.User.AvatarLocalFileID

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             testUser.UserCtx,
			expectedResults: 2, // 1 for evidence file, 1 for user avatar file
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.client.api,
			ctx:             orgMemberCtx,
			expectedResults: 2, // 1 for evidence file, 1 for user avatar file
		},
		{
			name:            "happy path, using api token",
			client:          tokenClient,
			ctx:             context.Background(),
			expectedResults: 1, // 1 for evidence file, service not able to access another user's avatar file
		},
		{
			name:            "happy path, using pat",
			client:          patClient,
			ctx:             context.Background(),
			expectedResults: 2, // 1 for evidence file, 1 for user avatar file since its the same user's personal access token
		},
		{
			name:            "another user, no Files should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
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

	(&Cleanup[*generated.FileDeleteOne]{client: suite.client.db.File, IDs: []string{evidenceFile.ID, userFileID}}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.EvidenceDeleteOne]{client: suite.client.db.Evidence, ID: evidence.CreateEvidence.Evidence.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUser.UserCtx, t)
}
