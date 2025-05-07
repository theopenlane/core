package graphapi_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestMutationCreateProgramWithMembers() {
	t := suite.T()

	members := []*openlaneclient.CreateMemberWithProgramInput{
		{
			UserID: viewOnlyUser.ID,
			Role:   &enums.RoleMember,
		},
		{
			UserID: adminUser.ID,
			Role:   &enums.RoleAdmin,
		},
	}

	testCases := []struct {
		name        string
		request     openlaneclient.CreateProgramWithMembersInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateProgramWithMembersInput{
				Program: &openlaneclient.CreateProgramInput{
					Name: "mitb program",
				},
				Members: members,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, minimal input, no member should work",
			request: openlaneclient.CreateProgramWithMembersInput{
				Program: &openlaneclient.CreateProgramInput{
					Name: "MITB Assessment - 2025",
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, minimal input, nil members should work",
			request: openlaneclient.CreateProgramWithMembersInput{
				Program: &openlaneclient.CreateProgramInput{
					Name: "MITB Assessment - 2025",
				},
				Members: nil,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateProgramWithMembers(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// check required fields
			assert.Equal(t, tc.request.Program.Name, resp.CreateProgramWithMembers.Program.Name)

			// the creator is automatically added as an admin, and the members are added in addition
			assert.Len(t, resp.CreateProgramWithMembers.Program.Members.Edges, len(tc.request.Members)+1)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateFullProgram() {
	t := suite.T()

	numControls := 5
	controlIDs := []string{}
	for range numControls {
		control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
		controlIDs = append(controlIDs, control.ID)
	}

	resp, err := suite.client.api.CreateStandard(testUser1.UserCtx, openlaneclient.CreateStandardInput{
		Name:       "Super Awesome Standard",
		ControlIDs: controlIDs,
	})
	require.NoError(t, err)

	publicStandard := (&StandardBuilder{client: suite.client, IsPublic: true}).MustNew(systemAdminUser.UserCtx, t)

	numAdminControls := 5
	adminControlIDs := []string{}
	for range numAdminControls {
		control := (&ControlBuilder{client: suite.client, StandardID: publicStandard.ID}).MustNew(systemAdminUser.UserCtx, t)
		adminControlIDs = append(adminControlIDs, control.ID)
	}

	members := []*openlaneclient.CreateMemberWithProgramInput{
		{
			UserID: viewOnlyUser.ID,
			Role:   &enums.RoleMember,
		},
		{
			UserID: adminUser.ID,
			Role:   &enums.RoleAdmin,
		},
	}

	testCases := []struct {
		name                 string
		request              openlaneclient.CreateFullProgramInput
		client               *openlaneclient.OpenlaneClient
		ctx                  context.Context
		expectedControlCount int
		expectedErr          string
	}{
		{
			name: "happy path, system standard id",
			request: openlaneclient.CreateFullProgramInput{
				Program: &openlaneclient.CreateProgramInput{
					Name: "test program",
				},
				Members:    members,
				StandardID: publicStandard.ID,
			},
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
			expectedControlCount: numAdminControls,
		},
		{
			name: "happy path, standard id",
			request: openlaneclient.CreateFullProgramInput{
				Program: &openlaneclient.CreateProgramInput{
					Name: "test program",
				},
				Members:    members,
				StandardID: resp.CreateStandard.Standard.ID,
			},
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
			expectedControlCount: numControls,
		},
		{
			name: "happy path, all the fields",
			request: openlaneclient.CreateFullProgramInput{
				Program: &openlaneclient.CreateProgramInput{
					Name: "mitb program",
				},
				Members: members,
				Controls: []*openlaneclient.CreateControlWithSubcontrolsInput{
					{
						Control: &openlaneclient.CreateControlInput{
							RefCode: "control-1",
						},

						Subcontrols: []*openlaneclient.CreateSubcontrolInput{
							{
								RefCode: "sc-1",
							},
							{
								RefCode: "sc-2",
							},
						},
					},
					{
						Control: &openlaneclient.CreateControlInput{
							RefCode: "control 2",
						},
					},
				},
				Risks: []*openlaneclient.CreateRiskInput{
					{
						Name: "risk 1",
					},
				},
				InternalPolicies: []*openlaneclient.CreateInternalPolicyInput{
					{
						Name: "policy 1",
					},
				},
				Procedures: []*openlaneclient.CreateProcedureInput{
					{
						Name: "procedure 1",
					},
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateFullProgram(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// check required fields
			assert.Equal(t, tc.request.Program.Name, resp.CreateFullProgram.Program.Name)

			// the creator is automatically added as an admin, and the members are added in addition
			assert.Len(t, resp.CreateFullProgram.Program.Members.Edges, len(tc.request.Members)+1)

			if tc.request.StandardID == "" {
				require.NotNil(t, resp.CreateFullProgram.Program.Controls.Edges)
				assert.Len(t, resp.CreateFullProgram.Program.Controls.Edges, len(tc.request.Controls))

				assert.NotNil(t, resp.CreateFullProgram.Program.Controls.Edges[0].Node.Subcontrols)
				assert.Equal(t, 2, len(resp.CreateFullProgram.Program.Controls.Edges[0].Node.Subcontrols.Edges))
			} else {
				assert.Len(t, resp.CreateFullProgram.Program.Controls.Edges, tc.expectedControlCount)
			}

			require.NotNil(t, resp.CreateFullProgram.Program.Risks.Edges)
			assert.Len(t, resp.CreateFullProgram.Program.Risks.Edges, len(tc.request.Risks))

			require.NotNil(t, resp.CreateFullProgram.Program.InternalPolicies.Edges)
			assert.Len(t, resp.CreateFullProgram.Program.InternalPolicies.Edges, len(tc.request.InternalPolicies))
		})
	}
}
