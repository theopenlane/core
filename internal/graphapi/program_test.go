package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryProgram(t *testing.T) {
	// create program1 with a linked procedure and policy
	program1 := (&ProgramBuilder{client: suite.client, WithProcedure: true, WithPolicy: true}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client, WithProcedure: true, WithPolicy: true}).MustNew(adminUser.UserCtx, t)

	archivedProgram := (&ProgramBuilder{client: suite.client, WithProcedure: true, WithPolicy: true, Status: enums.ProgramStatusArchived}).MustNew(adminUser.UserCtx, t)

	anonymousContext := createAnonymousTrustCenterContext(ulids.New().String(), testUser1.OrganizationID)

	testCases := []struct {
		name           string
		queryID        string
		client         *testclient.TestClient
		ctx            context.Context
		expectedResult *generated.Program
		errorMsg       string
	}{
		{
			name:           "happy path",
			queryID:        program1.ID,
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
			expectedResult: program1,
		},
		{
			name:           "happy path, program created by admin user",
			queryID:        program2.ID,
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
			expectedResult: program2,
		},
		{
			name:           "happy path using personal access token",
			queryID:        program1.ID,
			client:         suite.client.apiWithPAT,
			ctx:            context.Background(),
			expectedResult: program1,
		},
		{
			name:           "archived program - happy path using personal access token",
			queryID:        archivedProgram.ID,
			client:         suite.client.apiWithPAT,
			ctx:            context.Background(),
			expectedResult: archivedProgram,
		},
		{
			name:     "no access, user of same org",
			queryID:  program1.ID,
			client:   suite.client.api,
			ctx:      viewOnlyUser.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "no access, user of different org",
			queryID:  program1.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "no access, anonymous user",
			client:   suite.client.api,
			ctx:      anonymousContext,
			queryID:  program1.ID,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetProgramByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.expectedResult.ID, resp.Program.ID))
			assert.Check(t, is.Equal(tc.expectedResult.Name, resp.Program.Name))
			assert.Check(t, is.Len(resp.Program.Procedures.Edges, 1))
			assert.Check(t, is.Len(resp.Program.InternalPolicies.Edges, 1))
		})
	}

	// cleanup
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, IDs: []string{program1.ID, program2.ID}}).MustDelete(testUser1.UserCtx, t)
	// cleanup procedure and policy
	procedureIDs := []string{}
	for _, p := range program1.Edges.Procedures {
		procedureIDs = append(procedureIDs, p.ID)
	}
	policyIDs := []string{}
	for _, p := range program1.Edges.InternalPolicies {
		policyIDs = append(policyIDs, p.ID)
	}

	(&Cleanup[*generated.ProcedureDeleteOne]{client: suite.client.db.Procedure, IDs: procedureIDs}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.InternalPolicyDeleteOne]{client: suite.client.db.InternalPolicy, IDs: policyIDs}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryPrograms(t *testing.T) {
	// programs for the first organization with a linked procedure and policy
	program1 := (&ProgramBuilder{client: suite.client, WithProcedure: true, WithPolicy: true}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client, WithProcedure: true, WithPolicy: true}).MustNew(testUser1.UserCtx, t)

	// program created by an admin user of the first organization with a linked procedure and policy
	program3 := (&ProgramBuilder{client: suite.client, WithProcedure: true, WithPolicy: true}).MustNew(adminUser.UserCtx, t)

	// archived program for the first organization
	archivedProgram := (&ProgramBuilder{client: suite.client, WithProcedure: true, WithPolicy: true, Status: enums.ProgramStatusArchived}).MustNew(testUser1.UserCtx, t)

	// program for the other organization with a linked procedure and policy
	anotherUser := suite.userBuilder(context.Background(), t)
	program4 := (&ProgramBuilder{client: suite.client, WithProcedure: true, WithPolicy: true}).MustNew(anotherUser.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
		errorMsg        string
	}{
		{
			name:            "happy path, org owner should see all programs",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 3, // archived programs not listed by default
		},
		{
			name:            "happy path using personal access token",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 3, // archived programs not listed by default
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
			ctx:             anotherUser.UserCtx,
			expectedResults: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllPrograms(tc.ctx)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Len(resp.Programs.Edges, tc.expectedResults))

			for _, edge := range resp.Programs.Edges {
				assert.Assert(t, edge.Node != nil)
				assert.Check(t, is.Len(edge.Node.Procedures.Edges, 1))
				assert.Check(t, is.Len(edge.Node.InternalPolicies.Edges, 1))
			}
		})
	}

	// cleanup
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, IDs: []string{program1.ID, program2.ID, program3.ID, archivedProgram.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: program4.ID}).MustDelete(anotherUser.UserCtx, t)

	// cleanup procedures and policies
	procedureIDs := []string{}
	for _, p := range program1.Edges.Procedures {
		procedureIDs = append(procedureIDs, p.ID)
	}

	for _, p := range program2.Edges.Procedures {
		procedureIDs = append(procedureIDs, p.ID)
	}

	for _, p := range program3.Edges.Procedures {
		procedureIDs = append(procedureIDs, p.ID)
	}

	for _, p := range archivedProgram.Edges.Procedures {
		procedureIDs = append(procedureIDs, p.ID)
	}

	policyIDs := []string{}
	for _, p := range program1.Edges.InternalPolicies {
		policyIDs = append(policyIDs, p.ID)
	}

	for _, p := range program2.Edges.InternalPolicies {
		policyIDs = append(policyIDs, p.ID)
	}

	for _, p := range program3.Edges.InternalPolicies {
		policyIDs = append(policyIDs, p.ID)
	}

	for _, p := range archivedProgram.Edges.InternalPolicies {
		policyIDs = append(policyIDs, p.ID)
	}

	(&Cleanup[*generated.ProcedureDeleteOne]{client: suite.client.db.Procedure, IDs: procedureIDs}).MustDelete(testUser1.UserCtx, t)

	(&Cleanup[*generated.InternalPolicyDeleteOne]{client: suite.client.db.InternalPolicy, IDs: policyIDs}).MustDelete(testUser1.UserCtx, t)

	// we can ignore the cleanup for the new user, it won't conflict with other tests
}

func TestMutationCreateProgram(t *testing.T) {
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

	programIDsToCleanup := []string{}
	testCases := []struct {
		name          string
		request       testclient.CreateProgramInput
		addGroupToOrg bool
		client        *testclient.TestClient
		ctx           context.Context
		expectedErr   string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateProgramInput{
				Name: "mitb program",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all basic input",
			request: testclient.CreateProgramInput{
				Name:                 "mitb program",
				Description:          lo.ToPtr("being the best"),
				ProgramType:          &enums.ProgramTypeFramework,
				FrameworkName:        lo.ToPtr("SOC 2"),
				ProgramOwnerID:       &testUser1.ID,
				Status:               &enums.ProgramStatusInProgress,
				StartDate:            &startDate,
				EndDate:              &endDate,
				AuditorReady:         lo.ToPtr(false),
				AuditorWriteComments: lo.ToPtr(true),
				AuditorReadComments:  lo.ToPtr(true),
				AuditFirm:            lo.ToPtr("Meow Audit, LLC."),
				Auditor:              lo.ToPtr("Meowz Meow"),
				AuditorEmail:         lo.ToPtr("m@meow-audit.com"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, edges",
			request: testclient.CreateProgramInput{
				Name:              "mitb program",
				ProcedureIDs:      []string{procedure.ID},
				InternalPolicyIDs: []string{policy.ID},
				ProgramOwnerID:    &adminUser.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "add editor group",
			request: testclient.CreateProgramInput{
				Name:            "Test Program MITB",
				EditorIDs:       []string{testUser1.GroupID},
				BlockedGroupIDs: []string{blockedGroup.ID},
				ViewerIDs:       []string{viewerGroup.ID},
				ProgramOwnerID:  &testUser1.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "add editor group, no access to group",
			request: testclient.CreateProgramInput{
				Name:      "Test Program Meow",
				EditorIDs: []string{anotherGroup.ID},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateProgramInput{
				Name:        "mitb program",
				Description: lo.ToPtr("being the best"),
				ProgramType: &enums.ProgramTypeGapAnalysis,
				OwnerID:     &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateProgramInput{
				Name:        "mitb program",
				Description: lo.ToPtr("being the best"),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateProgramInput{
				Name: "mitb program",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user now authorized, added to group with creator permissions",
			request: testclient.CreateProgramInput{
				Name: "mitb program",
			},
			addGroupToOrg: true,
			client:        suite.client.api,
			ctx:           viewOnlyUser.UserCtx,
		},
		{
			name: "missing required field",
			request: testclient.CreateProgramInput{
				Description: lo.ToPtr("soc2 2024"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "invalid auditor email",
			request: testclient.CreateProgramInput{
				Name:         "mitb program",
				AuditorEmail: lo.ToPtr("invalid email"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "validator failed for field",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.addGroupToOrg {
				_, err := suite.client.api.UpdateOrganization(testUser1.UserCtx, testUser1.OrganizationID,
					testclient.UpdateOrganizationInput{
						AddProgramCreatorIDs: []string{groupMember.GroupID},
					}, nil)
				assert.NilError(t, err)
			}

			resp, err := tc.client.CreateProgram(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			programIDsToCleanup = append(programIDsToCleanup, resp.CreateProgram.Program.ID)

			// check required fields
			assert.Check(t, is.Equal(tc.request.Name, resp.CreateProgram.Program.Name))

			assert.Check(t, len(resp.CreateProgram.Program.DisplayID) != 0)
			assert.Check(t, is.Contains(resp.CreateProgram.Program.DisplayID, "PRG-"))

			// ensure the owner is set to the user's organization, not the  input
			if tc.request.OwnerID != nil && tc.ctx == testUser2.UserCtx {
				assert.Check(t, is.Equal(testUser2.OrganizationID, *resp.CreateProgram.Program.OwnerID))
			}

			// check optional fields
			if tc.request.Description == nil {
				assert.Check(t, is.Len(*resp.CreateProgram.Program.Description, 0))
			} else {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.CreateProgram.Program.Description))
			}

			if tc.request.ProgramType == nil {
				assert.Check(t, is.Equal(enums.ProgramTypeFramework, resp.CreateProgram.Program.ProgramType))
			} else {
				assert.Check(t, is.Equal(*tc.request.ProgramType, resp.CreateProgram.Program.ProgramType))
			}

			if tc.request.FrameworkName == nil {
				assert.Check(t, is.Len(*resp.CreateProgram.Program.FrameworkName, 0))
			} else {
				assert.Check(t, is.Equal(*tc.request.FrameworkName, *resp.CreateProgram.Program.FrameworkName))
			}

			if tc.request.Status == nil {
				assert.Check(t, is.Equal(enums.ProgramStatusNotStarted, resp.CreateProgram.Program.Status))
			} else {
				assert.Check(t, is.Equal(*tc.request.Status, resp.CreateProgram.Program.Status))
			}

			if tc.request.StartDate == nil {
				assert.Check(t, resp.CreateProgram.Program.StartDate == nil)
			} else {
				assert.Assert(t, resp.CreateProgram.Program.StartDate != nil)
				diff := resp.CreateProgram.Program.StartDate.Sub(startDate)
				assert.Check(t, diff >= -2*time.Minute && diff <= 2*time.Minute, "time difference is not within 2 minutes")
			}

			if tc.request.EndDate == nil {
				assert.Check(t, resp.CreateProgram.Program.EndDate == nil)
			} else {
				assert.Assert(t, resp.CreateProgram.Program.EndDate != nil)
				diff := resp.CreateProgram.Program.EndDate.Sub(endDate)
				assert.Check(t, diff >= -2*time.Minute && diff <= 2*time.Minute, "time difference is not within 2 minutes")
			}

			if tc.request.AuditorReady == nil {
				assert.Check(t, !resp.CreateProgram.Program.AuditorReady)
			} else {
				assert.Check(t, is.Equal(*tc.request.AuditorReady, resp.CreateProgram.Program.AuditorReady))
			}

			if tc.request.AuditorWriteComments == nil {
				assert.Check(t, !resp.CreateProgram.Program.AuditorWriteComments)
			} else {
				assert.Check(t, is.Equal(*tc.request.AuditorWriteComments, resp.CreateProgram.Program.AuditorWriteComments))
			}

			if tc.request.AuditorReadComments == nil {
				assert.Check(t, !resp.CreateProgram.Program.AuditorReadComments)
			} else {
				assert.Check(t, is.Equal(*tc.request.AuditorReadComments, resp.CreateProgram.Program.AuditorReadComments))
			}

			if tc.request.AuditFirm == nil {
				assert.Check(t, is.Len(*resp.CreateProgram.Program.AuditFirm, 0))
			} else {
				assert.Check(t, is.Equal(*tc.request.AuditFirm, *resp.CreateProgram.Program.AuditFirm))
			}

			if tc.request.Auditor == nil {
				assert.Check(t, is.Len(*resp.CreateProgram.Program.Auditor, 0))
			} else {
				assert.Check(t, is.Equal(*tc.request.Auditor, *resp.CreateProgram.Program.Auditor))
			}

			if tc.request.AuditorEmail == nil {
				assert.Check(t, is.Len(*resp.CreateProgram.Program.AuditorEmail, 0))
			} else {
				assert.Check(t, is.Equal(*tc.request.AuditorEmail, *resp.CreateProgram.Program.AuditorEmail))
			}

			// check edges
			if len(tc.request.ProcedureIDs) > 0 {
				assert.Assert(t, is.Len(resp.CreateProgram.Program.Procedures.Edges, 1))
				for _, edge := range resp.CreateProgram.Program.Procedures.Edges {
					assert.Check(t, is.Equal(procedure.ID, edge.Node.ID))
				}
			}

			if len(tc.request.InternalPolicyIDs) > 0 {
				assert.Assert(t, is.Len(resp.CreateProgram.Program.InternalPolicies.Edges, 1))
				for _, edge := range resp.CreateProgram.Program.InternalPolicies.Edges {
					assert.Check(t, is.Equal(policy.ID, edge.Node.ID))
				}
			}

			if len(tc.request.EditorIDs) > 0 {
				assert.Assert(t, is.Len(resp.CreateProgram.Program.Editors.Edges, 1))
				for _, edge := range resp.CreateProgram.Program.Editors.Edges {
					assert.Check(t, is.Equal(testUser1.GroupID, edge.Node.ID))
				}
			}

			if len(tc.request.BlockedGroupIDs) > 0 {
				assert.Assert(t, is.Len(resp.CreateProgram.Program.BlockedGroups.Edges, 1))
				for _, edge := range resp.CreateProgram.Program.BlockedGroups.Edges {
					assert.Check(t, is.Equal(blockedGroup.ID, edge.Node.ID))
				}
			}

			if len(tc.request.ViewerIDs) > 0 {
				assert.Assert(t, is.Len(resp.CreateProgram.Program.Viewers.Edges, 1))
				for _, edge := range resp.CreateProgram.Program.Viewers.Edges {
					assert.Check(t, is.Equal(viewerGroup.ID, edge.Node.ID))
				}
			}
		})
	}

	// cleanup policy and procedure
	(&Cleanup[*generated.ProcedureDeleteOne]{client: suite.client.db.Procedure, ID: procedure.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.InternalPolicyDeleteOne]{client: suite.client.db.InternalPolicy, ID: policy.ID}).MustDelete(testUser1.UserCtx, t)
	// cleanup group
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{groupMember.GroupID, blockedGroup.ID, viewerGroup.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, ID: anotherGroup.ID}).MustDelete(testUser2.UserCtx, t)

	// cleanup programs
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, IDs: programIDsToCleanup}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateProgram(t *testing.T) {
	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	archivedProgram := (&ProgramBuilder{client: suite.client, Status: enums.ProgramStatusArchived}).MustNew(testUser1.UserCtx, t)

	programMembers, err := suite.client.api.GetProgramMembersByProgramID(testUser1.UserCtx, &testclient.ProgramMembershipWhereInput{
		ProgramID: &program.ID,
	})
	assert.NilError(t, err)

	testUserProgramMemberID := ""
	for _, pm := range programMembers.ProgramMemberships.Edges {
		if pm.Node.UserID == testUser1.ID {
			testUserProgramMemberID = pm.Node.ID
		}
	}

	// create program user to remove
	programUser := suite.userBuilder(context.Background(), t)
	om := (&OrgMemberBuilder{client: suite.client, UserID: programUser.ID}).MustNew(testUser1.UserCtx, t)

	pm := (&ProgramMemberBuilder{client: suite.client, UserID: programUser.ID, ProgramID: program.ID}).MustNew(testUser1.UserCtx, t)

	// Create some edge objects
	procedure1 := (&ProcedureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	policy1 := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// Create some edge objects for another organization
	procedure2 := (&ProcedureBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	policy2 := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	// create another admin user and add them to the same organization and group as testUser1
	// this will allow us to test the group editor permissions
	anotherAdminUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser1.UserCtx, t, &anotherAdminUser, enums.RoleAdmin, testUser1.OrganizationID)

	gm1 := (&GroupMemberBuilder{client: suite.client, UserID: anotherAdminUser.ID, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)

	// create a viewer user and add them to the same organization as testUser1
	// also add them to the same group as testUser1, this should still allow them to edit the policy
	// despite not not being an organization admin
	anotherViewerUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser1.UserCtx, t, &anotherViewerUser, enums.RoleMember, testUser1.OrganizationID)

	gm2 := (&GroupMemberBuilder{client: suite.client, UserID: anotherViewerUser.ID, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)

	// create one more group that will be used to test the blocked group permissions and add anotherViewerUser to it
	blockGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&GroupMemberBuilder{client: suite.client, UserID: anotherViewerUser.ID, GroupID: blockGroup.ID}).MustNew(testUser1.UserCtx, t)

	// create a view only user and add them to the same organization as testUser1
	meowViewerUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser1.UserCtx, t, &meowViewerUser, enums.RoleMember, testUser1.OrganizationID)

	// create one more group that will be used to test the blocked group permissions and add anotherViewerUser to it
	viewerGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	gm3 := (&GroupMemberBuilder{client: suite.client, UserID: meowViewerUser.ID, GroupID: blockGroup.ID}).MustNew(testUser1.UserCtx, t)

	// add add user to the viewer group
	gm4 := (&GroupMemberBuilder{client: suite.client, UserID: viewOnlyUser.ID, GroupID: viewerGroup.ID}).MustNew(testUser1.UserCtx, t)

	// ensure the user does not currently have access to the program
	_, err = suite.client.api.GetProgramByID(viewOnlyUser.UserCtx, program.ID)
	assert.ErrorContains(t, err, notFoundErrorMsg)

	testCases := []struct {
		name              string
		programID         string
		request           testclient.UpdateProgramInput
		client            *testclient.TestClient
		ctx               context.Context
		expectedErr       string
		expectedEdgeCount int
	}{
		{
			name:      "happy path, update field",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				Description:  lo.ToPtr("new description"),
				ProgramType:  &enums.ProgramTypeRiskAssessment,
				AddEditorIDs: []string{testUser1.GroupID}, // add the group to the editor groups for the subsequent tests
				AddViewerIDs: []string{viewerGroup.ID},    // add the group to the viewer groups and ensure the user has access to the program
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "happy path, update multiple fields using pat",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				Status:               &enums.ProgramStatusReadyForAuditor,
				ProgramType:          &enums.ProgramTypeFramework,
				FrameworkName:        lo.ToPtr("SOC 2"),
				AuditFirm:            lo.ToPtr("Meow Audit, LLC."),
				Auditor:              lo.ToPtr("Meowz Meow"),
				AuditorEmail:         lo.ToPtr("m@meow-audit.com"),
				EndDate:              lo.ToPtr(time.Now().AddDate(0, 0, 30)),
				AuditorReady:         lo.ToPtr(true),
				AuditorWriteComments: lo.ToPtr(true),
				AuditorReadComments:  lo.ToPtr(true),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:      "remove program member, can remove self if org owner",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				RemoveProgramMembers: []string{testUserProgramMemberID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "add program member, cannot add self",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				AddProgramMembers: []*testclient.AddProgramMembershipInput{
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
			name:      "add program member, can add another user",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				AddProgramMembers: []*testclient.AddProgramMembershipInput{
					{
						UserID: adminUser.ID,
					},
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "happy path, remove program member",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				RemoveProgramMembers: []string{pm.ID},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:      "happy path, re-add program member as editor",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				AddProgramMembers: []*testclient.AddProgramMembershipInput{
					{
						UserID: pm.UserID,
						Role:   &enums.RoleAdmin,
					},
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "happy path, update edge - procedure",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				AddProcedureIDs: []string{procedure1.ID},
			},
			client:            suite.client.api,
			ctx:               testUser1.UserCtx,
			expectedEdgeCount: 1,
		},
		{
			name:      "happy path, update edge - policy",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				AddInternalPolicyIDs: []string{policy1.ID},
			},
			client:            suite.client.api,
			ctx:               testUser1.UserCtx,
			expectedEdgeCount: 1,
		},
		{
			name:      "update edge - procedure - not allowed to access procedure",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				AddProcedureIDs: []string{procedure2.ID},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:      "update edge - policy - not allowed to access procedure",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				AddInternalPolicyIDs: []string{policy2.ID},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:      "update not allowed, not enough permissions",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				Description: lo.ToPtr("newer description"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg, // user in in viewer group, but has no edit access
		},
		{
			name:      "update not allowed, no permissions",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				Description: lo.ToPtr("newer description"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:      "update allowed, user in editor group",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				Description: lo.ToPtr("soc2 2024"),
			},
			client: suite.client.api,
			ctx:    anotherAdminUser.UserCtx, // user assigned to the group which has editor permissions
		},
		{
			name:      "update not allowed, program is archived and status update is archived",
			programID: archivedProgram.ID,
			request: testclient.UpdateProgramInput{
				Description: lo.ToPtr("newer description"),
				Status:      lo.ToPtr(enums.ProgramStatusArchived),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: hooks.ErrArchivedProgramUpdateNotAllowed.Error(),
		},
		{
			name:      "update not allowed, program is archived",
			programID: archivedProgram.ID,
			request: testclient.UpdateProgramInput{
				Description: lo.ToPtr("newer description"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: hooks.ErrArchivedProgramUpdateNotAllowed.Error(),
		},
		{
			name:      "update allowed, program is archived but status is updated",
			programID: archivedProgram.ID,
			request: testclient.UpdateProgramInput{
				Status: lo.ToPtr(enums.ProgramStatusInProgress),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "update allowed, program is not archived but updated to archived state",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				Status: lo.ToPtr(enums.ProgramStatusArchived),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "update allowed, program is archived but updated to in progress state",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				Status: lo.ToPtr(enums.ProgramStatusInProgress),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateProgram(tc.ctx, tc.programID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// add checks for the updated fields if they were set in the request
			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.UpdateProgram.Program.Description))
			}

			if tc.request.Status != nil {
				assert.Check(t, is.Equal(*tc.request.Status, resp.UpdateProgram.Program.Status))
			}

			if tc.request.ProgramType != nil {
				assert.Check(t, is.Equal(*tc.request.ProgramType, resp.UpdateProgram.Program.ProgramType))
			}

			if tc.request.FrameworkName != nil {
				assert.Check(t, is.DeepEqual(tc.request.FrameworkName, resp.UpdateProgram.Program.FrameworkName))
			}

			if tc.request.StartDate != nil {
				assert.Assert(t, resp.UpdateProgram.Program.StartDate != nil)
				diff := resp.UpdateProgram.Program.StartDate.Sub(*tc.request.StartDate)
				assert.Assert(t, diff >= -2*time.Minute && diff <= 2*time.Minute, "time difference is not within 2 minutes")
			}

			if tc.request.EndDate != nil {
				assert.Assert(t, resp.UpdateProgram.Program.EndDate != nil)
				diff := resp.UpdateProgram.Program.EndDate.Sub(*tc.request.EndDate)
				assert.Assert(t, diff >= -2*time.Minute && diff <= 2*time.Minute, "time difference is not within 2 minutes")
			}

			if tc.request.AuditorReady != nil {
				assert.Check(t, is.Equal(*tc.request.AuditorReady, resp.UpdateProgram.Program.AuditorReady))
			}

			if tc.request.AuditorWriteComments != nil {
				assert.Check(t, is.Equal(*tc.request.AuditorWriteComments, resp.UpdateProgram.Program.AuditorWriteComments))
			}

			if tc.request.AuditorReadComments != nil {
				assert.Check(t, is.Equal(*tc.request.AuditorReadComments, resp.UpdateProgram.Program.AuditorReadComments))
			}

			if tc.request.AuditFirm != nil {
				assert.Check(t, is.DeepEqual(tc.request.AuditFirm, resp.UpdateProgram.Program.AuditFirm))
			}

			if tc.request.Auditor != nil {
				assert.Check(t, is.DeepEqual(tc.request.Auditor, resp.UpdateProgram.Program.Auditor))
			}

			if tc.request.AuditorEmail != nil {
				assert.Check(t, is.DeepEqual(tc.request.AuditorEmail, resp.UpdateProgram.Program.AuditorEmail))
			}

			// check edges
			if len(tc.request.AddProcedureIDs) > 0 {
				assert.Assert(t, is.Len(resp.UpdateProgram.Program.Procedures.Edges, 1))
				for _, edge := range resp.UpdateProgram.Program.Procedures.Edges {
					assert.Check(t, is.Equal(procedure1.ID, edge.Node.ID))
				}
			}

			if len(tc.request.AddInternalPolicyIDs) > 0 {
				assert.Assert(t, is.Len(resp.UpdateProgram.Program.InternalPolicies.Edges, 1))
				for _, edge := range resp.UpdateProgram.Program.InternalPolicies.Edges {
					assert.Check(t, is.Equal(policy1.ID, edge.Node.ID))
				}
			}

			if len(tc.request.AddEditorIDs) > 0 {
				assert.Assert(t, is.Len(resp.UpdateProgram.Program.Editors.Edges, 1))
				for _, edge := range resp.UpdateProgram.Program.Editors.Edges {
					assert.Check(t, is.Equal(testUser1.GroupID, edge.Node.ID))
				}
			}

			if len(tc.request.AddBlockedGroupIDs) > 0 {
				assert.Assert(t, is.Len(resp.UpdateProgram.Program.BlockedGroups, 1))
				for _, edge := range resp.UpdateProgram.Program.BlockedGroups.Edges {
					assert.Check(t, is.Equal(blockGroup.ID, edge.Node.ID))
				}
			}

			if len(tc.request.AddViewerIDs) > 0 {
				assert.Assert(t, is.Len(resp.UpdateProgram.Program.Viewers.Edges, 1))
				for _, edge := range resp.UpdateProgram.Program.Viewers.Edges {
					assert.Check(t, is.Equal(viewerGroup.ID, edge.Node.ID))
				}

				// ensure the user has access to the program now
				res, err := suite.client.api.GetProgramByID(viewOnlyUser.UserCtx, program.ID)
				assert.NilError(t, err)
				assert.Assert(t, res != nil)
				assert.Check(t, is.Equal(program.ID, res.Program.ID))
			}

			if len(tc.request.AddProgramMembers) > 0 {
				assert.Assert(t, is.Len(resp.UpdateProgram.Program.Members.Edges, 2))

				programUserFound := false
				adminUserFound := false
				for _, edge := range resp.UpdateProgram.Program.Members.Edges {
					if edge.Node.User.ID == programUser.ID {
						programUserFound = true
					} else if edge.Node.User.ID == adminUser.ID {
						adminUserFound = true
					}
				}
				// here originally, and then later re-added as an admin
				assert.Check(t, programUserFound, "program user not found in program members")
				assert.Check(t, adminUserFound, "admin user not found in program members")
			}

			// member was removed, ensure there is only one member left
			if len(tc.request.RemoveProgramMembers) > 0 {
				assert.Assert(t, is.Len(resp.UpdateProgram.Program.Members.Edges, 1))
			}
		})
	}

	// cleanup program
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: program.ID}).MustDelete(testUser1.UserCtx, t)
	// cleanup policy and procedure
	(&Cleanup[*generated.ProcedureDeleteOne]{client: suite.client.db.Procedure, ID: procedure1.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.InternalPolicyDeleteOne]{client: suite.client.db.InternalPolicy, ID: policy1.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ProcedureDeleteOne]{client: suite.client.db.Procedure, ID: procedure2.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.InternalPolicyDeleteOne]{client: suite.client.db.InternalPolicy, ID: policy2.ID}).MustDelete(testUser2.UserCtx, t)
	// cleanup group
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{blockGroup.ID, viewerGroup.ID}}).MustDelete(testUser1.UserCtx, t)
	// org member cleanup
	(&Cleanup[*generated.OrgMembershipDeleteOne]{client: suite.client.db.OrgMembership, IDs: []string{om.ID, gm1.Edges.OrgMembership.ID, gm2.Edges.OrgMembership.ID, gm3.Edges.OrgMembership.ID, gm4.Edges.OrgMembership.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteProgram(t *testing.T) {
	// create Programs to be deleted
	program1 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
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
			name:        "program already deleted, not found",
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
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteProgram.DeletedID))
		})
	}
}
