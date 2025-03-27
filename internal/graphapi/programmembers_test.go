package graphapi_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestMutationCreateProgramMembers() {
	t := suite.T()

	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	orgMember1 := (&OrgMemberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	orgMember2 := (&OrgMemberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	orgMember3 := (&OrgMemberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name      string
		programID string
		userID    string
		role      enums.Role
		client    *openlaneclient.OpenlaneClient
		ctx       context.Context
		errMsg    string
	}{
		{
			name:      "happy path, add admin",
			programID: program.ID,
			userID:    orgMember1.UserID,
			role:      enums.RoleAdmin,
			client:    suite.client.api,
			ctx:       testUser1.UserCtx,
		},
		{
			name:      "happy path, add member using personal access token",
			programID: program.ID,
			userID:    orgMember3.UserID,
			role:      enums.RoleMember,
			client:    suite.client.apiWithPAT,
			ctx:       context.Background(),
		},
		{
			name:      "cannot add self to program",
			programID: program.ID,
			userID:    adminUser.UserInfo.ID,
			role:      enums.RoleAdmin,
			client:    suite.client.api,
			ctx:       adminUser.UserCtx,
			errMsg:    notAuthorizedErrorMsg,
		},
		{
			name:      "add member, no access",
			programID: program.ID,
			userID:    orgMember2.UserID,
			role:      enums.RoleMember,
			client:    suite.client.api,
			ctx:       viewOnlyUser.UserCtx,
			errMsg:    notAuthorizedErrorMsg,
		},
		{
			name:      "owner relation not valid for programs",
			programID: program.ID,
			userID:    orgMember2.UserID,
			role:      enums.RoleOwner,
			client:    suite.client.api,
			ctx:       testUser1.UserCtx,
			errMsg:    "OWNER is not a valid ProgramMembershipRole",
		},
		{
			name:      "duplicate user, different role",
			programID: program.ID,
			userID:    orgMember1.UserID,
			role:      enums.RoleMember,
			client:    suite.client.api,
			ctx:       testUser1.UserCtx,
			errMsg:    "already exists",
		},
		{
			name:      "invalid user",
			programID: program.ID,
			userID:    "not-a-valid-user-id",
			role:      enums.RoleMember,
			client:    suite.client.api,
			ctx:       testUser1.UserCtx,
			errMsg:    "user not in organization",
		},
		{
			name:      "invalid program",
			programID: "not-a-valid-program-id",
			userID:    orgMember1.UserID,
			role:      enums.RoleMember,
			client:    suite.client.api,
			ctx:       testUser1.UserCtx,
			errMsg:    notAuthorizedErrorMsg,
		},
		{
			name:      "invalid role",
			programID: program.ID,
			userID:    orgMember1.UserID,
			role:      enums.RoleInvalid,
			client:    suite.client.api,
			ctx:       testUser1.UserCtx,
			errMsg:    "not a valid ProgramMembershipRole",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			role := tc.role
			input := openlaneclient.CreateProgramMembershipInput{
				ProgramID: tc.programID,
				UserID:    tc.userID,
				Role:      &role,
			}

			resp, err := tc.client.AddUserToProgramWithRole(tc.ctx, input)

			if tc.errMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errMsg)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.CreateProgramMembership)
			assert.Equal(t, tc.userID, resp.CreateProgramMembership.ProgramMembership.UserID)
			assert.Equal(t, tc.programID, resp.CreateProgramMembership.ProgramMembership.ProgramID)
			assert.Equal(t, tc.role, resp.CreateProgramMembership.ProgramMembership.Role)
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateProgramMembers() {
	t := suite.T()

	pm := (&ProgramMemberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// get all program members so we know the id of the test user program member
	programMembers, err := suite.client.api.GetProgramMembersByProgramID(testUser1.UserCtx, &openlaneclient.ProgramMembershipWhereInput{
		ProgramID: &pm.ProgramID,
	})
	require.NoError(t, err)

	testUser1ProgramMember := ""
	for _, pm := range programMembers.ProgramMemberships.Edges {
		if pm.Node.UserID == testUser1.UserInfo.ID {
			testUser1ProgramMember = pm.Node.ID
			break
		}
	}

	testCases := []struct {
		name            string
		programMemberID string
		role            enums.Role
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		errMsg          string
	}{
		{
			name:            "happy path, update to admin from member",
			programMemberID: pm.ID,
			role:            enums.RoleAdmin,
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
		},
		{
			name:            "update self from admin to member, not allowed",
			programMemberID: testUser1ProgramMember,
			role:            enums.RoleMember,
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			errMsg:          notAuthorizedErrorMsg,
		},
		{
			name:            "happy path, update to admin from member using personal access token",
			programMemberID: pm.ID,
			role:            enums.RoleAdmin,
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
		},
		{
			name:            "invalid role",
			programMemberID: pm.ID,
			role:            enums.RoleInvalid,
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			errMsg:          "not a valid ProgramMembershipRole",
		},
		{
			name:            "no access",
			programMemberID: pm.ID,
			role:            enums.RoleMember,
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			errMsg:          notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			role := tc.role
			input := openlaneclient.UpdateProgramMembershipInput{
				Role: &role,
			}

			resp, err := tc.client.UpdateUserRoleInProgram(tc.ctx, tc.programMemberID, input)

			if tc.errMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errMsg)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.UpdateProgramMembership)
			assert.Equal(t, tc.role, resp.UpdateProgramMembership.ProgramMembership.Role)
		})
	}
}
