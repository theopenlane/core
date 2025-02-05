package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQueryInternalPolicy() {
	t := suite.T()

	// create an InternalPolicy to be queried using testUser1
	internalPolicy := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add test cases for querying the internal policy
	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: internalPolicy.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: internalPolicy.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: internalPolicy.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "internalPolicy not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "internal policy not found, using not authorized user",
			queryID:  internalPolicy.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetInternalPolicyByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			require.NotEmpty(t, resp.InternalPolicy)

			assert.Equal(t, tc.queryID, resp.InternalPolicy.ID)
			assert.NotEmpty(t, resp.InternalPolicy.Name)
		})
	}
}

func (suite *GraphTestSuite) TestQueryInternalPolicies() {
	t := suite.T()

	// create multiple policies to be queried using testUser1
	(&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using api token",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "happy path, using pat",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no policies should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllInternalPolicies(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.InternalPolicies.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateInternalPolicy() {
	t := suite.T()

	anotherGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// group for the view only user
	groupMember := (&GroupMemberBuilder{client: suite.client, UserID: viewOnlyUser.ID}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name          string
		request       openlaneclient.CreateInternalPolicyInput
		addGroupToOrg bool
		client        *openlaneclient.OpenlaneClient
		ctx           context.Context
		expectedErr   string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateInternalPolicyInput{
				Name: "Test InternalPolicy",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input except edges",
			request: openlaneclient.CreateInternalPolicyInput{
				Name:            "Releasing a new version",
				Description:     lo.ToPtr("instructions on how to release a new version"),
				Status:          lo.ToPtr("draft"),
				PolicyType:      lo.ToPtr("sop"),
				Version:         lo.ToPtr("v1.0"),
				PurposeAndScope: lo.ToPtr("This internal policy is used to release a new version of the software"),
				Background:      lo.ToPtr("The background of the internal policy"),
				Details: map[string]interface{}{
					"details": "do stuff",
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, add editor group",
			request: openlaneclient.CreateInternalPolicyInput{
				Name:      "Test Policy",
				EditorIDs: []string{testUser1.GroupID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "add editor group, again - ensures the same group can be added to multiple policies",
			request: openlaneclient.CreateInternalPolicyInput{
				Name:            "Test Policy",
				EditorIDs:       []string{testUser1.GroupID},
				BlockedGroupIDs: []string{anotherGroup.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: openlaneclient.CreateInternalPolicyInput{
				Name:    "Test Internal Policy",
				OwnerID: &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: openlaneclient.CreateInternalPolicyInput{
				Name: "Test InternalPolicy",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: openlaneclient.CreateInternalPolicyInput{
				Name: "Test InternalPolicy",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user now authorized, added to group with creator permissions",
			request: openlaneclient.CreateInternalPolicyInput{
				Name: "Test InternalPolicy",
			},
			addGroupToOrg: true,
			client:        suite.client.api,
			ctx:           viewOnlyUser.UserCtx,
		},
		{
			name: "missing required field",
			request: openlaneclient.CreateInternalPolicyInput{
				Description: lo.ToPtr("instructions on how to release a new version"),
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
						AddInternalPolicyCreatorIDs: []string{groupMember.GroupID},
					}, nil)
				require.NoError(t, err)
			}

			resp, err := tc.client.CreateInternalPolicy(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// check required fields
			assert.Equal(t, tc.request.Name, resp.CreateInternalPolicy.InternalPolicy.Name)

			assert.NotEmpty(t, resp.CreateInternalPolicy.InternalPolicy.DisplayID)
			assert.Contains(t, resp.CreateInternalPolicy.InternalPolicy.DisplayID, "PLC-")

			// check optional fields with if checks if they were provided or not
			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.CreateInternalPolicy.InternalPolicy.Description)
			} else {
				assert.Empty(t, resp.CreateInternalPolicy.InternalPolicy.Description)
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, *resp.CreateInternalPolicy.InternalPolicy.Status)
			} else {
				assert.Empty(t, resp.CreateInternalPolicy.InternalPolicy.Status)
			}

			if tc.request.PolicyType != nil {
				assert.Equal(t, *tc.request.PolicyType, *resp.CreateInternalPolicy.InternalPolicy.PolicyType)
			} else {
				assert.Empty(t, resp.CreateInternalPolicy.InternalPolicy.PolicyType)
			}

			if tc.request.Version != nil {
				assert.Equal(t, *tc.request.Version, *resp.CreateInternalPolicy.InternalPolicy.Version)
			} else {
				assert.Empty(t, resp.CreateInternalPolicy.InternalPolicy.Version)
			}

			if tc.request.PurposeAndScope != nil {
				assert.Equal(t, *tc.request.PurposeAndScope, *resp.CreateInternalPolicy.InternalPolicy.PurposeAndScope)
			} else {
				assert.Empty(t, resp.CreateInternalPolicy.InternalPolicy.PurposeAndScope)
			}

			if tc.request.Background != nil {
				assert.Equal(t, *tc.request.Background, *resp.CreateInternalPolicy.InternalPolicy.Background)
			} else {
				assert.Empty(t, resp.CreateInternalPolicy.InternalPolicy.Background)
			}

			if tc.request.Details != nil {
				assert.Equal(t, tc.request.Details, resp.CreateInternalPolicy.InternalPolicy.Details)
			} else {
				assert.Empty(t, resp.CreateInternalPolicy.InternalPolicy.Details)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateInternalPolicy() {
	t := suite.T()

	InternalPolicy := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateInternalPolicyInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update name field",
			request: openlaneclient.UpdateInternalPolicyInput{
				Name:         lo.ToPtr("Updated InternalPolicy Name"),
				AddEditorIDs: []string{testUser1.GroupID}, // add the group to the editor groups for subsequent tests
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: openlaneclient.UpdateInternalPolicyInput{
				Status:      lo.ToPtr("published"),
				Description: lo.ToPtr("Updated description"),
				Version:     lo.ToPtr("v2.0"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, not enough permissions",
			request: openlaneclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr("Updated InternalPolicy Name"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update allowed, user in editor group",
			request: openlaneclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr("Updated Procedure Name Again"),
			},
			client: suite.client.api,
			ctx:    anotherAdminUser.UserCtx, // user assigned to the group which has editor permissions
		},
		{
			name: "member update allowed, user in editor group",
			request: openlaneclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr("Updated Procedure Name Again"),
			},
			client: suite.client.api,
			ctx:    anotherViewerUser.UserCtx, // user assigned to the group which has editor permissions
		},
		{
			name: "happy path, block the group from editing",
			request: openlaneclient.UpdateInternalPolicyInput{
				AddBlockedGroupIDs: []string{blockGroup.ID}, // block the group
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "member update no longer allowed, user in blocked group",
			request: openlaneclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr("Updated Procedure Name Again"),
			},
			client:      suite.client.api,
			ctx:         anotherViewerUser.UserCtx, // user assigned to the group which was blocked
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "happy path, remove the group",
			request: openlaneclient.UpdateInternalPolicyInput{
				RemoveEditorIDs: []string{testUser1.GroupID}, // remove the group from the editor groups
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "update not allowed, editor group was removed",
			request: openlaneclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr("Updated Procedure Name Again Again"),
			},
			client:      suite.client.api,
			ctx:         anotherAdminUser.UserCtx, // user assigned to the group which no longer has editor permissions
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: openlaneclient.UpdateInternalPolicyInput{
				Description: lo.ToPtr("Updated description"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateInternalPolicy(tc.ctx, InternalPolicy.ID, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// check updated fields
			if tc.request.Name != nil {
				assert.Equal(t, *tc.request.Name, resp.UpdateInternalPolicy.InternalPolicy.Name)
			}

			if tc.request.Description != nil {
				assert.Equal(t, tc.request.Description, resp.UpdateInternalPolicy.InternalPolicy.Description)
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, *resp.UpdateInternalPolicy.InternalPolicy.Status)
			}

			if tc.request.PolicyType != nil {
				assert.Equal(t, *tc.request.PolicyType, *resp.UpdateInternalPolicy.InternalPolicy.PolicyType)
			}

			if tc.request.Version != nil {
				assert.Equal(t, *tc.request.Version, *resp.UpdateInternalPolicy.InternalPolicy.Version)
			}

			if tc.request.PurposeAndScope != nil {
				assert.Equal(t, *tc.request.PurposeAndScope, *resp.UpdateInternalPolicy.InternalPolicy.PurposeAndScope)
			}

			if tc.request.Background != nil {
				assert.Equal(t, *tc.request.Background, *resp.UpdateInternalPolicy.InternalPolicy.Background)
			}

			if tc.request.Details != nil {
				assert.Equal(t, tc.request.Details, resp.UpdateInternalPolicy.InternalPolicy.Details)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteInternalPolicy() {
	t := suite.T()

	// create internal policies to be deleted
	InternalPolicy1 := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	InternalPolicy2 := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  InternalPolicy1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: InternalPolicy1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  InternalPolicy1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: InternalPolicy2.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteInternalPolicy(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteInternalPolicy.DeletedID)
		})
	}
}
