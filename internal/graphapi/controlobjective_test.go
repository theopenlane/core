package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/ulids"
)

func (suite *GraphTestSuite) TestQueryControlObjective() {
	t := suite.T()

	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can create a ControlObjective
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)

	// add test cases for querying the ControlObjective
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
			name:     "control objective not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "control objective not found, using not authorized user",
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			// setup the control objective if it is not already created
			if tc.queryID == "" {
				resp, err := suite.client.api.CreateControlObjective(testUser1.UserCtx,
					openlaneclient.CreateControlObjectiveInput{
						Name:       "ControlObjective",
						ProgramIDs: []string{program.ID},
					})

				require.NoError(t, err)
				require.NotNil(t, resp)

				tc.queryID = resp.CreateControlObjective.ControlObjective.ID
			}

			resp, err := tc.client.GetControlObjectiveByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			require.NotEmpty(t, resp.ControlObjective)

			assert.Equal(t, tc.queryID, resp.ControlObjective.ID)
			assert.NotEmpty(t, resp.ControlObjective.Name)
		})
	}
}

func (suite *GraphTestSuite) TestQueryControlObjectives() {
	t := suite.T()

	// create multiple objects to be queried using testUser1
	(&ControlObjectiveBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&ControlObjectiveBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			name:            "another user, no control objectives should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllControlObjectives(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.ControlObjectives.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateControlObjective() {
	t := suite.T()

	program1 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	programAnotherUser := (&ProgramBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	// group for the view only user
	groupMember := (&GroupMemberBuilder{client: suite.client, UserID: viewOnlyUser.ID}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can create a control objective associated with the program1
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program1.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)

	// create groups to be associated with the control objective
	blockedGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	viewerGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name          string
		request       openlaneclient.CreateControlObjectiveInput
		addGroupToOrg bool
		client        *openlaneclient.OpenlaneClient
		ctx           context.Context
		expectedErr   string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateControlObjectiveInput{
				Name: "ControlObjective",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateControlObjectiveInput{
				Name:                 "Another ControlObjective",
				Category:             lo.ToPtr("Category"),
				Subcategory:          lo.ToPtr("Subcategory"),
				DesiredOutcome:       lo.ToPtr("Desired Outcome"),
				Status:               lo.ToPtr("mitigated"),
				ControlObjectiveType: lo.ToPtr("operational"),
				Revision:             lo.ToPtr("v1.0.0"),
				ProgramIDs:           []string{program1.ID, program2.ID}, // multiple programs
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "add groups",
			request: openlaneclient.CreateControlObjectiveInput{
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
			request: openlaneclient.CreateControlObjectiveInput{
				Name:    "ControlObjective",
				OwnerID: &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "using api token",
			request: openlaneclient.CreateControlObjectiveInput{
				Name: "ControlObjective",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: openlaneclient.CreateControlObjectiveInput{
				Name: "ControlObjective",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user now authorized, added to group with creator permissions",
			request: openlaneclient.CreateControlObjectiveInput{
				Name: "ControlObjective",
			},
			addGroupToOrg: true,
			client:        suite.client.api,
			ctx:           viewOnlyUser.UserCtx,
		},
		{
			name: "user authorized, they were added to the program",
			request: openlaneclient.CreateControlObjectiveInput{
				Name:       "ControlObjective",
				ProgramIDs: []string{program1.ID},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "user not authorized, user not authorized to one of the programs",
			request: openlaneclient.CreateControlObjectiveInput{
				Name:       "ControlObjective",
				ProgramIDs: []string{program1.ID, program2.ID},
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "missing required name",
			request:     openlaneclient.CreateControlObjectiveInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "user not authorized, no permissions to one of the programs",
			request: openlaneclient.CreateControlObjectiveInput{
				Name:       "ControlObjective",
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
						AddControlObjectiveCreatorIDs: []string{groupMember.GroupID},
					}, nil)
				require.NoError(t, err)
			}

			resp, err := tc.client.CreateControlObjective(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// check required fields
			require.NotEmpty(t, resp.CreateControlObjective.ControlObjective.ID)
			assert.Equal(t, tc.request.Name, resp.CreateControlObjective.ControlObjective.Name)

			assert.NotEmpty(t, resp.CreateControlObjective.ControlObjective.DisplayID)
			assert.Contains(t, resp.CreateControlObjective.ControlObjective.DisplayID, "CLO-")

			if tc.request.DesiredOutcome != nil {
				assert.Equal(t, *tc.request.DesiredOutcome, *resp.CreateControlObjective.ControlObjective.DesiredOutcome)
			} else {
				assert.Empty(t, resp.CreateControlObjective.ControlObjective.DesiredOutcome)
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, *resp.CreateControlObjective.ControlObjective.Status)
			} else {
				assert.Empty(t, resp.CreateControlObjective.ControlObjective.Status)
			}

			if tc.request.Category != nil {
				assert.Equal(t, *tc.request.Category, *resp.CreateControlObjective.ControlObjective.Category)
			} else {
				assert.Empty(t, resp.CreateControlObjective.ControlObjective.Category)
			}

			if tc.request.Subcategory != nil {
				assert.Equal(t, *tc.request.Subcategory, *resp.CreateControlObjective.ControlObjective.Subcategory)
			} else {
				assert.Empty(t, resp.CreateControlObjective.ControlObjective.Subcategory)
			}

			if tc.request.ControlObjectiveType != nil {
				assert.Equal(t, *tc.request.ControlObjectiveType, *resp.CreateControlObjective.ControlObjective.ControlObjectiveType)
			} else {
				assert.Empty(t, resp.CreateControlObjective.ControlObjective.ControlObjectiveType)
			}

			if tc.request.Revision != nil {
				assert.Equal(t, *tc.request.Revision, *resp.CreateControlObjective.ControlObjective.Revision)
			} else {
				assert.Equal(t, models.DefaultRevision, *resp.CreateControlObjective.ControlObjective.Revision)
			}

			if tc.request.Source != nil {
				assert.Equal(t, *tc.request.Source, *resp.CreateControlObjective.ControlObjective.Source)
			} else {
				assert.Equal(t, enums.ControlSourceUserDefined, *resp.CreateControlObjective.ControlObjective.Source)
			}

			// ensure the org owner has access to the control objective that was created by an api token
			if tc.client == suite.client.apiWithToken {
				res, err := suite.client.api.GetControlObjectiveByID(testUser1.UserCtx, resp.CreateControlObjective.ControlObjective.ID)
				require.NoError(t, err)
				require.NotEmpty(t, res)
				assert.Equal(t, resp.CreateControlObjective.ControlObjective.ID, res.ControlObjective.ID)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateControlObjective() {
	t := suite.T()

	program := (&ProgramBuilder{client: suite.client, EditorIDs: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	controlObjective := (&ControlObjectiveBuilder{client: suite.client, ProgramID: program.ID}).MustNew(testUser1.UserCtx, t)

	// create another admin user and add them to the same organization and group as testUser1
	// this will allow us to test the group editor/viewer permissions
	anotherAdminUser := suite.userBuilder(context.Background())
	suite.addUserToOrganization(testUser1.UserCtx, &anotherAdminUser, enums.RoleAdmin, testUser1.OrganizationID)

	groupMember := (&GroupMemberBuilder{client: suite.client, UserID: anotherAdminUser.ID}).MustNew(testUser1.UserCtx, t)

	// ensure the user does not currently have access to the control objective
	res, err := suite.client.api.GetControlObjectiveByID(anotherAdminUser.UserCtx, controlObjective.ID)
	require.Error(t, err)
	require.Nil(t, res)

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateControlObjectiveInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: openlaneclient.UpdateControlObjectiveInput{
				DesiredOutcome: lo.ToPtr("Updated outcome"),
				AddViewerIDs:   []string{groupMember.GroupID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: openlaneclient.UpdateControlObjectiveInput{
				Status:               lo.ToPtr("mitigated"),
				Tags:                 []string{"tag1", "tag2"},
				Category:             lo.ToPtr("Category Updated"),
				Subcategory:          lo.ToPtr("Subcategory Updated"),
				ControlObjectiveType: lo.ToPtr("operational"),
				Source:               &enums.ControlSourceUserDefined,
				DesiredOutcome:       lo.ToPtr("Updated outcome again"),
				Revision:             lo.ToPtr("v1.1.0"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, revision bump",
			request: openlaneclient.UpdateControlObjectiveInput{
				Status:       lo.ToPtr("open"),
				RevisionBump: &models.Major,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "invalid revision",
			request: openlaneclient.UpdateControlObjectiveInput{
				Revision: lo.ToPtr("1.1"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "revision, invalid semver value",
		},
		{
			name: "update not allowed, not permissions in same org",
			request: openlaneclient.UpdateControlObjectiveInput{
				Status: lo.ToPtr("testing"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: openlaneclient.UpdateControlObjectiveInput{
				DesiredOutcome: lo.ToPtr("update this"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateControlObjective(tc.ctx, controlObjective.ID, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			if tc.request.DesiredOutcome != nil {
				assert.Equal(t, *tc.request.DesiredOutcome, *resp.UpdateControlObjective.ControlObjective.DesiredOutcome)
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, *resp.UpdateControlObjective.ControlObjective.Status)
			}

			if tc.request.Tags != nil {
				assert.ElementsMatch(t, tc.request.Tags, resp.UpdateControlObjective.ControlObjective.Tags)
			}

			if tc.request.Revision != nil {
				assert.Equal(t, *tc.request.Revision, *resp.UpdateControlObjective.ControlObjective.Revision)
			}

			if tc.request.RevisionBump == &models.Major {
				assert.NotEqual(t, "v1.0.0", *resp.UpdateControlObjective.ControlObjective.Revision)
			}

			if tc.request.Category != nil {
				assert.Equal(t, *tc.request.Category, *resp.UpdateControlObjective.ControlObjective.Category)
			}

			if tc.request.Subcategory != nil {
				assert.Equal(t, *tc.request.Subcategory, *resp.UpdateControlObjective.ControlObjective.Subcategory)
			}

			if tc.request.ControlObjectiveType != nil {
				assert.Equal(t, *tc.request.ControlObjectiveType, *resp.UpdateControlObjective.ControlObjective.ControlObjectiveType)
			}

			if tc.request.Source != nil {
				assert.Equal(t, *tc.request.Source, *resp.UpdateControlObjective.ControlObjective.Source)
			}

			if len(tc.request.AddViewerIDs) > 0 {
				require.Len(t, resp.UpdateControlObjective.ControlObjective.Viewers, 1)
				found := false
				for _, edge := range resp.UpdateControlObjective.ControlObjective.Viewers {
					if edge.ID == tc.request.AddViewerIDs[0] {
						found = true
						break
					}
				}

				assert.True(t, found)

				// ensure the user has access to the control objective now
				res, err := suite.client.api.GetControlObjectiveByID(anotherAdminUser.UserCtx, controlObjective.ID)
				require.NoError(t, err)
				require.NotEmpty(t, res)
				assert.Equal(t, controlObjective.ID, res.ControlObjective.ID)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteControlObjective() {
	t := suite.T()

	// create objects to be deleted
	ControlObjective1 := (&ControlObjectiveBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	ControlObjective2 := (&ControlObjectiveBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  ControlObjective1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: ControlObjective1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  ControlObjective1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: ControlObjective2.ID,
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
			resp, err := tc.client.DeleteControlObjective(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteControlObjective.DeletedID)
		})
	}
}
