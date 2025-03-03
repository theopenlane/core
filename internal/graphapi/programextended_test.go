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
			assert.Len(t, resp.CreateProgramWithMembers.Program.Members, len(tc.request.Members)+1)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateFullProgram() {
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
		request     openlaneclient.CreateFullProgramInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, all the fields",
			request: openlaneclient.CreateFullProgramInput{
				Program: &openlaneclient.CreateProgramInput{
					Name: "mitb program",
				},
				Members: members,
				Standard: &openlaneclient.CreateStandardInput{
					Name: "mitb standard",
				},
				Controls: []*openlaneclient.CreateControlWithSubcontrolsInput{
					{
						Control: &openlaneclient.CreateControlInput{
							RefCode: "control-1",
						},
						// TODO: (sfunk): fix with controls schema PR, validation is now
						// requiring control ID as input which we don't want to require in
						// this mutation
						// Subcontrols: []*openlaneclient.CreateSubcontrolInput{
						// 	{
						// 		Name: "sc-1",
						// 	},
						// 	{
						// 		Name: "sc-2",
						// 	},
						// },
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
			assert.Len(t, resp.CreateFullProgram.Program.Members, len(tc.request.Members)+1)

			require.NotNil(t, resp.CreateFullProgram.Program.Controls)
			assert.Len(t, resp.CreateFullProgram.Program.Controls, len(tc.request.Controls))

			// assert.NotNil(t, resp.CreateFullProgram.Program.Controls[0].Subcontrols)
			// assert.Equal(t, 2, len(resp.CreateFullProgram.Program.Controls[0].Subcontrols))

			require.NotNil(t, resp.CreateFullProgram.Program.Risks)
			assert.Len(t, resp.CreateFullProgram.Program.Risks, len(tc.request.Risks))

			require.NotNil(t, resp.CreateFullProgram.Program.InternalPolicies)
			assert.Len(t, resp.CreateFullProgram.Program.InternalPolicies, len(tc.request.InternalPolicies))
		})
	}
}
