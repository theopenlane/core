package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
)

func (suite *GraphTestSuite) TestQueryNarrative() {
	t := suite.T()

	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can create a Narrative
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)

	// add test cases for querying the Narrative
	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:   "happy path",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:     "read only user, same org, no access to the program",
			client:   suite.client.api,
			ctx:      viewOnlyUser.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:   "admin user, access to the program",
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "happy path using personal access token",
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:     "narrative not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "narrative not found, using not authorized user",
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			// setup the narrative if it is not already created
			if tc.queryID == "" {
				resp, err := suite.client.api.CreateNarrative(testUser1.UserCtx,
					openlaneclient.CreateNarrativeInput{
						Name:       "Narrative",
						ProgramIDs: []string{program.ID},
					})

				require.NoError(t, err)
				require.NotNil(t, resp)

				tc.queryID = resp.CreateNarrative.Narrative.ID
			}

			resp, err := tc.client.GetNarrativeByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			require.NotEmpty(t, resp.Narrative)

			assert.Equal(t, tc.queryID, resp.Narrative.ID)
			assert.NotEmpty(t, resp.Narrative.Name)

			require.Len(t, resp.Narrative.Programs, 1)
			assert.NotEmpty(t, resp.Narrative.Programs[0].ID)
		})
	}
}

func (suite *GraphTestSuite) TestQueryNarratives() {
	t := suite.T()

	// create multiple objects to be queried using testUser1
	(&NarrativeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&NarrativeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	org := (&OrganizationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	userCtxAnotherOrg := auth.NewTestContextWithOrgID(testUser1.ID, org.ID)

	// add narrative for the user to another org; this should not be returned for JWT auth, since it's
	// restricted to a single org. PAT auth would return it if both orgs are authorized on the token
	(&NarrativeBuilder{client: suite.client}).MustNew(userCtxAnotherOrg, t)

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
			name:            "happy path, using read only user of the same org, no programs or groups associated",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 0,
		},
		{
			name:            "happy path, no access to the program or group",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: 0,
		},
		{
			name:            "happy path, using pat",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no narratives should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllNarratives(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.Narratives.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateNarrative() {
	t := suite.T()

	program1 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	programAnotherUser := (&ProgramBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	// group for the view only user
	groupMember := (&GroupMemberBuilder{client: suite.client, UserID: viewOnlyUser.ID}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can create a narrative associated with the program1
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program1.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)

	// create groups to be associated with the narrative
	blockedGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	viewerGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name          string
		request       openlaneclient.CreateNarrativeInput
		addGroupToOrg bool
		client        *openlaneclient.OpenlaneClient
		ctx           context.Context
		expectedErr   string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateNarrativeInput{
				Name: "Narrative",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateNarrativeInput{
				Name:        "Another Narrative",
				Description: lo.ToPtr("A description of the Narrative"),
				Details:     lo.ToPtr("Details of the Narrative"),
				ProgramIDs:  []string{program1.ID, program2.ID}, // multiple programs
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "add groups",
			request: openlaneclient.CreateNarrativeInput{
				Name:            "Test Procedure",
				EditorIDs:       []string{testUser1.GroupID},
				BlockedGroupIDs: []string{blockedGroup.ID},
				ViewerIDs:       []string{viewerGroup.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: openlaneclient.CreateNarrativeInput{
				Name:    "Narrative",
				OwnerID: &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "using api token",
			request: openlaneclient.CreateNarrativeInput{
				Name: "Narrative",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: openlaneclient.CreateNarrativeInput{
				Name: "Narrative",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user now authorized, added to group with creator permissions",
			request: openlaneclient.CreateNarrativeInput{
				Name: "Narrative",
			},
			addGroupToOrg: true,
			client:        suite.client.api,
			ctx:           viewOnlyUser.UserCtx,
		},
		{
			name: "user authorized, they were added to the program",
			request: openlaneclient.CreateNarrativeInput{
				Name:       "Narrative",
				ProgramIDs: []string{program1.ID},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "user authorized, user not authorized to one of the programs",
			request: openlaneclient.CreateNarrativeInput{
				Name:       "Narrative",
				ProgramIDs: []string{program1.ID, program2.ID},
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "missing required name",
			request:     openlaneclient.CreateNarrativeInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "user not authorized, no permissions to one of the programs",
			request: openlaneclient.CreateNarrativeInput{
				Name:       "Narrative",
				ProgramIDs: []string{programAnotherUser.ID, program1.ID},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.addGroupToOrg {
				_, err := suite.client.api.UpdateOrganization(testUser1.UserCtx, testUser1.OrganizationID,
					openlaneclient.UpdateOrganizationInput{
						AddNarrativeCreatorIDs: []string{groupMember.GroupID},
					}, nil)
				require.NoError(t, err)
			}

			resp, err := tc.client.CreateNarrative(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// check required fields
			require.NotEmpty(t, resp.CreateNarrative.Narrative.ID)
			assert.Equal(t, tc.request.Name, resp.CreateNarrative.Narrative.Name)

			// ensure the program is set
			if len(tc.request.ProgramIDs) > 0 {
				require.NotEmpty(t, resp.CreateNarrative.Narrative.Programs)
				require.Len(t, resp.CreateNarrative.Narrative.Programs, len(tc.request.ProgramIDs))

				for i, p := range resp.CreateNarrative.Narrative.Programs {
					assert.Equal(t, tc.request.ProgramIDs[i], p.ID)
				}
			} else {
				assert.Empty(t, resp.CreateNarrative.Narrative.Programs)
			}

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.CreateNarrative.Narrative.Description)
			} else {
				assert.Empty(t, resp.CreateNarrative.Narrative.Description)
			}

			if tc.request.Details != nil {
				assert.Equal(t, tc.request.Details, resp.CreateNarrative.Narrative.Details)
			} else {
				assert.Empty(t, resp.CreateNarrative.Narrative.Details)
			}

			if len(tc.request.EditorIDs) > 0 {
				require.Len(t, resp.CreateNarrative.Narrative.Editors, 1)
				for _, edge := range resp.CreateNarrative.Narrative.Editors {
					assert.Equal(t, testUser1.GroupID, edge.ID)
				}
			}

			if len(tc.request.BlockedGroupIDs) > 0 {
				require.Len(t, resp.CreateNarrative.Narrative.BlockedGroups, 1)
				for _, edge := range resp.CreateNarrative.Narrative.BlockedGroups {
					assert.Equal(t, blockedGroup.ID, edge.ID)
				}
			}

			if len(tc.request.ViewerIDs) > 0 {
				require.Len(t, resp.CreateNarrative.Narrative.Viewers, 1)
				for _, edge := range resp.CreateNarrative.Narrative.Viewers {
					assert.Equal(t, viewerGroup.ID, edge.ID)
				}
			}

			// ensure the org owner has access to the narrative that was created by an api token
			if tc.client == suite.client.apiWithToken {
				res, err := suite.client.api.GetNarrativeByID(testUser1.UserCtx, resp.CreateNarrative.Narrative.ID)
				require.NoError(t, err)
				require.NotEmpty(t, res)
				assert.Equal(t, resp.CreateNarrative.Narrative.ID, res.Narrative.ID)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateNarrative() {
	t := suite.T()

	program := (&ProgramBuilder{client: suite.client, EditorIDs: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	narrative := (&NarrativeBuilder{client: suite.client, ProgramID: program.ID}).MustNew(testUser1.UserCtx, t)

	// create another admin user and add them to the same organization and group as testUser1
	// this will allow us to test the group editor/viewer permissions
	anotherAdminUser := suite.userBuilder(context.Background())
	suite.addUserToOrganization(testUser1.UserCtx, &anotherAdminUser, enums.RoleAdmin, testUser1.OrganizationID)

	groupMember := (&GroupMemberBuilder{client: suite.client, UserID: anotherAdminUser.ID}).MustNew(testUser1.UserCtx, t)

	// ensure the user does not currently have access to the narrative
	res, err := suite.client.api.GetNarrativeByID(anotherAdminUser.UserCtx, narrative.ID)
	require.Error(t, err)
	require.Nil(t, res)

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateNarrativeInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: openlaneclient.UpdateNarrativeInput{
				Tags:         []string{"tag1", "tag2"},
				Description:  lo.ToPtr("Updated description"),
				AddViewerIDs: []string{groupMember.GroupID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: openlaneclient.UpdateNarrativeInput{
				AppendTags:  []string{"tag3", "tag4"},
				Name:        lo.ToPtr("Updated Name"),
				Description: lo.ToPtr("Updated Description"),
				Details:     lo.ToPtr("Updated Details"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, not permissions in same org",
			request: openlaneclient.UpdateNarrativeInput{
				AppendTags: []string{"tag3"},
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: openlaneclient.UpdateNarrativeInput{
				AppendTags: []string{"tag3"},
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateNarrative(tc.ctx, narrative.ID, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			if tc.request.Name != nil {
				assert.Equal(t, *tc.request.Name, resp.UpdateNarrative.Narrative.Name)
			}

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.UpdateNarrative.Narrative.Description)
			}

			if tc.request.Tags != nil {
				require.Len(t, resp.UpdateNarrative.Narrative.Tags, 2)
				assert.ElementsMatch(t, tc.request.Tags, resp.UpdateNarrative.Narrative.Tags)
			}

			if tc.request.AppendTags != nil {
				assert.Len(t, resp.UpdateNarrative.Narrative.Tags, 4)
				assert.Contains(t, resp.UpdateNarrative.Narrative.Tags, tc.request.AppendTags[0])
				assert.Contains(t, resp.UpdateNarrative.Narrative.Tags, tc.request.AppendTags[1])
			}

			if tc.request.Details != nil {
				assert.Equal(t, tc.request.Details, resp.UpdateNarrative.Narrative.Details)
			}

			if len(tc.request.AddViewerIDs) > 0 {
				require.Len(t, resp.UpdateNarrative.Narrative.Viewers, 1)
				found := false
				for _, edge := range resp.UpdateNarrative.Narrative.Viewers {
					if edge.ID == tc.request.AddViewerIDs[0] {
						found = true
						break
					}
				}

				assert.True(t, found)

				// ensure the user has access to the narrative now
				res, err := suite.client.api.GetNarrativeByID(anotherAdminUser.UserCtx, narrative.ID)
				require.NoError(t, err)
				require.NotEmpty(t, res)
				assert.Equal(t, narrative.ID, res.Narrative.ID)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteNarrative() {
	t := suite.T()

	// create objects to be deleted
	Narrative1 := (&NarrativeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	Narrative2 := (&NarrativeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  Narrative1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: Narrative1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  Narrative1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: Narrative2.ID,
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
			resp, err := tc.client.DeleteNarrative(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteNarrative.DeletedID)
		})
	}
}
