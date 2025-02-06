package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/ulids"
)

func (suite *GraphTestSuite) TestQueryProgram() {
	t := suite.T()

	// create program with a linked procedure and policy
	program := (&ProgramBuilder{client: suite.client, WithProcedure: true, WithPolicy: true}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: program.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: program.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "no access, user of same org",
			queryID:  program.ID,
			client:   suite.client.api,
			ctx:      viewOnlyUser.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "no access, user of different org",
			queryID:  program.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetProgramByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Equal(t, program.ID, resp.Program.ID)
			assert.Equal(t, program.Name, resp.Program.Name)
			assert.Len(t, resp.Program.Procedures, 1)
			assert.Len(t, resp.Program.InternalPolicies, 1)
		})
	}
}

func (suite *GraphTestSuite) TestQueryPrograms() {
	t := suite.T()

	// programs for the first organization with a linked procedure and policy
	(&ProgramBuilder{client: suite.client, WithProcedure: true, WithPolicy: true}).MustNew(testUser1.UserCtx, t)
	(&ProgramBuilder{client: suite.client, WithProcedure: true, WithPolicy: true}).MustNew(testUser1.UserCtx, t)

	// program created by an admin user of the first organization with a linked procedure and policy
	(&ProgramBuilder{client: suite.client, WithProcedure: true, WithPolicy: true}).MustNew(adminUser.UserCtx, t)

	// program for the other organization with a linked procedure and policy
	(&ProgramBuilder{client: suite.client, WithProcedure: true, WithPolicy: true}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		expectedResults int
		errorMsg        string
	}{
		{
			name:            "happy path, org owner should see all programs (3)",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 3,
		},
		{
			name:            "happy path using personal access token",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 3,
		},
		{
			name:            "view only user has not been added to any programs",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 0,
		},
		{
			name:            "admin user should see the program they created",
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: 1,
		},
		{
			name:            "owner of the other organization should see the program they created",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllPrograms(tc.ctx)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Len(t, resp.Programs.Edges, tc.expectedResults)

			for _, edge := range resp.Programs.Edges {
				require.NotNil(t, edge.Node)
				assert.Len(t, edge.Node.Procedures, 1)
				assert.Len(t, edge.Node.InternalPolicies, 1)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateProgram() {
	t := suite.T()

	startDate := time.Now().AddDate(0, 0, 1)
	endDate := time.Now().AddDate(0, 0, 360)

	groupMember := (&GroupMemberBuilder{client: suite.client, UserID: viewOnlyUser.ID}).MustNew(testUser1.UserCtx, t)

	// Create some edge objects
	procedure := (&ProcedureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	policy := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	blockedGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	viewerGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// group that the user does not have access to (for testing permissions)
	anotherGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name          string
		request       openlaneclient.CreateProgramInput
		addGroupToOrg bool
		client        *openlaneclient.OpenlaneClient
		ctx           context.Context
		expectedErr   string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateProgramInput{
				Name: "mitb program",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all basic input",
			request: openlaneclient.CreateProgramInput{
				Name:                 "mitb program",
				Description:          lo.ToPtr("being the best"),
				Status:               &enums.ProgramStatusInProgress,
				StartDate:            &startDate,
				EndDate:              &endDate,
				AuditorReady:         lo.ToPtr(false),
				AuditorWriteComments: lo.ToPtr(true),
				AuditorReadComments:  lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, edges",
			request: openlaneclient.CreateProgramInput{
				Name:              "mitb program",
				ProcedureIDs:      []string{procedure.ID},
				InternalPolicyIDs: []string{policy.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "add editor group",
			request: openlaneclient.CreateProgramInput{
				Name:            "Test Program MITB",
				EditorIDs:       []string{testUser1.GroupID},
				BlockedGroupIDs: []string{blockedGroup.ID},
				ViewerIDs:       []string{viewerGroup.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "add editor group, no access to group",
			request: openlaneclient.CreateProgramInput{
				Name:      "Test Program Meow",
				EditorIDs: []string{anotherGroup.ID},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "happy path, using pat",
			request: openlaneclient.CreateProgramInput{
				Name:        "mitb program",
				Description: lo.ToPtr("being the best"),
				OwnerID:     &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: openlaneclient.CreateProgramInput{
				Name:        "mitb program",
				Description: lo.ToPtr("being the best"),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: openlaneclient.CreateProgramInput{
				Name: "mitb program",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user now authorized, added to group with creator permissions",
			request: openlaneclient.CreateProgramInput{
				Name: "mitb program",
			},
			addGroupToOrg: true,
			client:        suite.client.api,
			ctx:           viewOnlyUser.UserCtx,
		},
		{
			name: "user not authorized, no permissions",
			request: openlaneclient.CreateProgramInput{
				Name:    "mitb program",
				OwnerID: &testUser1.OrganizationID,
			},
			client: suite.client.api,
			ctx:    testUser2.UserCtx,
		},
		{
			name: "missing required field",
			request: openlaneclient.CreateProgramInput{
				Description: lo.ToPtr("soc2 2024"),
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
					openlaneclient.UpdateOrganizationInput{
						AddProgramCreatorIDs: []string{groupMember.GroupID},
					}, nil)
				require.NoError(t, err)
			}

			resp, err := tc.client.CreateProgram(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// check required fields
			assert.Equal(t, tc.request.Name, resp.CreateProgram.Program.Name)

			assert.NotEmpty(t, resp.CreateProgram.Program.DisplayID)
			assert.Contains(t, resp.CreateProgram.Program.DisplayID, "PRG-")

			// ensure the owner is set to the user's organization, not the  input
			if tc.request.OwnerID != nil && tc.ctx == testUser2.UserCtx {
				assert.Equal(t, testUser2.OrganizationID, *resp.CreateProgram.Program.OwnerID)
			}

			// check optional fields
			if tc.request.Description == nil {
				assert.Empty(t, resp.CreateProgram.Program.Description)
			} else {
				assert.Equal(t, tc.request.Description, resp.CreateProgram.Program.Description)
			}

			if tc.request.Status == nil {
				assert.Equal(t, enums.ProgramStatusNotStarted, resp.CreateProgram.Program.Status)
			} else {
				assert.Equal(t, *tc.request.Status, resp.CreateProgram.Program.Status)
			}

			if tc.request.StartDate == nil {
				assert.Empty(t, resp.CreateProgram.Program.StartDate)
			} else {
				assert.WithinDuration(t, startDate, *resp.CreateProgram.Program.StartDate, 1*time.Minute)
			}

			if tc.request.EndDate == nil {
				assert.Empty(t, resp.CreateProgram.Program.EndDate)
			} else {
				assert.WithinDuration(t, endDate, *resp.CreateProgram.Program.EndDate, 1*time.Minute)
			}

			if tc.request.AuditorReady == nil {
				assert.False(t, resp.CreateProgram.Program.AuditorReady)
			} else {
				assert.Equal(t, *tc.request.AuditorReady, resp.CreateProgram.Program.AuditorReady)
			}

			if tc.request.AuditorWriteComments == nil {
				assert.False(t, resp.CreateProgram.Program.AuditorWriteComments)
			} else {
				assert.Equal(t, *tc.request.AuditorWriteComments, resp.CreateProgram.Program.AuditorWriteComments)
			}

			if tc.request.AuditorReadComments == nil {
				assert.False(t, resp.CreateProgram.Program.AuditorReadComments)
			} else {
				assert.Equal(t, *tc.request.AuditorReadComments, resp.CreateProgram.Program.AuditorReadComments)
			}

			// check edges
			if len(tc.request.ProcedureIDs) > 0 {
				require.Len(t, resp.CreateProgram.Program.Procedures, 1)
				for _, edge := range resp.CreateProgram.Program.Procedures {
					assert.Equal(t, procedure.ID, edge.ID)
				}
			}

			if len(tc.request.InternalPolicyIDs) > 0 {
				require.Len(t, resp.CreateProgram.Program.InternalPolicies, 1)
				for _, edge := range resp.CreateProgram.Program.InternalPolicies {
					assert.Equal(t, policy.ID, edge.ID)
				}
			}

			if len(tc.request.EditorIDs) > 0 {
				require.Len(t, resp.CreateProgram.Program.Editors, 1)
				for _, edge := range resp.CreateProgram.Program.Editors {
					assert.Equal(t, testUser1.GroupID, edge.ID)
				}
			}

			if len(tc.request.BlockedGroupIDs) > 0 {
				require.Len(t, resp.CreateProgram.Program.BlockedGroups, 1)
				for _, edge := range resp.CreateProgram.Program.BlockedGroups {
					assert.Equal(t, blockedGroup.ID, edge.ID)
				}
			}

			if len(tc.request.ViewerIDs) > 0 {
				require.Len(t, resp.CreateProgram.Program.Viewers, 1)
				for _, edge := range resp.CreateProgram.Program.Viewers {
					assert.Equal(t, viewerGroup.ID, edge.ID)
				}
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateProgram() {
	t := suite.T()

	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	programMembers, err := suite.client.api.GetProgramMembersByProgramID(testUser1.UserCtx, &openlaneclient.ProgramMembershipWhereInput{
		ProgramID: &program.ID,
	})

	require.NoError(t, err)

	testUserProgramMemberID := ""
	for _, pm := range programMembers.ProgramMemberships.Edges {
		if pm.Node.UserID == testUser1.ID {
			testUserProgramMemberID = pm.Node.ID
		}
	}

	// create program user to remove
	programUser := suite.userBuilder(context.Background())
	(&OrgMemberBuilder{client: suite.client, UserID: programUser.ID, OrgID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)

	pm := (&ProgramMemberBuilder{client: suite.client, UserID: programUser.ID, ProgramID: program.ID}).MustNew(testUser1.UserCtx, t)

	// Create some edge objects
	procedure1 := (&ProcedureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	policy1 := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// Create some edge objects for another organization
	procedure2 := (&ProcedureBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	policy2 := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	// create another admin user and add them to the same organization and group as testUser1
	// this will allow us to test the group editor permissions
	anotherAdminUser := suite.userBuilder(context.Background())
	suite.addUserToOrganization(testUser1.UserCtx, &anotherAdminUser, enums.RoleAdmin, testUser1.OrganizationID)

	(&GroupMemberBuilder{client: suite.client, UserID: anotherAdminUser.ID, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)

	// create a viewer user and add them to the same organization as testUser1
	// also add them to the same group as testUser1, this should still allow them to edit the policy
	// despite not not being an organization admin
	anotherViewerUser := suite.userBuilder(context.Background())
	suite.addUserToOrganization(testUser1.UserCtx, &anotherViewerUser, enums.RoleMember, testUser1.OrganizationID)

	(&GroupMemberBuilder{client: suite.client, UserID: anotherViewerUser.ID, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)

	// create one more group that will be used to test the blocked group permissions and add anotherViewerUser to it
	blockGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&GroupMemberBuilder{client: suite.client, UserID: anotherViewerUser.ID, GroupID: blockGroup.ID}).MustNew(testUser1.UserCtx, t)

	// create a view only user and add them to the same organization as testUser1
	meowViewerUser := suite.userBuilder(context.Background())
	suite.addUserToOrganization(testUser1.UserCtx, &meowViewerUser, enums.RoleMember, testUser1.OrganizationID)

	// create one more group that will be used to test the blocked group permissions and add anotherViewerUser to it
	viewerGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&GroupMemberBuilder{client: suite.client, UserID: meowViewerUser.ID, GroupID: blockGroup.ID}).MustNew(testUser1.UserCtx, t)

	// add add user to the viewer group
	(&GroupMemberBuilder{client: suite.client, UserID: viewOnlyUser.ID, GroupID: viewerGroup.ID}).MustNew(testUser1.UserCtx, t)

	// ensure the user does not currently have access to the program
	res, err := suite.client.api.GetProgramByID(viewOnlyUser.UserCtx, program.ID)
	require.Error(t, err)
	require.Nil(t, res)

	testCases := []struct {
		name              string
		request           openlaneclient.UpdateProgramInput
		client            *openlaneclient.OpenlaneClient
		ctx               context.Context
		expectedErr       string
		expectedEdgeCount int
	}{
		{
			name: "happy path, update field",
			request: openlaneclient.UpdateProgramInput{
				Description:  lo.ToPtr("new description"),
				AddEditorIDs: []string{testUser1.GroupID}, // add the group to the editor groups for the subsequent tests
				AddViewerIDs: []string{viewerGroup.ID},    // add the group to the viewer groups and ensure the user has access to the program
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields using pat",
			request: openlaneclient.UpdateProgramInput{
				Status:               &enums.ProgramStatusReadyForAuditor,
				EndDate:              lo.ToPtr(time.Now().AddDate(0, 0, 30)),
				AuditorReady:         lo.ToPtr(true),
				AuditorWriteComments: lo.ToPtr(true),
				AuditorReadComments:  lo.ToPtr(true),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "remove program member, cannot remove self",
			request: openlaneclient.UpdateProgramInput{
				RemoveProgramMembers: []string{testUserProgramMemberID},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "add program member, cannot add self",
			request: openlaneclient.UpdateProgramInput{
				AddProgramMembers: []*openlaneclient.CreateProgramMembershipInput{
					{
						UserID: adminUser.ID,
					},
				},
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "add program member, can add another user",
			request: openlaneclient.UpdateProgramInput{
				AddProgramMembers: []*openlaneclient.CreateProgramMembershipInput{
					{
						UserID: adminUser.ID,
					},
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, remove program member",
			request: openlaneclient.UpdateProgramInput{
				RemoveProgramMembers: []string{pm.ID},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, update edge - procedure",
			request: openlaneclient.UpdateProgramInput{
				AddProcedureIDs: []string{procedure1.ID},
			},
			client:            suite.client.api,
			ctx:               testUser1.UserCtx,
			expectedEdgeCount: 1,
		},
		{
			name: "happy path, update edge - policy",
			request: openlaneclient.UpdateProgramInput{
				AddInternalPolicyIDs: []string{policy1.ID},
			},
			client:            suite.client.api,
			ctx:               testUser1.UserCtx,
			expectedEdgeCount: 1,
		},
		{
			name: "update edge - procedure - not allowed to access procedure",
			request: openlaneclient.UpdateProgramInput{
				AddProcedureIDs: []string{procedure2.ID},
			},
			client:            suite.client.api,
			ctx:               testUser1.UserCtx,
			expectedEdgeCount: 0, // procedure is not visible to the user
		},
		{
			name: "update edge - policy - not allowed to access procedure",
			request: openlaneclient.UpdateProgramInput{
				AddInternalPolicyIDs: []string{policy2.ID},
			},
			client:            suite.client.api,
			ctx:               testUser1.UserCtx,
			expectedEdgeCount: 0, // policy is not visible to the user
		},
		{
			name: "update not allowed, not enough permissions",
			request: openlaneclient.UpdateProgramInput{
				Description: lo.ToPtr("newer description"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg, // user in in viewer group, but has no edit access
		},
		{
			name: "update not allowed, no permissions",
			request: openlaneclient.UpdateProgramInput{
				Description: lo.ToPtr("newer description"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "update allowed, user in editor group",
			request: openlaneclient.UpdateProgramInput{
				Description: lo.ToPtr("soc2 2024"),
			},
			client: suite.client.api,
			ctx:    anotherAdminUser.UserCtx, // user assigned to the group which has editor permissions
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateProgram(tc.ctx, program.ID, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// add checks for the updated fields if they were set in the request
			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.UpdateProgram.Program.Description)
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, resp.UpdateProgram.Program.Status)
			}

			if tc.request.StartDate != nil {
				assert.WithinDuration(t, *tc.request.StartDate, *resp.UpdateProgram.Program.StartDate, time.Minute)
			}

			if tc.request.EndDate != nil {
				assert.WithinDuration(t, *tc.request.EndDate, *resp.UpdateProgram.Program.EndDate, time.Minute)
			}

			if tc.request.AuditorReady != nil {
				assert.Equal(t, *tc.request.AuditorReady, resp.UpdateProgram.Program.AuditorReady)
			}

			if tc.request.AuditorWriteComments != nil {
				assert.Equal(t, *tc.request.AuditorWriteComments, resp.UpdateProgram.Program.AuditorWriteComments)
			}

			if tc.request.AuditorReadComments != nil {
				assert.Equal(t, *tc.request.AuditorReadComments, resp.UpdateProgram.Program.AuditorReadComments)
			}

			// check edges
			if len(tc.request.AddProcedureIDs) > 0 {
				require.Len(t, resp.UpdateProgram.Program.Procedures, 1)
				for _, edge := range resp.UpdateProgram.Program.Procedures {
					assert.Equal(t, procedure1.ID, edge.ID)
				}
			}

			if len(tc.request.AddInternalPolicyIDs) > 0 {
				require.Len(t, resp.UpdateProgram.Program.InternalPolicies, 1)
				for _, edge := range resp.UpdateProgram.Program.InternalPolicies {
					assert.Equal(t, policy1.ID, edge.ID)
				}
			}

			if len(tc.request.AddEditorIDs) > 0 {
				require.Len(t, resp.UpdateProgram.Program.Editors, 1)
				for _, edge := range resp.UpdateProgram.Program.Editors {
					assert.Equal(t, testUser1.GroupID, edge.ID)
				}
			}

			if len(tc.request.AddBlockedGroupIDs) > 0 {
				require.Len(t, resp.UpdateProgram.Program.BlockedGroups, 1)
				for _, edge := range resp.UpdateProgram.Program.BlockedGroups {
					assert.Equal(t, blockGroup.ID, edge.ID)
				}
			}

			if len(tc.request.AddViewerIDs) > 0 {
				require.Len(t, resp.UpdateProgram.Program.Viewers, 1)
				for _, edge := range resp.UpdateProgram.Program.Viewers {
					assert.Equal(t, viewerGroup.ID, edge.ID)
				}

				// ensure the user has access to the program now
				res, err := suite.client.api.GetProgramByID(viewOnlyUser.UserCtx, program.ID)
				require.NoError(t, err)
				require.NotEmpty(t, res)
				assert.Equal(t, program.ID, res.Program.ID)
			}

			if len(tc.request.AddProgramMembers) > 0 {
				require.Len(t, resp.UpdateProgram.Program.Members, 3)

				// it should have the owner and the admin user and the other user added in the test setup
				require.Equal(t, testUser1.ID, resp.UpdateProgram.Program.Members[0].User.ID)
				require.Equal(t, programUser.ID, resp.UpdateProgram.Program.Members[1].User.ID)
				require.Equal(t, adminUser.ID, resp.UpdateProgram.Program.Members[2].User.ID)
			}

			// member was removed, ensure there are two members left
			if len(tc.request.RemoveProgramMembers) > 0 {
				require.Len(t, resp.UpdateProgram.Program.Members, 2)

				// it should have the owner and the admin user
				require.Equal(t, testUser1.ID, resp.UpdateProgram.Program.Members[0].User.ID)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteProgram() {
	t := suite.T()

	// create Programs to be deleted
	program1 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete program",
			idToDelete:  program1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete program",
			idToDelete: program1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "Program already deleted, not found",
			idToDelete:  program1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete program using personal access token",
			idToDelete: program2.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown program, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteProgram(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteProgram.DeletedID)
		})
	}
}
