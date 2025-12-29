package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestMutationCreateProgramMembers(t *testing.T) {
	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	orgMember1 := (&OrgMemberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	orgMember2 := (&OrgMemberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	orgMember3 := (&OrgMemberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name      string
		programID string
		userID    string
		role      enums.Role
		client    *testclient.TestClient
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
			input := testclient.CreateProgramMembershipInput{
				ProgramID: tc.programID,
				UserID:    tc.userID,
				Role:      &role,
			}

			resp, err := tc.client.AddUserToProgramWithRole(tc.ctx, input)

			if tc.errMsg != "" {
				assert.ErrorContains(t, err, tc.errMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.userID, resp.CreateProgramMembership.ProgramMembership.UserID))
			assert.Check(t, is.Equal(tc.programID, resp.CreateProgramMembership.ProgramMembership.ProgramID))
			assert.Check(t, is.Equal(tc.role, resp.CreateProgramMembership.ProgramMembership.Role))
		})
	}

	// cleanup program
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: program.ID}).MustDelete(testUser1.UserCtx, t)
	// cleanup org members
	(&Cleanup[*generated.OrgMembershipDeleteOne]{client: suite.client.db.OrgMembership, IDs: []string{orgMember1.ID, orgMember2.ID, orgMember3.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateProgramMembers(t *testing.T) {
	pm := (&ProgramMemberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// get all program members so we know the id of the test user program member
	programMembers, err := suite.client.api.GetProgramMembersByProgramID(testUser1.UserCtx, &testclient.ProgramMembershipWhereInput{
		ProgramID: &pm.ProgramID,
	})
	assert.NilError(t, err)

	testUser1ProgramMember := ""
	for _, pm := range programMembers.ProgramMemberships.Edges {
		if pm.Node.UserID == testUser1.UserInfo.ID {
			testUser1ProgramMember = pm.Node.ID
			break
		}
	}

	// add an admin user to the program as member
	(&ProgramMemberBuilder{client: suite.client, UserID: adminUser.ID, ProgramID: pm.ProgramID, Role: enums.RoleMember.String()}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name            string
		programMemberID string
		role            enums.Role
		client          *testclient.TestClient
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
			name:            "update self from admin to member allowed because user is org owner",
			programMemberID: testUser1ProgramMember,
			role:            enums.RoleMember,
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
		},
		{
			name:            "update self from member to admin of self not allowed",
			programMemberID: testUser1ProgramMember,
			role:            enums.RoleAdmin,
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
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
			input := testclient.UpdateProgramMembershipInput{
				Role: &role,
			}

			resp, err := tc.client.UpdateUserRoleInProgram(tc.ctx, tc.programMemberID, input)
			if tc.errMsg != "" {
				assert.ErrorContains(t, err, tc.errMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.role, resp.UpdateProgramMembership.ProgramMembership.Role))
		})
	}

	// cleanup program
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: pm.ProgramID}).MustDelete(testUser1.UserCtx, t)
	// cleanup org members
	(&Cleanup[*generated.OrgMembershipDeleteOne]{client: suite.client.db.OrgMembership, IDs: []string{pm.Edges.OrgMembership.ID}}).MustDelete(testUser1.UserCtx, t)
}
