package graphapi_test

import (
	"context"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

const errStartDateLaterThanEndDate = "mutation's start date cannot be later than end date"

func TestQueryProgram(t *testing.T) {
	// create program1 with a linked procedure and policy
	program1 := (&th.ProgramBuilder{Client: suite.Client, WithProcedure: true, WithPolicy: true}).MustNew(th.SharedTestUser1.UserCtx, t)
	program2 := (&th.ProgramBuilder{Client: suite.Client, WithProcedure: true, WithPolicy: true}).MustNew(th.SharedAdminUser.UserCtx, t)

	archivedProgram := (&th.ProgramBuilder{Client: suite.Client, WithProcedure: true, WithPolicy: true, Status: enums.ProgramStatusArchived}).MustNew(th.SharedAdminUser.UserCtx, t)

	anonymousContext := th.CreateAnonymousTrustCenterContext(ulids.New().String(), th.SharedTestUser1.OrganizationID)

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
			client:         suite.Client.API,
			ctx:            th.SharedTestUser1.UserCtx,
			expectedResult: program1,
		},
		{
			name:           "happy path, program created by admin user",
			queryID:        program2.ID,
			client:         suite.Client.API,
			ctx:            th.SharedTestUser1.UserCtx,
			expectedResult: program2,
		},
		{
			name:           "happy path using personal access token",
			queryID:        program1.ID,
			client:         suite.Client.APIWithPAT,
			ctx:            context.Background(),
			expectedResult: program1,
		},
		{
			name:           "archived program - happy path using personal access token",
			queryID:        archivedProgram.ID,
			client:         suite.Client.APIWithPAT,
			ctx:            context.Background(),
			expectedResult: archivedProgram,
		},
		{
			name:     "no access, user of same org",
			queryID:  program1.ID,
			client:   suite.Client.API,
			ctx:      th.SharedViewOnlyUser.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "no access, user of different org",
			queryID:  program1.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "no access, anonymous user",
			client:   suite.Client.API,
			ctx:      anonymousContext,
			queryID:  program1.ID,
			errorMsg: th.NotFoundErrorMsg,
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
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, IDs: []string{program1.ID, program2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	// cleanup procedure and policy
	procedureIDs := []string{}
	for _, p := range program1.Edges.Procedures {
		procedureIDs = append(procedureIDs, p.ID)
	}
	policyIDs := []string{}
	for _, p := range program1.Edges.InternalPolicies {
		policyIDs = append(policyIDs, p.ID)
	}

	(&th.Cleanup[*generated.ProcedureDeleteOne]{Client: suite.Client.DB.Procedure, IDs: procedureIDs}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, IDs: policyIDs}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryPrograms(t *testing.T) {
	// programs for the first organization with a linked procedure and policy
	program1 := (&th.ProgramBuilder{Client: suite.Client, WithProcedure: true, WithPolicy: true}).MustNew(th.SharedTestUser1.UserCtx, t)
	program2 := (&th.ProgramBuilder{Client: suite.Client, WithProcedure: true, WithPolicy: true}).MustNew(th.SharedTestUser1.UserCtx, t)

	// program created by an admin user of the first organization with a linked procedure and policy
	program3 := (&th.ProgramBuilder{Client: suite.Client, WithProcedure: true, WithPolicy: true}).MustNew(th.SharedAdminUser.UserCtx, t)

	// archived program for the first organization
	archivedProgram := (&th.ProgramBuilder{Client: suite.Client, WithProcedure: true, WithPolicy: true, Status: enums.ProgramStatusArchived}).MustNew(th.SharedTestUser1.UserCtx, t)

	// program for the other organization with a linked procedure and policy
	anotherUser := suite.UserBuilder(context.Background(), t)
	program4 := (&th.ProgramBuilder{Client: suite.Client, WithProcedure: true, WithPolicy: true}).MustNew(anotherUser.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
		errorMsg        string
	}{
		{
			name:            "happy path, org owner should see all programs",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser1.UserCtx,
			expectedResults: 3, // archived programs not listed by default
		},
		{
			name:            "happy path using personal access token",
			client:          suite.Client.APIWithPAT,
			ctx:             context.Background(),
			expectedResults: 3, // archived programs not listed by default
		},
		{
			name:            "view only user has not been added to any programs",
			client:          suite.Client.API,
			ctx:             th.SharedViewOnlyUser.UserCtx,
			expectedResults: 0,
		},
		{
			name:            "super admin should see all programs in the org",
			client:          suite.Client.API,
			ctx:             th.SharedSuperAdminUser.UserCtx,
			expectedResults: 3, // archived programs not listed by default
		},
		{
			name:            "admin user should see the program they created",
			client:          suite.Client.API,
			ctx:             th.SharedAdminUser.UserCtx,
			expectedResults: 1,
		},
		{
			name:            "owner of the other organization should see the program they created",
			client:          suite.Client.API,
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
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, IDs: []string{program1.ID, program2.ID, program3.ID, archivedProgram.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, ID: program4.ID}).MustDelete(anotherUser.UserCtx, t)

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

	(&th.Cleanup[*generated.ProcedureDeleteOne]{Client: suite.Client.DB.Procedure, IDs: procedureIDs}).MustDelete(th.SharedTestUser1.UserCtx, t)

	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, IDs: policyIDs}).MustDelete(th.SharedTestUser1.UserCtx, t)

	// we can ignore the cleanup for the new user, it won't conflict with other tests
}

func TestMutationCreateProgram(t *testing.T) {
	startDate := time.Now().AddDate(0, 0, 1)
	endDate := time.Now().AddDate(0, 0, 360)

	groupMember := (&th.GroupMemberBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	groupMemberUserCtx := auth.NewTestContextWithOrgID(groupMember.UserID, th.SharedTestUser1.OrganizationID)

	// Create some edge objects
	procedure := (&th.ProcedureBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	policy := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	blockedGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	viewerGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// group that the user does not have access to (for testing permissions)
	anotherGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser2.UserCtx, t)

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
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, all basic input",
			request: testclient.CreateProgramInput{
				Name:                 "mitb program",
				Description:          lo.ToPtr("being the best"),
				FrameworkName:        lo.ToPtr("SOC 2"),
				ProgramOwnerID:       &th.SharedTestUser1.ID,
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
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, edges",
			request: testclient.CreateProgramInput{
				Name:              "mitb program",
				ProcedureIDs:      []string{procedure.ID},
				InternalPolicyIDs: []string{policy.ID},
				ProgramOwnerID:    &th.SharedAdminUser.ID,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "add editor group",
			request: testclient.CreateProgramInput{
				Name:            "Test Program MITB",
				EditorIDs:       []string{th.SharedTestUser1.GroupID},
				BlockedGroupIDs: []string{blockedGroup.ID},
				ViewerIDs:       []string{viewerGroup.ID},
				ProgramOwnerID:  &th.SharedTestUser1.ID,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "add editor group, no access to group",
			request: testclient.CreateProgramInput{
				Name:      "Test Program Meow",
				EditorIDs: []string{anotherGroup.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateProgramInput{
				Name:        "mitb program",
				Description: lo.ToPtr("being the best"),
				OwnerID:     &th.SharedTestUser1.OrganizationID,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateProgramInput{
				Name:        "mitb program",
				Description: lo.ToPtr("being the best"),
			},
			client: suite.Client.APIWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, date valid",
			request: testclient.CreateProgramInput{
				Name:      "Valid Date Program",
				StartDate: lo.ToPtr(startDate),
				EndDate:   lo.ToPtr(endDate),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "only start date",
			request: testclient.CreateProgramInput{
				Name:      "Start Date Only",
				StartDate: lo.ToPtr(startDate),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "only end date",
			request: testclient.CreateProgramInput{
				Name:    "End Date Only",
				EndDate: lo.ToPtr(endDate),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateProgramInput{
				Name: "mitb program",
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "user now authorized, added to group with creator permissions",
			request: testclient.CreateProgramInput{
				Name: "mitb program",
			},
			addGroupToOrg: true,
			client:        suite.Client.API,
			ctx:           groupMemberUserCtx,
		},
		{
			name: "missing required field",
			request: testclient.CreateProgramInput{
				Description: lo.ToPtr("soc2 2024"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "invalid auditor email",
			request: testclient.CreateProgramInput{
				Name:         "mitb program",
				AuditorEmail: lo.ToPtr("invalid email"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "validator failed for field",
		},
		{
			name: "invalid date",
			request: testclient.CreateProgramInput{
				Name:      "Invalid Date Program",
				StartDate: lo.ToPtr(time.Now().AddDate(0, 10, 22)),
				EndDate:   lo.ToPtr(time.Now().AddDate(0, 9, 17)),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: errStartDateLaterThanEndDate,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.addGroupToOrg {
				_, err := suite.Client.API.UpdateOrganization(th.SharedTestUser1.UserCtx, th.SharedTestUser1.OrganizationID,
					testclient.UpdateOrganizationInput{
						AddProgramCreatorIDs: []string{groupMember.GroupID},
					}, nil, nil)
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
			if tc.request.OwnerID != nil && tc.ctx == th.SharedTestUser2.UserCtx {
				assert.Check(t, is.Equal(th.SharedTestUser2.OrganizationID, *resp.CreateProgram.Program.OwnerID))
			}

			// check optional fields
			if tc.request.Description == nil {
				assert.Check(t, is.Len(*resp.CreateProgram.Program.Description, 0))
			} else {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.CreateProgram.Program.Description))
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
					assert.Check(t, is.Equal(th.SharedTestUser1.GroupID, edge.Node.ID))
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
	(&th.Cleanup[*generated.ProcedureDeleteOne]{Client: suite.Client.DB.Procedure, ID: procedure.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, ID: policy.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	// cleanup group
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, IDs: []string{groupMember.GroupID, blockedGroup.ID, viewerGroup.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, ID: anotherGroup.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)

	// cleanup programs
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, IDs: programIDsToCleanup}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationUpdateProgram(t *testing.T) {
	program := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	archivedProgram := (&th.ProgramBuilder{Client: suite.Client, Status: enums.ProgramStatusArchived}).MustNew(th.SharedTestUser1.UserCtx, t)

	// data to test the date validation logic in the update mutation, we want to ensure that the validation is working and that it is possible to update the dates successfully when they are valid
	baseStart := time.Now().AddDate(0, 0, 5)
	baseEnd := time.Now().AddDate(0, 0, 10)

	program.StartDate = baseStart
	program.EndDate = baseEnd

	programMembers, err := suite.Client.API.GetProgramMembersByProgramID(th.SharedTestUser1.UserCtx, &testclient.ProgramMembershipWhereInput{
		ProgramID: &program.ID,
	})
	assert.NilError(t, err)

	testUserProgramMemberID := ""
	for _, pm := range programMembers.ProgramMemberships.Edges {
		if pm.Node.UserID == th.SharedTestUser1.ID {
			testUserProgramMemberID = pm.Node.ID
		}
	}

	// create program user to remove
	programUser := suite.UserBuilder(context.Background(), t)
	om := (&th.OrgMemberBuilder{Client: suite.Client, UserID: programUser.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	pm := (&th.ProgramMemberBuilder{Client: suite.Client, UserID: programUser.ID, ProgramID: program.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// Create some edge objects
	procedure1 := (&th.ProcedureBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	policy1 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// Create some edge objects for another organization
	procedure2 := (&th.ProcedureBuilder{Client: suite.Client}).MustNew(th.SharedTestUser2.UserCtx, t)
	policy2 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser2.UserCtx, t)

	// create another admin user and add them to the same organization and group as th.SharedTestUser1
	// this will allow us to test the group editor permissions
	anotherAdminUser := suite.UserBuilder(context.Background(), t)
	suite.AddUserToOrganization(th.SharedTestUser1.UserCtx, t, &anotherAdminUser, enums.RoleAdmin, th.SharedTestUser1.OrganizationID)

	gm1 := (&th.GroupMemberBuilder{Client: suite.Client, UserID: anotherAdminUser.ID, GroupID: th.SharedTestUser1.GroupID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// create a viewer user and add them to the same organization as th.SharedTestUser1
	// also add them to the same group as th.SharedTestUser1, this should still allow them to edit the policy
	// despite not not being an organization admin
	anotherViewerUser := suite.UserBuilder(context.Background(), t)
	suite.AddUserToOrganization(th.SharedTestUser1.UserCtx, t, &anotherViewerUser, enums.RoleMember, th.SharedTestUser1.OrganizationID)

	gm2 := (&th.GroupMemberBuilder{Client: suite.Client, UserID: anotherViewerUser.ID, GroupID: th.SharedTestUser1.GroupID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// create one more group that will be used to test the blocked group permissions and add anotherViewerUser to it
	blockGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	(&th.GroupMemberBuilder{Client: suite.Client, UserID: anotherViewerUser.ID, GroupID: blockGroup.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// create a view only user and add them to the same organization as th.SharedTestUser1
	meowViewerUser := suite.UserBuilder(context.Background(), t)
	suite.AddUserToOrganization(th.SharedTestUser1.UserCtx, t, &meowViewerUser, enums.RoleMember, th.SharedTestUser1.OrganizationID)

	// create one more group that will be used to test the blocked group permissions and add anotherViewerUser to it
	viewerGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	gm3 := (&th.GroupMemberBuilder{Client: suite.Client, UserID: meowViewerUser.ID, GroupID: blockGroup.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// add add user to the viewer group
	gm4 := (&th.GroupMemberBuilder{Client: suite.Client, UserID: th.SharedViewOnlyUser.ID, GroupID: viewerGroup.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// ensure the user does not currently have access to the program
	_, err = suite.Client.API.GetProgramByID(th.SharedViewOnlyUser.UserCtx, program.ID)
	assert.ErrorContains(t, err, th.NotFoundErrorMsg)

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
				AddEditorIDs: []string{th.SharedTestUser1.GroupID}, // add the group to the editor groups for the subsequent tests
				AddViewerIDs: []string{viewerGroup.ID},             // add the group to the viewer groups and ensure the user has access to the program
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name:      "happy path, update multiple fields using pat",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				Status:               &enums.ProgramStatusReadyForAuditor,
				FrameworkName:        lo.ToPtr("SOC 2"),
				AuditFirm:            lo.ToPtr("Meow Audit, LLC."),
				Auditor:              lo.ToPtr("Meowz Meow"),
				AuditorEmail:         lo.ToPtr("m@meow-audit.com"),
				EndDate:              lo.ToPtr(time.Now().AddDate(0, 0, 30)),
				AuditorReady:         lo.ToPtr(true),
				AuditorWriteComments: lo.ToPtr(true),
				AuditorReadComments:  lo.ToPtr(true),
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name:      "remove program member, can remove self if org owner",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				RemoveProgramMembers: []string{testUserProgramMemberID},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name:      "add program member, cannot add self",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				AddProgramMembers: []*testclient.AddProgramMembershipInput{
					{
						UserID: th.SharedAdminUser.ID,
					},
				},
			},
			client:      suite.Client.API,
			ctx:         th.SharedAdminUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:      "add program member, can add another user",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				AddProgramMembers: []*testclient.AddProgramMembershipInput{
					{
						UserID: th.SharedAdminUser.ID,
					},
				},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name:      "happy path, remove program member",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				RemoveProgramMembers: []string{pm.ID},
			},
			client: suite.Client.APIWithPAT,
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
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name:      "happy path, update edge - procedure",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				AddProcedureIDs: []string{procedure1.ID},
			},
			client:            suite.Client.API,
			ctx:               th.SharedTestUser1.UserCtx,
			expectedEdgeCount: 1,
		},
		{
			name:      "happy path, update edge - policy",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				AddInternalPolicyIDs: []string{policy1.ID},
			},
			client:            suite.Client.API,
			ctx:               th.SharedTestUser1.UserCtx,
			expectedEdgeCount: 1,
		},
		{
			name:      "happy path, valid start and end date update",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				StartDate: lo.ToPtr(baseStart.AddDate(0, 0, -1)),
				EndDate:   lo.ToPtr(baseEnd.AddDate(0, 0, 1)),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name:      "happy path, valid start date update",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				StartDate: lo.ToPtr(time.Now().AddDate(0, 0, 2)),
				EndDate:   lo.ToPtr(baseEnd),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name:      "happy path, valid end date update",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				StartDate: lo.ToPtr(baseStart),
				EndDate:   lo.ToPtr(time.Now().AddDate(1, 2, 5)),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name:      "invalid start and end date update",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				StartDate: lo.ToPtr(time.Now().AddDate(2, 2, 5)),
				EndDate:   lo.ToPtr(time.Now().AddDate(1, 2, 5)),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: errStartDateLaterThanEndDate,
		},
		{
			name:      "invalid start update",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				StartDate: lo.ToPtr(time.Now().AddDate(0, 0, 15)),
				EndDate:   lo.ToPtr(baseEnd),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: errStartDateLaterThanEndDate,
		},
		{
			name:      "invalid end update",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				StartDate: lo.ToPtr(baseStart),
				EndDate:   lo.ToPtr(time.Now().AddDate(0, 0, -15)),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: errStartDateLaterThanEndDate,
		},
		{
			name:      "update edge - procedure - not allowed to access procedure",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				AddProcedureIDs: []string{procedure2.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:      "update edge - policy - not allowed to access procedure",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				AddInternalPolicyIDs: []string{policy2.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:      "update not allowed, not enough permissions",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				Description: lo.ToPtr("newer description"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg, // user in in viewer group, but has no edit access
		},
		{
			name:      "update not allowed, no permissions",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				Description: lo.ToPtr("newer description"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:      "update allowed, user in editor group",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				Description: lo.ToPtr("soc2 2024"),
			},
			client: suite.Client.API,
			ctx:    anotherAdminUser.UserCtx, // user assigned to the group which has editor permissions
		},
		{
			name:      "update not allowed, program is archived and status update is archived",
			programID: archivedProgram.ID,
			request: testclient.UpdateProgramInput{
				Description: lo.ToPtr("newer description"),
				Status:      lo.ToPtr(enums.ProgramStatusArchived),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: hooks.ErrArchivedProgramUpdateNotAllowed.Error(),
		},
		{
			name:      "update not allowed, program is archived",
			programID: archivedProgram.ID,
			request: testclient.UpdateProgramInput{
				Description: lo.ToPtr("newer description"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: hooks.ErrArchivedProgramUpdateNotAllowed.Error(),
		},
		{
			name:      "update allowed, program is archived but status is updated",
			programID: archivedProgram.ID,
			request: testclient.UpdateProgramInput{
				Status: lo.ToPtr(enums.ProgramStatusInProgress),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name:      "update allowed, program is not archived but updated to archived state",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				Status: lo.ToPtr(enums.ProgramStatusArchived),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name:      "update allowed, program is archived but updated to in progress state",
			programID: program.ID,
			request: testclient.UpdateProgramInput{
				Status: lo.ToPtr(enums.ProgramStatusInProgress),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
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
					assert.Check(t, is.Equal(th.SharedTestUser1.GroupID, edge.Node.ID))
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
				res, err := suite.Client.API.GetProgramByID(th.SharedViewOnlyUser.UserCtx, program.ID)
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
					} else if edge.Node.User.ID == th.SharedAdminUser.ID {
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
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, ID: program.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	// cleanup policy and procedure
	(&th.Cleanup[*generated.ProcedureDeleteOne]{Client: suite.Client.DB.Procedure, ID: procedure1.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, ID: policy1.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ProcedureDeleteOne]{Client: suite.Client.DB.Procedure, ID: procedure2.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, ID: policy2.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
	// cleanup group
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, IDs: []string{blockGroup.ID, viewerGroup.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	// org member cleanup
	(&th.Cleanup[*generated.OrgMembershipDeleteOne]{Client: suite.Client.DB.OrgMembership, IDs: []string{om.ID, gm1.Edges.OrgMembership.ID, gm2.Edges.OrgMembership.ID, gm3.Edges.OrgMembership.ID, gm4.Edges.OrgMembership.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteProgram(t *testing.T) {
	// create Programs to be deleted
	program1 := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	program2 := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:       "happy path, delete program",
			idToDelete: program1.ID,
			client:     suite.Client.API,
			ctx:        th.SharedTestUser1.UserCtx,
		},
		{
			name:        "program already deleted, not found",
			idToDelete:  program1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete program using personal access token",
			idToDelete: program2.ID,
			client:     suite.Client.APIWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown program, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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
