package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestMutationCreateProgramWithMembers(t *testing.T) {
	// setup a separate user
	t.Parallel()

	localTestOrg := suite.SeedFreshMinimalOrgUsers(t, false)
	user := localTestOrg.Owner

	member := localTestOrg.Member
	admin := localTestOrg.Admin

	members := []*testclient.CreateMemberWithProgramInput{
		{
			UserID: member.ID,
			Role:   &enums.RoleMember,
		},
		{
			UserID: admin.ID,
			Role:   &enums.RoleAdmin,
		},
	}

	publicStandard := (&th.StandardBuilder{Client: suite.Client, IsPublic: true}).MustNew(th.SharedSystemAdminUser.UserCtx, t)

	numAdminControls := 5
	adminControlIDs := []string{}
	for range numAdminControls {
		control := (&th.ControlBuilder{Client: suite.Client, StandardID: publicStandard.ID}).MustNew(th.SharedSystemAdminUser.UserCtx, t)
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
			client: suite.Client.API,
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
			client: suite.Client.API,
			ctx:    user.UserCtx,
		},
		{
			name: "happy path, minimal input, no member should work",
			request: testclient.CreateProgramWithMembersInput{
				Program: &testclient.CreateProgramInput{
					Name: "MITB Assessment - 2025",
				},
			},
			client: suite.Client.API,
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
			client: suite.Client.API,
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

	th.CleanupOrganizationDataWithContext(localTestOrg.Owner.UserCtx, t)
}

func TestMutationCreateFullProgram(t *testing.T) {
	// setup a separate user
	t.Parallel()

	localTestOrg := suite.SeedFreshMinimalOrgUsers(t, false)
	user := localTestOrg.Owner

	member := localTestOrg.Member
	admin := localTestOrg.Admin

	numControls := 5
	controlIDs := []string{}
	for range numControls {
		control := (&th.ControlBuilder{Client: suite.Client}).MustNew(user.UserCtx, t)
		controlIDs = append(controlIDs, control.ID)
	}

	resp, err := suite.Client.API.CreateStandard(user.UserCtx, testclient.CreateStandardInput{
		Name:       "Super Awesome Standard",
		ControlIDs: controlIDs,
	}, nil, nil)
	assert.NilError(t, err)

	orgStandard := resp.CreateStandard.Standard

	publicStandard := (&th.StandardBuilder{Client: suite.Client, IsPublic: true}).MustNew(th.SharedSystemAdminUser.UserCtx, t)

	numAdminControls := 5
	adminControlIDs := []string{}
	for range numAdminControls {
		control := (&th.ControlBuilder{Client: suite.Client, StandardID: publicStandard.ID}).MustNew(th.SharedSystemAdminUser.UserCtx, t)
		adminControlIDs = append(adminControlIDs, control.ID)
	}

	members := []*testclient.CreateMemberWithProgramInput{
		{
			UserID: member.ID,
			Role:   &enums.RoleMember,
		},
		{
			UserID: admin.ID,
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
			client:               suite.Client.API,
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
			client:               suite.Client.API,
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
			client: suite.Client.API,
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
	th.CleanupOrganizationDataWithContext(localTestOrg.Owner.UserCtx, t)
	(&th.Cleanup[*generated.StandardDeleteOne]{Client: suite.Client.DB.Standard, ID: publicStandard.ID}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
}
