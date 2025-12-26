package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/common/enums"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestMutationCreateProgramWithMembers(t *testing.T) {
	t.Parallel()

	// setup a separate user
	user := suite.userBuilder(context.Background(), t)

	member := (&OrgMemberBuilder{client: suite.client}).MustNew(user.UserCtx, t)
	admin := (&OrgMemberBuilder{client: suite.client, Role: enums.RoleAdmin.String()}).MustNew(user.UserCtx, t)

	members := []*testclient.CreateMemberWithProgramInput{
		{
			UserID: member.UserID,
			Role:   &enums.RoleMember,
		},
		{
			UserID: admin.UserID,
			Role:   &enums.RoleAdmin,
		},
	}

	publicStandard := (&StandardBuilder{client: suite.client, IsPublic: true}).MustNew(systemAdminUser.UserCtx, t)

	numAdminControls := 5
	adminControlIDs := []string{}
	for range numAdminControls {
		control := (&ControlBuilder{client: suite.client, StandardID: publicStandard.ID}).MustNew(systemAdminUser.UserCtx, t)
		adminControlIDs = append(adminControlIDs, control.ID)
	}

	testCases := []struct {
		name        string
		request     testclient.CreateProgramWithMembersInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input with standard id",
			request: testclient.CreateProgramWithMembersInput{
				Program: &testclient.CreateProgramInput{
					Name: "mitb program",
				},
				Members:    members,
				StandardID: &publicStandard.ID,
			},
			client: suite.client.api,
			ctx:    user.UserCtx,
		},
		{
			name: "happy path, minimal input",
			request: testclient.CreateProgramWithMembersInput{
				Program: &testclient.CreateProgramInput{
					Name: "mitb program",
				},
				Members: members,
			},
			client: suite.client.api,
			ctx:    user.UserCtx,
		},
		{
			name: "happy path, minimal input, no member should work",
			request: testclient.CreateProgramWithMembersInput{
				Program: &testclient.CreateProgramInput{
					Name: "MITB Assessment - 2025",
				},
			},
			client: suite.client.api,
			ctx:    user.UserCtx,
		},
		{
			name: "happy path, minimal input, nil members should work",
			request: testclient.CreateProgramWithMembersInput{
				Program: &testclient.CreateProgramInput{
					Name: "MITB Assessment - 2025",
				},
				Members: nil,
			},
			client: suite.client.api,
			ctx:    user.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateProgramWithMembers(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// check required fields
			assert.Check(t, is.Equal(tc.request.Program.Name, resp.CreateProgramWithMembers.Program.Name))

			// the creator is automatically added as an admin, and the members are added in addition
			assert.Check(t, is.Len(resp.CreateProgramWithMembers.Program.Members.Edges, len(tc.request.Members)+1))
		})
	}

	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: adminControlIDs}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: publicStandard.ID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationCreateFullProgram(t *testing.T) {
	t.Parallel()

	// setup a separate user
	user := suite.userBuilder(context.Background(), t)

	member := (&OrgMemberBuilder{client: suite.client}).MustNew(user.UserCtx, t)
	admin := (&OrgMemberBuilder{client: suite.client, Role: enums.RoleAdmin.String()}).MustNew(user.UserCtx, t)

	numControls := 5
	controlIDs := []string{}
	for range numControls {
		control := (&ControlBuilder{client: suite.client}).MustNew(user.UserCtx, t)
		controlIDs = append(controlIDs, control.ID)
	}

	resp, err := suite.client.api.CreateStandard(user.UserCtx, testclient.CreateStandardInput{
		Name:       "Super Awesome Standard",
		ControlIDs: controlIDs,
	}, nil)
	assert.NilError(t, err)

	orgStandard := resp.CreateStandard.Standard

	publicStandard := (&StandardBuilder{client: suite.client, IsPublic: true}).MustNew(systemAdminUser.UserCtx, t)

	numAdminControls := 5
	adminControlIDs := []string{}
	for range numAdminControls {
		control := (&ControlBuilder{client: suite.client, StandardID: publicStandard.ID}).MustNew(systemAdminUser.UserCtx, t)
		adminControlIDs = append(adminControlIDs, control.ID)
	}

	members := []*testclient.CreateMemberWithProgramInput{
		{
			UserID: member.UserID,
			Role:   &enums.RoleMember,
		},
		{
			UserID: admin.UserID,
			Role:   &enums.RoleAdmin,
		},
	}

	testCases := []struct {
		name                 string
		request              testclient.CreateFullProgramInput
		client               *testclient.TestClient
		ctx                  context.Context
		expectedControlCount int
		expectedErr          string
	}{
		{
			name: "happy path, system standard id",
			request: testclient.CreateFullProgramInput{
				Program: &testclient.CreateProgramInput{
					Name: "test program",
				},
				Members:    members,
				StandardID: lo.ToPtr(publicStandard.ID),
			},
			client:               suite.client.api,
			ctx:                  user.UserCtx,
			expectedControlCount: numAdminControls,
		},
		{
			name: "happy path, standard id",
			request: testclient.CreateFullProgramInput{
				Program: &testclient.CreateProgramInput{
					Name: "test program",
				},
				Members:    members,
				StandardID: &orgStandard.ID,
			},
			client:               suite.client.api,
			ctx:                  user.UserCtx,
			expectedControlCount: numControls,
		},
		{
			name: "happy path, all the fields",
			request: testclient.CreateFullProgramInput{
				Program: &testclient.CreateProgramInput{
					Name: "mitb program",
				},
				Members: members,
				Controls: []*testclient.CreateControlWithSubcontrolsInput{
					{
						Control: &testclient.CreateControlInput{
							RefCode: "control-1",
						},

						Subcontrols: []*testclient.CreateSubcontrolInput{
							{
								RefCode: "sc-1",
							},
							{
								RefCode: "sc-2",
							},
						},
					},
					{
						Control: &testclient.CreateControlInput{
							RefCode: "control 2",
						},
					},
				},
				Risks: []*testclient.CreateRiskInput{
					{
						Name: "risk 1",
					},
				},
				InternalPolicies: []*testclient.CreateInternalPolicyInput{
					{
						Name: "policy 1",
					},
				},
				Procedures: []*testclient.CreateProcedureInput{
					{
						Name: "procedure 1",
					},
				},
			},
			client: suite.client.api,
			ctx:    user.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateFullProgram(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// check required fields
			assert.Check(t, is.Equal(tc.request.Program.Name, resp.CreateFullProgram.Program.Name))

			// the creator is automatically added as an admin, and the members are added in addition
			assert.Check(t, is.Len(resp.CreateFullProgram.Program.Members.Edges, len(tc.request.Members)+1))

			if tc.request.StandardID == nil {
				assert.Assert(t, resp.CreateFullProgram.Program.Controls.Edges != nil)
				assert.Check(t, is.Len(resp.CreateFullProgram.Program.Controls.Edges, len(tc.request.Controls)))

				assert.Check(t, resp.CreateFullProgram.Program.Controls.Edges[0].Node.Subcontrols.Edges != nil)
				assert.Check(t, is.Equal(2, len(resp.CreateFullProgram.Program.Controls.Edges[0].Node.Subcontrols.Edges)))
			} else {
				assert.Check(t, is.Len(resp.CreateFullProgram.Program.Controls.Edges, tc.expectedControlCount))
			}

			assert.Assert(t, resp.CreateFullProgram.Program.Risks.Edges != nil)
			assert.Check(t, is.Len(resp.CreateFullProgram.Program.Risks.Edges, len(tc.request.Risks)))

			assert.Assert(t, resp.CreateFullProgram.Program.InternalPolicies.Edges != nil)
			assert.Check(t, is.Len(resp.CreateFullProgram.Program.InternalPolicies.Edges, len(tc.request.InternalPolicies)))
		})
	}

	// cleanup seeded input
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: controlIDs}).MustDelete(user.UserCtx, t)
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: orgStandard.ID}).MustDelete(user.UserCtx, t)
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: publicStandard.ID}).MustDelete(systemAdminUser.UserCtx, t)
}
