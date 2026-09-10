package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/utils/ulids"
)

func TestQueryNarrative(t *testing.T) {
	program := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// add adminUser to the program so that they can create a Narrative
	(&th.ProgramMemberBuilder{Client: suite.Client, ProgramID: program.ID,
		UserID: th.SharedAdminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(th.SharedTestUser1.UserCtx, t)
	anonymousContext := th.CreateAnonymousTrustCenterContext(ulids.New().String(), th.SharedTestUser1.OrganizationID)

	narratives := []string{}

	// add test cases for querying the Narrative
	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:   "happy path",
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name:     "read only user, same org, no access to the program",
			client:   suite.Client.API,
			ctx:      th.SharedViewOnlyUser.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:   "admin user, access to the program",
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name:   "happy path using personal access token",
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name:     "narrative not found, invalid ID",
			queryID:  "invalid",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "narrative not found, using not authorized user",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "no access, anonymous user",
			client:   suite.Client.API,
			ctx:      anonymousContext,
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			// setup the narrative if it is not already created
			if tc.queryID == "" {
				resp, err := suite.Client.API.CreateNarrative(th.SharedTestUser1.UserCtx,
					testclient.CreateNarrativeInput{
						Name:       "Narrative",
						ProgramIDs: []string{program.ID},
					})

				assert.NilError(t, err)
				assert.Assert(t, resp != nil)

				tc.queryID = resp.CreateNarrative.Narrative.ID
				narratives = append(narratives, tc.queryID)

			}

			resp, err := tc.client.GetNarrativeByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.Narrative.ID))
			assert.Check(t, len(resp.Narrative.Name) != 0)

			assert.Check(t, is.Len(resp.Narrative.Programs.Edges, 1))
			assert.Check(t, len(resp.Narrative.Programs.Edges[0].Node.ID) != 0)

		})
	}

	// delete created narratives
	(&th.Cleanup[*generated.NarrativeDeleteOne]{Client: suite.Client.DB.Narrative, IDs: narratives}).MustDelete(th.SharedTestUser1.UserCtx, t)
	// delete created program
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, ID: program.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryNarratives(t *testing.T) {
	// create multiple objects to be queried using th.SharedTestUser1
	nrt1 := (&th.NarrativeBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	nrt2 := (&th.NarrativeBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	userAnotherOrg := suite.UserBuilder(context.Background(), t)

	// add narrative for the user to another org; this should not be returned for JWT auth, since it's
	// restricted to a single org. PAT auth would return it if both orgs are authorized on the token
	(&th.NarrativeBuilder{Client: suite.Client}).MustNew(userAnotherOrg.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser1.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using read only user of the same org, no programs or groups associated",
			client:          suite.Client.API,
			ctx:             th.SharedViewOnlyUser.UserCtx,
			expectedResults: 0,
		},
		{
			name:            "happy path, scope access to all narratives in org",
			client:          suite.Client.APIWithToken,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "happy path, using pat",
			client:          suite.Client.APIWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no narratives should be returned",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllNarratives(tc.ctx)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Narratives.Edges, tc.expectedResults), "expected %d narratives, got %d", tc.expectedResults, len(resp.Narratives.Edges))
		})
	}

	// delete created narrative
	(&th.Cleanup[*generated.NarrativeDeleteOne]{Client: suite.Client.DB.Narrative, IDs: []string{nrt1.ID, nrt2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	th.CleanupOrganizationDataWithContext(userAnotherOrg.UserCtx, t)
}
func TestMutationCreateNarrative(t *testing.T) {
	program1 := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	program2 := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	programAnotherUser := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser2.UserCtx, t)

	// group for the view only user
	groupMember := (&th.GroupMemberBuilder{Client: suite.Client, UserID: th.SharedViewOnlyUser.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// add adminUser to the program so that they can create a narrative associated with the program1
	(&th.ProgramMemberBuilder{Client: suite.Client, ProgramID: program1.ID,
		UserID: th.SharedAdminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(th.SharedTestUser1.UserCtx, t)

	// create groups to be associated with the narrative
	blockedGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	viewerGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	narratives := []string{}

	testCases := []struct {
		name          string
		request       testclient.CreateNarrativeInput
		addGroupToOrg bool
		client        *testclient.TestClient
		ctx           context.Context
		expectedErr   string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateNarrativeInput{
				Name: "Narrative",
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateNarrativeInput{
				Name:        "Another Narrative",
				Description: lo.ToPtr("A description of the Narrative"),
				Details:     lo.ToPtr("Details of the Narrative"),
				ProgramIDs:  []string{program1.ID, program2.ID}, // multiple programs
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "add groups",
			request: testclient.CreateNarrativeInput{
				Name:            "Test Procedure",
				EditorIDs:       []string{th.SharedTestUser1.GroupID},
				BlockedGroupIDs: []string{blockedGroup.ID},
				ViewerIDs:       []string{viewerGroup.ID},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateNarrativeInput{
				Name:    "Narrative",
				OwnerID: &th.SharedTestUser1.OrganizationID,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "using api token",
			request: testclient.CreateNarrativeInput{
				Name: "Narrative",
			},
			client: suite.Client.APIWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateNarrativeInput{
				Name: "Narrative",
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "user now authorized, added to group with creator permissions",
			request: testclient.CreateNarrativeInput{
				Name: "Narrative",
			},
			addGroupToOrg: true,
			client:        suite.Client.API,
			ctx:           th.SharedViewOnlyUser.UserCtx,
		},
		{
			name: "user authorized, they were added to the program",
			request: testclient.CreateNarrativeInput{
				Name:       "Narrative",
				ProgramIDs: []string{program1.ID},
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name: "user authorized, user not authorized to one of the programs",
			request: testclient.CreateNarrativeInput{
				Name:       "Narrative",
				ProgramIDs: []string{program1.ID, program2.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedAdminUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:        "missing required name",
			request:     testclient.CreateNarrativeInput{},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "user not authorized, no permissions to one of the programs",
			request: testclient.CreateNarrativeInput{
				Name:       "Narrative",
				ProgramIDs: []string{programAnotherUser.ID, program1.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedAdminUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.addGroupToOrg {
				_, err := suite.Client.API.UpdateOrganization(th.SharedTestUser1.UserCtx, th.SharedTestUser1.OrganizationID,
					testclient.UpdateOrganizationInput{
						AddNarrativeCreatorIDs: []string{groupMember.GroupID},
					}, nil, nil)
				assert.NilError(t, err)
			}

			resp, err := tc.client.CreateNarrative(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// check required fields
			assert.Check(t, len(resp.CreateNarrative.Narrative.ID) != 0)
			assert.Check(t, is.Equal(tc.request.Name, resp.CreateNarrative.Narrative.Name))

			// ensure the program is set
			if len(tc.request.ProgramIDs) > 0 {
				assert.Check(t, is.Len(resp.CreateNarrative.Narrative.Programs.Edges, len(tc.request.ProgramIDs)))

				for i, p := range resp.CreateNarrative.Narrative.Programs.Edges {
					assert.Check(t, is.Equal(tc.request.ProgramIDs[i], p.Node.ID))
				}
			} else {
				assert.Check(t, is.Len(resp.CreateNarrative.Narrative.Programs.Edges, 0))
			}

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.CreateNarrative.Narrative.Description))
			} else {
				assert.Check(t, is.Equal(*resp.CreateNarrative.Narrative.Description, ""))
			}

			if tc.request.Details != nil {
				assert.Check(t, is.DeepEqual(tc.request.Details, resp.CreateNarrative.Narrative.Details))
			} else {
				assert.Check(t, is.Equal(*resp.CreateNarrative.Narrative.Details, ""))
			}

			if len(tc.request.EditorIDs) > 0 {
				assert.Check(t, is.Len(resp.CreateNarrative.Narrative.Editors.Edges, 1))
				for _, edge := range resp.CreateNarrative.Narrative.Editors.Edges {
					assert.Check(t, is.Equal(th.SharedTestUser1.GroupID, edge.Node.ID))
				}
			}

			if len(tc.request.BlockedGroupIDs) > 0 {
				assert.Check(t, is.Len(resp.CreateNarrative.Narrative.BlockedGroups.Edges, 1))
				for _, edge := range resp.CreateNarrative.Narrative.BlockedGroups.Edges {
					assert.Check(t, is.Equal(blockedGroup.ID, edge.Node.ID))
				}
			}

			if len(tc.request.ViewerIDs) > 0 {
				assert.Check(t, is.Len(resp.CreateNarrative.Narrative.Viewers.Edges, 1))
				for _, edge := range resp.CreateNarrative.Narrative.Viewers.Edges {
					assert.Check(t, is.Equal(viewerGroup.ID, edge.Node.ID))
				}
			}

			// ensure the org owner has access to the narrative that was created by an api token
			if tc.client == suite.Client.APIWithToken {
				res, err := suite.Client.API.GetNarrativeByID(th.SharedTestUser1.UserCtx, resp.CreateNarrative.Narrative.ID)
				assert.NilError(t, err)
				assert.Assert(t, res != nil)
				assert.Check(t, is.Equal(resp.CreateNarrative.Narrative.ID, res.Narrative.ID))
			}

			narratives = append(narratives, resp.CreateNarrative.Narrative.ID)
		})
	}

	// delete created narratives
	(&th.Cleanup[*generated.NarrativeDeleteOne]{Client: suite.Client.DB.Narrative, IDs: narratives}).MustDelete(th.SharedTestUser1.UserCtx, t)
	// delete created programs
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, IDs: []string{program1.ID, program2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, ID: programAnotherUser.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
	// delete created groups
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, IDs: []string{blockedGroup.ID, viewerGroup.ID, groupMember.GroupID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationUpdateNarrative(t *testing.T) {
	program := (&th.ProgramBuilder{Client: suite.Client, EditorIDs: th.SharedTestUser1.GroupID}).MustNew(th.SharedTestUser1.UserCtx, t)
	narrative := (&th.NarrativeBuilder{Client: suite.Client, ProgramID: program.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// create another admin user and add them to the same organization and group as th.SharedTestUser1
	// this will allow us to test the group editor/viewer permissions
	anotherAdminUser := suite.UserBuilder(context.Background(), t)
	suite.AddUserToOrganization(th.SharedTestUser1.UserCtx, t, &anotherAdminUser, enums.RoleAdmin, th.SharedTestUser1.OrganizationID)

	groupMember := (&th.GroupMemberBuilder{Client: suite.Client, UserID: anotherAdminUser.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// ensure the user does not currently have access to the narrative
	_, err := suite.Client.API.GetNarrativeByID(anotherAdminUser.UserCtx, narrative.ID)
	assert.ErrorContains(t, err, th.NotFoundErrorMsg)

	testCases := []struct {
		name        string
		request     testclient.UpdateNarrativeInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: testclient.UpdateNarrativeInput{
				Tags:         []string{"tag1", "tag2"},
				Description:  lo.ToPtr("Updated description"),
				AddViewerIDs: []string{groupMember.GroupID},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: testclient.UpdateNarrativeInput{
				AppendTags:  []string{"tag3", "tag4"},
				Name:        lo.ToPtr("Updated Name"),
				Description: lo.ToPtr("Updated Description"),
				Details:     lo.ToPtr("Updated Details"),
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, not permissions in same org",
			request: testclient.UpdateNarrativeInput{
				AppendTags: []string{"tag3"},
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateNarrativeInput{
				AppendTags: []string{"tag3"},
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateNarrative(tc.ctx, narrative.ID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.Name != nil {
				assert.Check(t, is.Equal(*tc.request.Name, resp.UpdateNarrative.Narrative.Name))
			}

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.UpdateNarrative.Narrative.Description))
			}

			if tc.request.Tags != nil {
				assert.Check(t, is.Len(resp.UpdateNarrative.Narrative.Tags, 2))
				assert.DeepEqual(t, tc.request.Tags, resp.UpdateNarrative.Narrative.Tags)
			}

			if tc.request.AppendTags != nil {
				assert.Check(t, is.Len(resp.UpdateNarrative.Narrative.Tags, 4))
				assert.Check(t, is.Contains(resp.UpdateNarrative.Narrative.Tags, tc.request.AppendTags[0]))
				assert.Check(t, is.Contains(resp.UpdateNarrative.Narrative.Tags, tc.request.AppendTags[1]))
			}

			if tc.request.Details != nil {
				assert.Check(t, is.DeepEqual(tc.request.Details, resp.UpdateNarrative.Narrative.Details))
			}

			if len(tc.request.AddViewerIDs) > 0 {
				assert.Check(t, is.Len(resp.UpdateNarrative.Narrative.Viewers.Edges, 1))
				found := false
				for _, edge := range resp.UpdateNarrative.Narrative.Viewers.Edges {
					if edge.Node.ID == tc.request.AddViewerIDs[0] {
						found = true
						break
					}
				}

				assert.Check(t, found)

				// ensure the user has access to the narrative now
				res, err := suite.Client.API.GetNarrativeByID(anotherAdminUser.UserCtx, narrative.ID)
				assert.NilError(t, err)
				assert.Assert(t, res != nil)
				assert.Check(t, is.Equal(narrative.ID, res.Narrative.ID))
			}
		})
	}

	// delete created narrative
	(&th.Cleanup[*generated.NarrativeDeleteOne]{Client: suite.Client.DB.Narrative, ID: narrative.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	// delete created program
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, ID: program.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	// delete created group
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, ID: groupMember.GroupID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteNarrative(t *testing.T) {
	// create objects to be deleted
	narrative1 := (&th.NarrativeBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	narrative2 := (&th.NarrativeBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  narrative1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: narrative1.ID,
			client:     suite.Client.API,
			ctx:        th.SharedTestUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  narrative1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: narrative2.ID,
			client:     suite.Client.APIWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteNarrative(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteNarrative.DeletedID))
		})
	}
}
