package graphapi_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestQueryProcedure(t *testing.T) {
	// create an Procedure to be queried using testUser1
	procedure := (&ProcedureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	anonymousContext := createAnonymousTrustCenterContext(ulids.New().String(), testUser1.OrganizationID)

	// add test cases for querying the procedure
	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: procedure.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: procedure.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: procedure.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "procedure not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "procedure not found, using not authorized user",
			queryID:  procedure.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "no access, anonymous user",
			client:   suite.client.api,
			ctx:      anonymousContext,
			queryID:  procedure.ID,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetProcedureByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {

				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.Procedure.ID))
			assert.Check(t, len(resp.Procedure.Name) != 0)
		})
	}

	// cleanup
	(&Cleanup[*generated.ProcedureDeleteOne]{client: suite.client.db.Procedure, ID: procedure.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryProcedures(t *testing.T) {
	// create multiple Procedures to be queried using testUser1
	p1 := (&ProcedureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	p2 := (&ProcedureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add procedure for another org; it should not be returned in the list
	p3 := (&ProcedureBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

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
			expectedResults: 2,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using api token",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "happy path, using pat",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no procedures should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllProcedures(tc.ctx)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Procedures.Edges, tc.expectedResults), "expected %d, got %d", tc.expectedResults, len(resp.Procedures.Edges))
		})
	}

	// cleanup procedures created for the test
	(&Cleanup[*generated.ProcedureDeleteOne]{client: suite.client.db.Procedure, IDs: []string{p1.ID, p2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ProcedureDeleteOne]{client: suite.client.db.Procedure, ID: p3.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationCreateProcedure(t *testing.T) {
	anotherGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// group for the view only user
	groupMember := (&GroupMemberBuilder{client: suite.client, UserID: viewOnlyUser.ID}).MustNew(testUser1.UserCtx, t)

	approverGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	delegateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name          string
		request       testclient.CreateProcedureInput
		addGroupToOrg bool
		client        *testclient.TestClient
		ctx           context.Context
		expectedErr   string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateProcedureInput{
				Name: "Test Procedure",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input except edges, you should have to have view access to a group to add it as an edge on this object",
			request: testclient.CreateProcedureInput{
				Name:          "Releasing a new version",
				Details:       lo.ToPtr("instructions on how to release a new version"),
				Status:        &enums.DocumentDraft,
				ProcedureType: lo.ToPtr("sop"),
				Revision:      lo.ToPtr("v1.0.0"),
				ApproverID:    &approverGroup.ID,
				DelegateID:    &delegateGroup.ID,
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "add editor group",
			request: testclient.CreateProcedureInput{
				Name:            "Test Procedure",
				EditorIDs:       []string{testUser1.GroupID},
				BlockedGroupIDs: []string{anotherGroup.ID},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "add editor group, again - ensures the same group can be added to multiple procedures",
			request: testclient.CreateProcedureInput{
				Name:            "Test Procedure",
				EditorIDs:       []string{testUser1.GroupID},
				BlockedGroupIDs: []string{anotherGroup.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateProcedureInput{
				Name:    "Test Procedure",
				OwnerID: &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path with details, using pat",
			request: testclient.CreateProcedureInput{
				Name:    "Test Procedure",
				OwnerID: &testUser1.OrganizationID,
				Details: lo.ToPtr(gofakeit.Sentence()),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateProcedureInput{
				Name: "Test Procedure",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateProcedureInput{
				Name: "Test Procedure",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user now authorized, add group to org first",
			request: testclient.CreateProcedureInput{
				Name: "Test Procedure",
			},
			addGroupToOrg: true,
			client:        suite.client.api,
			ctx:           viewOnlyUser.UserCtx,
		},
		{
			name: "missing required field",
			request: testclient.CreateProcedureInput{
				Details: lo.ToPtr("instructions on how to release a new version"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.addGroupToOrg {
				_, err := suite.client.api.UpdateOrganization(testUser1.UserCtx, testUser1.OrganizationID,
					testclient.UpdateOrganizationInput{
						AddProcedureCreatorIDs: []string{groupMember.GroupID},
					}, nil)
				assert.NilError(t, err)
			}

			resp, err := tc.client.CreateProcedure(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// check required fields
			assert.Check(t, is.Equal(tc.request.Name, resp.CreateProcedure.Procedure.Name))

			assert.Check(t, len(resp.CreateProcedure.Procedure.DisplayID) != 0)
			assert.Check(t, is.Contains(resp.CreateProcedure.Procedure.DisplayID, "PRD-"))

			// check optional fields with if checks if they were provided or not
			if tc.request.Status != nil {
				assert.Check(t, is.Equal(*tc.request.Status, *resp.CreateProcedure.Procedure.Status))
			} else {
				// default status is draft
				assert.Check(t, is.Equal(enums.DocumentDraft, *resp.CreateProcedure.Procedure.Status))
			}

			if tc.request.ProcedureType != nil {
				assert.Check(t, is.Equal(*tc.request.ProcedureType, *resp.CreateProcedure.Procedure.ProcedureType))
			} else {
				assert.Check(t, is.Len(*resp.CreateProcedure.Procedure.ProcedureType, 0))
			}

			if tc.request.Revision != nil {
				assert.Check(t, is.Equal(*tc.request.Revision, *resp.CreateProcedure.Procedure.Revision))
			} else {
				// default revision is v0.0.1
				assert.Check(t, is.Equal(models.DefaultRevision, *resp.CreateProcedure.Procedure.Revision))
			}

			if tc.request.Details != nil {
				assert.Check(t, is.DeepEqual(tc.request.Details, resp.CreateProcedure.Procedure.Details))
				assert.Check(t, resp.CreateProcedure.Procedure.Summary != nil)
			} else {
				assert.Check(t, is.Len(*resp.CreateProcedure.Procedure.Details, 0))
				assert.Check(t, is.Len(*resp.CreateProcedure.Procedure.Summary, 0))
			}

			if tc.request.EditorIDs != nil {
				assert.Check(t, is.Len(resp.CreateProcedure.Procedure.Editors.Edges, len(tc.request.EditorIDs)))
			} else {
				assert.Check(t, is.Len(resp.CreateProcedure.Procedure.Editors.Edges, 0))
			}

			if tc.request.BlockedGroupIDs != nil {
				assert.Check(t, is.Len(resp.CreateProcedure.Procedure.BlockedGroups.Edges, len(tc.request.BlockedGroupIDs)))
			} else {
				assert.Check(t, is.Len(resp.CreateProcedure.Procedure.BlockedGroups.Edges, 0))
			}

			if tc.request.ApproverID != nil {
				assert.Check(t, is.Equal(*tc.request.ApproverID, resp.CreateProcedure.Procedure.Approver.ID))
			} else {
				assert.Check(t, is.Nil(resp.CreateProcedure.Procedure.Approver))
			}

			if tc.request.DelegateID != nil {
				assert.Check(t, is.Equal(*tc.request.DelegateID, resp.CreateProcedure.Procedure.Delegate.ID))
			} else {
				assert.Check(t, is.Nil(resp.CreateProcedure.Procedure.Delegate))
			}

			// cleanup
			(&Cleanup[*generated.ProcedureDeleteOne]{client: suite.client.db.Procedure, ID: resp.CreateProcedure.Procedure.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}

	// cleanup group created for the test
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{anotherGroup.ID, groupMember.GroupID, approverGroup.ID, delegateGroup.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateProcedure(t *testing.T) {
	// create procedure to be updated
	procedure := (&ProcedureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// create another admin user and add them to the same organization and group as testUser1
	// this will allow us to test the group editor permissions
	anotherAdminUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser1.UserCtx, t, &anotherAdminUser, enums.RoleAdmin, testUser1.OrganizationID)

	(&GroupMemberBuilder{client: suite.client, UserID: anotherAdminUser.ID, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)

	// create a viewer user and add them to the same organization as testUser1
	// also add them to the same group as testUser1, this should still allow them to edit the procedure
	// despite not not being an organization admin
	anotherViewerUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser1.UserCtx, t, &anotherViewerUser, enums.RoleMember, testUser1.OrganizationID)

	(&GroupMemberBuilder{client: suite.client, UserID: anotherViewerUser.ID, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)

	// create one more group that will be used to test the blocked group permissions and add anotherViewerUser to it
	blockGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&GroupMemberBuilder{client: suite.client, UserID: anotherViewerUser.ID, GroupID: blockGroup.ID}).MustNew(testUser1.UserCtx, t)

	approverGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	delegateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	anotherApproverGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	anotherDelegateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	log.Error().Msg("starting test cases")
	log.Error().Str("procedureID", procedure.ID).Str("revision", procedure.Revision).Msg("Procedure created for testing")

	testCases := []struct {
		name        string
		request     testclient.UpdateProcedureInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update name field, and add group",
			request: testclient.UpdateProcedureInput{
				Name:         lo.ToPtr("Updated Procedure Name"),
				AddEditorIDs: []string{testUser1.GroupID}, // add the group to the editor groups for subsequent tests
				ApproverID:   &approverGroup.ID,
				DelegateID:   &delegateGroup.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: testclient.UpdateProcedureInput{
				Status:       &enums.DocumentPublished,
				Details:      lo.ToPtr("Updated description"),
				RevisionBump: &models.Minor,
				ApproverID:   &anotherApproverGroup.ID,
				DelegateID:   &anotherDelegateGroup.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "update not allowed, not enough permissions",
			request: testclient.UpdateProcedureInput{
				Name: lo.ToPtr("Updated Procedure Name"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, not enough permissions",
			request: testclient.UpdateProcedureInput{
				Name: lo.ToPtr("Updated Procedure Name Meow"),
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx, // admin users do not automatically inherit permissions
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update allowed, details updated",
			request: testclient.UpdateProcedureInput{
				Details: lo.ToPtr(gofakeit.Sentence()),
			},
			client: suite.client.api,
			ctx:    anotherAdminUser.UserCtx, // user assigned to the group which has editor permissions
		},
		{
			name: "update allowed, user in editor group",
			request: testclient.UpdateProcedureInput{
				Name: lo.ToPtr("Updated Procedure Name Again"),
			},
			client: suite.client.api,
			ctx:    anotherAdminUser.UserCtx, // user assigned to the group which has editor permissions
		},
		{
			name: "member update allowed, user in editor group",
			request: testclient.UpdateProcedureInput{
				Name: lo.ToPtr("Updated Procedure Name Again"),
			},
			client: suite.client.api,
			ctx:    anotherViewerUser.UserCtx, // user assigned to the group which has editor permissions
		},
		{
			name: "happy path, block the group from editing",
			request: testclient.UpdateProcedureInput{
				AddBlockedGroupIDs: []string{blockGroup.ID}, // block the group
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "member update no longer allowed, user in blocked group",
			request: testclient.UpdateProcedureInput{
				Name: lo.ToPtr("Updated Procedure Name Again"),
			},
			client:      suite.client.api,
			ctx:         anotherViewerUser.UserCtx, // user assigned to the group which was blocked
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "happy path, remove the group",
			request: testclient.UpdateProcedureInput{
				RemoveEditorIDs: []string{testUser1.GroupID}, // remove the group from the editor groups
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "update not allowed, editor group was removed",
			request: testclient.UpdateProcedureInput{
				Name: lo.ToPtr("Updated Procedure Name Again Again"),
			},
			client:      suite.client.api,
			ctx:         anotherAdminUser.UserCtx, // user assigned to the group which no longer has editor permissions
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateProcedureInput{
				Details: lo.ToPtr("Updated details"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			tc.ctx = resetContext(tc.ctx, t)

			resp, err := tc.client.UpdateProcedure(tc.ctx, procedure.ID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// check updated fields
			if tc.request.Name != nil {
				assert.Check(t, is.Equal(*tc.request.Name, resp.UpdateProcedure.Procedure.Name))
			}

			if tc.request.Status != nil {
				assert.Check(t, is.Equal(*tc.request.Status, *resp.UpdateProcedure.Procedure.Status))
			}

			if tc.request.ProcedureType != nil {
				assert.Check(t, is.Equal(*tc.request.ProcedureType, *resp.UpdateProcedure.Procedure.ProcedureType))
			}

			if tc.request.Revision != nil {
				assert.Check(t, is.Equal(*tc.request.Revision, *resp.UpdateProcedure.Procedure.Revision))
			}

			if tc.request.RevisionBump == &models.Minor {
				assert.Check(t, is.Equal("v0.1.0", *resp.UpdateProcedure.Procedure.Revision))
			}

			if tc.request.Details != nil {
				assert.Check(t, is.DeepEqual(tc.request.Details, resp.UpdateProcedure.Procedure.Details))
				assert.Check(t, resp.UpdateProcedure.Procedure.Summary != nil)
				assert.Check(t, *resp.UpdateProcedure.Procedure.Summary != procedure.Summary)
			}

			if tc.request.ApproverID != nil {
				assert.Check(t, is.Equal(*tc.request.ApproverID, resp.UpdateProcedure.Procedure.Approver.ID))
			}

			if tc.request.DelegateID != nil {
				assert.Check(t, is.Equal(*tc.request.DelegateID, resp.UpdateProcedure.Procedure.Delegate.ID))
			}
		})
	}

	// cleanup
	(&Cleanup[*generated.ProcedureDeleteOne]{client: suite.client.db.Procedure, ID: procedure.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{blockGroup.ID, approverGroup.ID, delegateGroup.ID, anotherApproverGroup.ID, anotherDelegateGroup.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteProcedure(t *testing.T) {
	// create procedures to be deleted
	procedure1 := (&ProcedureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	procedure2 := (&ProcedureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  procedure1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: procedure1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  procedure1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: procedure2.ID,
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
			resp, err := tc.client.DeleteProcedure(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteProcedure.DeletedID))
		})
	}
}

func TestMutationUpdateBulkProcedure(t *testing.T) {
	newUser := suite.userBuilder(context.Background(), t)
	// create procedures to be updated
	procedure1 := (&ProcedureBuilder{client: suite.client}).MustNew(newUser.UserCtx, t)
	procedure2 := (&ProcedureBuilder{client: suite.client}).MustNew(newUser.UserCtx, t)
	procedure3 := (&ProcedureBuilder{client: suite.client}).MustNew(newUser.UserCtx, t)

	approverGroup := (&GroupBuilder{client: suite.client}).MustNew(newUser.UserCtx, t)
	delegateGroup := (&GroupBuilder{client: suite.client}).MustNew(newUser.UserCtx, t)

	// create another user and add them to the same organization and group as testUser1
	// this will allow us to test the group editor permissions
	anotherAdminUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(newUser.UserCtx, t, &anotherAdminUser, enums.RoleAdmin, newUser.OrganizationID)

	groupMember := (&GroupMemberBuilder{client: suite.client, UserID: anotherAdminUser.ID}).MustNew(newUser.UserCtx, t)

	procedureAnotherUser := (&ProcedureBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	// ensure the user does not currently have access to update the procedure
	res, err := suite.client.api.UpdateBulkProcedure(testUser2.UserCtx, []string{procedure1.ID}, testclient.UpdateProcedureInput{
		Status: lo.ToPtr(enums.DocumentPublished),
	})

	assert.Assert(t, is.Nil(err))
	// make sure nothing was updated
	assert.Equal(t, len(res.UpdateBulkProcedure.Procedures), 0)

	testCases := []struct {
		name                 string
		ids                  []string
		input                testclient.UpdateProcedureInput
		client               *testclient.TestClient
		ctx                  context.Context
		expectedErr          string
		expectedUpdatedCount int
	}{
		{
			name: "happy path, update multiple procedures",
			ids:  []string{procedure1.ID, procedure2.ID, procedure3.ID},
			input: testclient.UpdateProcedureInput{
				Status:       &enums.DocumentPublished,
				ApproverID:   &approverGroup.ID,
				RevisionBump: &models.Minor,
			},
			client:               suite.client.api,
			ctx:                  newUser.UserCtx,
			expectedUpdatedCount: 3,
		},
		{
			name: "happy path, editor permissions",
			ids:  []string{procedure1.ID, procedure2.ID},
			input: testclient.UpdateProcedureInput{
				AddEditorIDs: []string{groupMember.GroupID},
				RevisionBump: &models.Major,
			},
			client:               suite.client.api,
			ctx:                  newUser.UserCtx,
			expectedUpdatedCount: 2,
		},
		{
			name:        "empty ids array",
			ids:         []string{},
			input:       testclient.UpdateProcedureInput{Details: lo.ToPtr("test")},
			client:      suite.client.api,
			ctx:         newUser.UserCtx,
			expectedErr: "ids is required",
		},
		{
			name: "mixed success and failure - some procedures not authorized",
			ids:  []string{procedure1.ID, procedureAnotherUser.ID}, // second should fail authorization
			input: testclient.UpdateProcedureInput{
				Status: &enums.DocumentDraft,
			},
			client:               suite.client.api,
			ctx:                  newUser.UserCtx,
			expectedUpdatedCount: 1, // only procedure1 should be updated
		},
		{
			name: "update not allowed, no permissions to procedures",
			ids:  []string{procedure1.ID},
			input: testclient.UpdateProcedureInput{
				Status: &enums.DocumentPublished,
			},
			client:               suite.client.api,
			ctx:                  testUser2.UserCtx,
			expectedUpdatedCount: 0, // should not find any procedures to update
		},
	}

	for _, tc := range testCases {
		t.Run("Bulk Update "+tc.name, func(t *testing.T) {
			tc.ctx = resetContext(tc.ctx, t)

			resp, err := tc.client.UpdateBulkProcedure(tc.ctx, tc.ids, tc.input)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.UpdateBulkProcedure.Procedures, tc.expectedUpdatedCount))
			assert.Check(t, is.Len(resp.UpdateBulkProcedure.UpdatedIDs, tc.expectedUpdatedCount))

			procedureMap := make(map[string]*testclient.UpdateBulkProcedure_UpdateBulkProcedure_Procedures)
			for _, procedure := range resp.UpdateBulkProcedure.Procedures {
				procedureMap[procedure.ID] = procedure
			}

			for _, expectedID := range tc.ids {
				responseProcedure, found := procedureMap[expectedID]
				if !found {
					continue
				}

				if tc.input.Name != nil {
					assert.Check(t, is.Equal(*tc.input.Name, responseProcedure.Name))
				}

				if tc.input.Status != nil {
					assert.Check(t, is.Equal(*tc.input.Status, *responseProcedure.Status))
				}

				if tc.input.Tags != nil {
					assert.Check(t, is.DeepEqual(tc.input.Tags, responseProcedure.Tags))
				}

				if tc.input.ProcedureType != nil {
					assert.Check(t, is.Equal(*tc.input.ProcedureType, *responseProcedure.ProcedureType))
				}

				if tc.input.Details != nil {
					assert.Check(t, is.DeepEqual(tc.input.Details, responseProcedure.Details))
				}

				if tc.input.ApproverID != nil {
					assert.Check(t, responseProcedure.Approver != nil)
					assert.Check(t, is.Equal(*tc.input.ApproverID, responseProcedure.Approver.ID))
				}

				if tc.input.DelegateID != nil {
					assert.Check(t, responseProcedure.Delegate != nil)
					assert.Check(t, is.Equal(*tc.input.DelegateID, responseProcedure.Delegate.ID))
				}

				if tc.input.RevisionBump == &models.Minor {
					assert.Check(t, is.Equal("v0.1.0", *responseProcedure.Revision))
				}

				if tc.input.RevisionBump == &models.Major {
					assert.Check(t, is.Equal("v1.0.0", *responseProcedure.Revision))
				}

				if len(tc.input.AddEditorIDs) > 0 {
					// ensure the user has access to the procedure now
					res, err := suite.client.api.UpdateProcedure(anotherAdminUser.UserCtx, responseProcedure.ID, testclient.UpdateProcedureInput{
						Tags: []string{"bulk-test-tag"},
					})
					assert.NilError(t, err)
					assert.Check(t, res != nil)
					assert.Check(t, is.Equal(responseProcedure.ID, res.UpdateProcedure.Procedure.ID))
				}
			}

			for _, updatedID := range resp.UpdateBulkProcedure.UpdatedIDs {
				found := false
				for _, expectedID := range tc.ids {
					if expectedID == updatedID {
						found = true
						break
					}
				}
				assert.Check(t, found, "Updated ID %s should be in the original request", updatedID)
			}
		})
	}

	(&Cleanup[*generated.ProcedureDeleteOne]{client: suite.client.db.Procedure, IDs: []string{procedure1.ID, procedure2.ID, procedure3.ID}}).MustDelete(newUser.UserCtx, t)
	(&Cleanup[*generated.ProcedureDeleteOne]{client: suite.client.db.Procedure, ID: procedureAnotherUser.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{approverGroup.ID, delegateGroup.ID, groupMember.GroupID}}).MustDelete(newUser.UserCtx, t)
}
