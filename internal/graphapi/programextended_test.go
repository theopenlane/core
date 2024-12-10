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
