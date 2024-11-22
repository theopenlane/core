package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQueryProcedure() {
	t := suite.T()

	// create an Procedure to be queried using testUser1
	procedure := (&ProcedureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add test cases for querying the procedure
	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
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
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetProcedureByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			require.NotEmpty(t, resp.Procedure)

			assert.Equal(t, tc.queryID, resp.Procedure.ID)
			assert.NotEmpty(t, resp.Procedure.Name)
		})
	}
}

func (suite *GraphTestSuite) TestQueryProcedures() {
	t := suite.T()

	// create multiple Procedures to be queried using testUser1
	(&ProcedureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&ProcedureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
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
			name:            "another user, no Procedures should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllProcedures(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.Procedures.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateProcedure() {
	t := suite.T()

	anotherGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.CreateProcedureInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateProcedureInput{
				Name: "Test Procedure",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input except edges",
			request: openlaneclient.CreateProcedureInput{
				Name:            "Releasing a new version",
				Description:     lo.ToPtr("instructions on how to release a new version"),
				Status:          lo.ToPtr("draft"),
				ProcedureType:   lo.ToPtr("sop"),
				Version:         lo.ToPtr("v1.0"),
				PurposeAndScope: lo.ToPtr("This procedure is used to release a new version of the software"),
				Background:      lo.ToPtr("The background of the procedure"),
				Satisfies:       lo.ToPtr("The procedure satisfies the requirements"),
				Details: map[string]interface{}{
					"details": "do stuff",
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "add editor group",
			request: openlaneclient.CreateProcedureInput{
				Name:            "Test Procedure",
				EditorIDs:       []string{testUser1.GroupID},
				BlockedGroupIDs: []string{anotherGroup.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "add editor group, again - ensures the same group can be added to multiple procedures",
			request: openlaneclient.CreateProcedureInput{
				Name:            "Test Procedure",
				EditorIDs:       []string{testUser1.GroupID},
				BlockedGroupIDs: []string{anotherGroup.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: openlaneclient.CreateProcedureInput{
				Name:    "Test Procedure",
				OwnerID: &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: openlaneclient.CreateProcedureInput{
				Name: "Test Procedure",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: openlaneclient.CreateProcedureInput{
				Name: "Test Procedure",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "missing required field",
			request: openlaneclient.CreateProcedureInput{
				Description: lo.ToPtr("instructions on how to release a new version"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateProcedure(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// check required fields
			assert.Equal(t, tc.request.Name, resp.CreateProcedure.Procedure.Name)

			// check optional fields with if checks if they were provided or not
			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.CreateProcedure.Procedure.Description)
			} else {
				assert.Empty(t, resp.CreateProcedure.Procedure.Description)
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, *resp.CreateProcedure.Procedure.Status)
			} else {
				assert.Empty(t, resp.CreateProcedure.Procedure.Status)
			}

			if tc.request.ProcedureType != nil {
				assert.Equal(t, *tc.request.ProcedureType, *resp.CreateProcedure.Procedure.ProcedureType)
			} else {
				assert.Empty(t, resp.CreateProcedure.Procedure.ProcedureType)
			}

			if tc.request.Version != nil {
				assert.Equal(t, *tc.request.Version, *resp.CreateProcedure.Procedure.Version)
			} else {
				assert.Empty(t, resp.CreateProcedure.Procedure.Version)
			}

			if tc.request.PurposeAndScope != nil {
				assert.Equal(t, *tc.request.PurposeAndScope, *resp.CreateProcedure.Procedure.PurposeAndScope)
			} else {
				assert.Empty(t, resp.CreateProcedure.Procedure.PurposeAndScope)
			}

			if tc.request.Background != nil {
				assert.Equal(t, *tc.request.Background, *resp.CreateProcedure.Procedure.Background)
			} else {
				assert.Empty(t, resp.CreateProcedure.Procedure.Background)
			}

			if tc.request.Satisfies != nil {
				assert.Equal(t, *tc.request.Satisfies, *resp.CreateProcedure.Procedure.Satisfies)
			} else {
				assert.Empty(t, resp.CreateProcedure.Procedure.Satisfies)
			}

			if tc.request.Details != nil {
				assert.Equal(t, tc.request.Details, resp.CreateProcedure.Procedure.Details)
			} else {
				assert.Empty(t, resp.CreateProcedure.Procedure.Details)
			}

			if tc.request.EditorIDs != nil {
				assert.Len(t, resp.CreateProcedure.Procedure.Editors, len(tc.request.EditorIDs))
			} else {
				assert.Empty(t, resp.CreateProcedure.Procedure.Editors)
			}

			if tc.request.BlockedGroupIDs != nil {
				assert.Len(t, resp.CreateProcedure.Procedure.BlockedGroups, len(tc.request.BlockedGroupIDs))
			} else {
				assert.Empty(t, resp.CreateProcedure.Procedure.BlockedGroups)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateProcedure() {
	t := suite.T()

	// create procedure to be updated
	Procedure := (&ProcedureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// create another admin user and add them to the same organization and group as testUser1
	// this will allow us to test the group editor permissions
	anotherAdminUser := suite.userBuilder(context.Background())
	suite.addUserToOrganization(&anotherAdminUser, enums.RoleAdmin, testUser1.OrganizationID)

	(&GroupMemberBuilder{client: suite.client, UserID: anotherAdminUser.ID, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)

	// create a viewer user and add them to the same organization as testUser1
	// also add them to the same group as testUser1, this should still allow them to edit the procedure
	// despite not not being an organization admin
	anotherViewerUser := suite.userBuilder(context.Background())
	suite.addUserToOrganization(&anotherViewerUser, enums.RoleMember, testUser1.OrganizationID)

	(&GroupMemberBuilder{client: suite.client, UserID: anotherViewerUser.ID, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)

	// create one more group that will be used to test the blocked group permissions and add anotherViewerUser to it
	blockGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&GroupMemberBuilder{client: suite.client, UserID: anotherViewerUser.ID, GroupID: blockGroup.ID}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateProcedureInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update name field, and add group",
			request: openlaneclient.UpdateProcedureInput{
				Name:         lo.ToPtr("Updated Procedure Name"),
				AddEditorIDs: []string{testUser1.GroupID}, // add the group to the editor groups for subsequent tests
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: openlaneclient.UpdateProcedureInput{
				Status:      lo.ToPtr("published"),
				Description: lo.ToPtr("Updated description"),
				Version:     lo.ToPtr("v2.0"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, not enough permissions",
			request: openlaneclient.UpdateProcedureInput{
				Name: lo.ToPtr("Updated Procedure Name"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, not enough permissions",
			request: openlaneclient.UpdateProcedureInput{
				Name: lo.ToPtr("Updated Procedure Name Meow"),
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx, // admin users do not automatically inherit permissions
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update allowed, user in editor group",
			request: openlaneclient.UpdateProcedureInput{
				Name: lo.ToPtr("Updated Procedure Name Again"),
			},
			client: suite.client.api,
			ctx:    anotherAdminUser.UserCtx, // user assigned to the group which has editor permissions
		},
		{
			name: "member update allowed, user in editor group",
			request: openlaneclient.UpdateProcedureInput{
				Name: lo.ToPtr("Updated Procedure Name Again"),
			},
			client: suite.client.api,
			ctx:    anotherViewerUser.UserCtx, // user assigned to the group which has editor permissions
		},
		{
			name: "happy path, block the group from editing",
			request: openlaneclient.UpdateProcedureInput{
				AddBlockedGroupIDs: []string{blockGroup.ID}, // block the group
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "member update no longer allowed, user in blocked group",
			request: openlaneclient.UpdateProcedureInput{
				Name: lo.ToPtr("Updated Procedure Name Again"),
			},
			client:      suite.client.api,
			ctx:         anotherViewerUser.UserCtx, // user assigned to the group which was blocked
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "happy path, remove the group",
			request: openlaneclient.UpdateProcedureInput{
				RemoveEditorIDs: []string{testUser1.GroupID}, // remove the group from the editor groups
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "update not allowed, editor group was removed",
			request: openlaneclient.UpdateProcedureInput{
				Name: lo.ToPtr("Updated Procedure Name Again Again"),
			},
			client:      suite.client.api,
			ctx:         anotherAdminUser.UserCtx, // user assigned to the group which no longer has editor permissions
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: openlaneclient.UpdateProcedureInput{
				Description: lo.ToPtr("Updated description"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateProcedure(tc.ctx, Procedure.ID, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// check updated fields
			if tc.request.Name != nil {
				assert.Equal(t, *tc.request.Name, resp.UpdateProcedure.Procedure.Name)
			}

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.UpdateProcedure.Procedure.Description)
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, *resp.UpdateProcedure.Procedure.Status)
			}

			if tc.request.ProcedureType != nil {
				assert.Equal(t, *tc.request.ProcedureType, *resp.UpdateProcedure.Procedure.ProcedureType)
			}

			if tc.request.Version != nil {
				assert.Equal(t, *tc.request.Version, *resp.UpdateProcedure.Procedure.Version)
			}

			if tc.request.PurposeAndScope != nil {
				assert.Equal(t, *tc.request.PurposeAndScope, *resp.UpdateProcedure.Procedure.PurposeAndScope)
			}

			if tc.request.Background != nil {
				assert.Equal(t, *tc.request.Background, *resp.UpdateProcedure.Procedure.Background)
			}

			if tc.request.Satisfies != nil {
				assert.Equal(t, *tc.request.Satisfies, *resp.UpdateProcedure.Procedure.Satisfies)
			}

			if tc.request.Details != nil {
				assert.Equal(t, tc.request.Details, resp.UpdateProcedure.Procedure.Details)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteProcedure() {
	t := suite.T()

	// create Procedures to be deleted
	Procedure1 := (&ProcedureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	Procedure2 := (&ProcedureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  Procedure1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: Procedure1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  Procedure1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: Procedure2.ID,
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
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteProcedure.DeletedID)
		})
	}
}
