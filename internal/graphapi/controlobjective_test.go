package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/utils/ulids"
)

func TestQueryControlObjective(t *testing.T) {
	program := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// add adminUser to the program so that they can create a ControlObjective
	(&th.ProgramMemberBuilder{Client: suite.Client, ProgramID: program.ID,
		UserID: th.SharedAdminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(th.SharedTestUser1.UserCtx, t)
	anonymousContext := th.CreateAnonymousTrustCenterContext(ulids.New().String(), th.SharedTestUser1.OrganizationID)

	controlObjectiveIDs := []string{}
	// add test cases for querying the ControlObjective
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
			name:     "control objective not found, invalid ID",
			queryID:  "invalid",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "control objective not found, using not authorized user",
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
			// setup the control objective if it is not already created
			if tc.queryID == "" {
				resp, err := suite.Client.API.CreateControlObjective(th.SharedTestUser1.UserCtx,
					testclient.CreateControlObjectiveInput{
						Name:       "ControlObjective",
						ProgramIDs: []string{program.ID},
					})

				assert.NilError(t, err)
				assert.Assert(t, resp != nil)

				tc.queryID = resp.CreateControlObjective.ControlObjective.ID

				controlObjectiveIDs = append(controlObjectiveIDs, tc.queryID)
			}

			resp, err := tc.client.GetControlObjectiveByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.ControlObjective.ID))
			assert.Check(t, len(resp.ControlObjective.Name) != 0)
		})
	}

	(&th.Cleanup[*generated.ControlObjectiveDeleteOne]{Client: suite.Client.DB.ControlObjective, IDs: controlObjectiveIDs}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, ID: program.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryControlObjectives(t *testing.T) {
	// create multiple objects to be queried using th.SharedTestUser1
	co1 := (&th.ControlObjectiveBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	co2 := (&th.ControlObjectiveBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	userAnotherOrg := suite.UserBuilder(context.Background(), t)

	// add control objective for the user to another org; this should not be returned for JWT auth, since it's
	// restricted to a single org. PAT auth would return it if both orgs are authorized on the token
	(&th.ControlObjectiveBuilder{Client: suite.Client}).MustNew(userAnotherOrg.UserCtx, t)

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
			name:            "happy path, api token with scopes",
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
			name:            "another user, no control objectives should be returned",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllControlObjectives(tc.ctx)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.ControlObjectives.Edges, tc.expectedResults))
		})
	}

	(&th.Cleanup[*generated.ControlObjectiveDeleteOne]{Client: suite.Client.DB.ControlObjective, IDs: []string{co1.ID, co2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	th.CleanupOrganizationDataWithContext(userAnotherOrg.UserCtx, t)
}

func TestMutationCreateControlObjective(t *testing.T) {
	program1 := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	program2 := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	programAnotherUser := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser2.UserCtx, t)

	// group for the view only user
	groupMember := (&th.GroupMemberBuilder{Client: suite.Client, UserID: th.SharedViewOnlyUser.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// add adminUser to the program so that they can create a control objective associated with the program1
	(&th.ProgramMemberBuilder{Client: suite.Client, ProgramID: program1.ID,
		UserID: th.SharedAdminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(th.SharedTestUser1.UserCtx, t)

	// create groups to be associated with the control objective
	blockedGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	viewerGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name          string
		request       testclient.CreateControlObjectiveInput
		addGroupToOrg bool
		client        *testclient.TestClient
		ctx           context.Context
		expectedErr   string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateControlObjectiveInput{
				Name: "ControlObjective",
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateControlObjectiveInput{
				Name:                 "Another ControlObjective",
				Category:             lo.ToPtr("Category"),
				Subcategory:          lo.ToPtr("Subcategory"),
				DesiredOutcome:       lo.ToPtr("Desired Outcome"),
				Status:               &enums.ObjectiveActiveStatus,
				ControlObjectiveType: lo.ToPtr("operational"),
				Revision:             lo.ToPtr("v1.0.0"),
				ProgramIDs:           []string{program1.ID, program2.ID}, // multiple programs
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "add groups",
			request: testclient.CreateControlObjectiveInput{
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
			request: testclient.CreateControlObjectiveInput{
				Name:    "ControlObjective",
				OwnerID: &th.SharedTestUser1.OrganizationID,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "using api token",
			request: testclient.CreateControlObjectiveInput{
				Name: "ControlObjective",
			},
			client: suite.Client.APIWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateControlObjectiveInput{
				Name: "ControlObjective",
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "user now authorized, added to group with creator permissions",
			request: testclient.CreateControlObjectiveInput{
				Name: "ControlObjective",
			},
			addGroupToOrg: true,
			client:        suite.Client.API,
			ctx:           th.SharedViewOnlyUser.UserCtx,
		},
		{
			name: "user authorized, they were added to the program",
			request: testclient.CreateControlObjectiveInput{
				Name:       "ControlObjective",
				ProgramIDs: []string{program1.ID},
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name: "user not authorized, user not authorized to one of the programs",
			request: testclient.CreateControlObjectiveInput{
				Name:       "ControlObjective",
				ProgramIDs: []string{program1.ID, program2.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedAdminUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:        "missing required name",
			request:     testclient.CreateControlObjectiveInput{},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "user not authorized, no permissions to one of the programs",
			request: testclient.CreateControlObjectiveInput{
				Name:       "ControlObjective",
				ProgramIDs: []string{programAnotherUser.ID, program1.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.addGroupToOrg {
				_, err := suite.Client.API.UpdateOrganization(th.SharedTestUser1.UserCtx, th.SharedTestUser1.OrganizationID,
					testclient.UpdateOrganizationInput{
						AddControlObjectiveCreatorIDs: []string{groupMember.GroupID},
					}, nil, nil)
				assert.NilError(t, err)
			}

			resp, err := tc.client.CreateControlObjective(tc.ctx, tc.request)
			if tc.expectedErr != "" {

				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// check required fields
			assert.Assert(t, len(resp.CreateControlObjective.ControlObjective.ID) != 0)
			assert.Check(t, is.Equal(tc.request.Name, resp.CreateControlObjective.ControlObjective.Name))

			assert.Check(t, len(resp.CreateControlObjective.ControlObjective.DisplayID) != 0)
			assert.Check(t, is.Contains(resp.CreateControlObjective.ControlObjective.DisplayID, "CLO-"))

			if tc.request.DesiredOutcome != nil {
				assert.Check(t, is.Equal(*tc.request.DesiredOutcome, *resp.CreateControlObjective.ControlObjective.DesiredOutcome))
			} else {
				assert.Check(t, is.Equal(*resp.CreateControlObjective.ControlObjective.DesiredOutcome, ""))
			}

			if tc.request.Status != nil {
				assert.Check(t, is.Equal(*tc.request.Status, *resp.CreateControlObjective.ControlObjective.Status))
			}

			if tc.request.Category != nil {
				assert.Check(t, is.Equal(*tc.request.Category, *resp.CreateControlObjective.ControlObjective.Category))
			} else {
				assert.Check(t, is.Equal(*resp.CreateControlObjective.ControlObjective.Category, ""))
			}

			if tc.request.Subcategory != nil {
				assert.Check(t, is.Equal(*tc.request.Subcategory, *resp.CreateControlObjective.ControlObjective.Subcategory))
			} else {
				assert.Check(t, is.Equal(*resp.CreateControlObjective.ControlObjective.Subcategory, ""))
			}

			if tc.request.ControlObjectiveType != nil {
				assert.Check(t, is.Equal(*tc.request.ControlObjectiveType, *resp.CreateControlObjective.ControlObjective.ControlObjectiveType))
			} else {
				assert.Check(t, is.Equal(*resp.CreateControlObjective.ControlObjective.ControlObjectiveType, ""))
			}

			if tc.request.Revision != nil {
				assert.Check(t, is.Equal(*tc.request.Revision, *resp.CreateControlObjective.ControlObjective.Revision))
			} else {
				assert.Check(t, is.Equal(models.DefaultRevision, *resp.CreateControlObjective.ControlObjective.Revision))
			}

			if tc.request.Source != nil {
				assert.Check(t, is.Equal(*tc.request.Source, *resp.CreateControlObjective.ControlObjective.Source))
			} else {
				assert.Check(t, is.Equal(enums.ControlSourceUserDefined, *resp.CreateControlObjective.ControlObjective.Source))
			}

			// ensure the org owner has access to the control objective that was created by an api token
			if tc.client == suite.Client.APIWithToken {
				res, err := suite.Client.API.GetControlObjectiveByID(th.SharedTestUser1.UserCtx, resp.CreateControlObjective.ControlObjective.ID)
				assert.NilError(t, err)
				assert.Assert(t, res != nil)
				assert.Check(t, is.Equal(resp.CreateControlObjective.ControlObjective.ID, res.ControlObjective.ID))
			}

			(&th.Cleanup[*generated.ControlObjectiveDeleteOne]{Client: suite.Client.DB.ControlObjective, ID: resp.CreateControlObjective.ControlObjective.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		})
	}

	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, IDs: []string{program1.ID, program2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, ID: programAnotherUser.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, IDs: []string{blockedGroup.ID, viewerGroup.ID, groupMember.GroupID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationUpdateControlObjective(t *testing.T) {
	program := (&th.ProgramBuilder{Client: suite.Client, EditorIDs: th.SharedTestUser1.GroupID}).MustNew(th.SharedTestUser1.UserCtx, t)
	controlObjective := (&th.ControlObjectiveBuilder{Client: suite.Client, ProgramID: program.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// create another admin user and add them to the same organization and group as th.SharedTestUser1
	// this will allow us to test the group editor/viewer permissions
	anotherViewUser := suite.UserBuilder(context.Background(), t)
	suite.AddUserToOrganization(th.SharedTestUser1.UserCtx, t, &anotherViewUser, enums.RoleMember, th.SharedTestUser1.OrganizationID)

	groupMember := (&th.GroupMemberBuilder{Client: suite.Client, UserID: anotherViewUser.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// ensure the user does not currently have access to the control objective
	_, err := suite.Client.API.GetControlObjectiveByID(anotherViewUser.UserCtx, controlObjective.ID)
	assert.ErrorContains(t, err, th.NotFoundErrorMsg)

	testCases := []struct {
		name        string
		request     testclient.UpdateControlObjectiveInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: testclient.UpdateControlObjectiveInput{
				DesiredOutcome: lo.ToPtr("Updated outcome"),
				AddViewerIDs:   []string{groupMember.GroupID},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: testclient.UpdateControlObjectiveInput{
				Status:               &enums.ObjectiveActiveStatus,
				Tags:                 []string{"tag1", "tag2"},
				Category:             lo.ToPtr("Category Updated"),
				Subcategory:          lo.ToPtr("Subcategory Updated"),
				ControlObjectiveType: lo.ToPtr("operational"),
				Source:               &enums.ControlSourceUserDefined,
				DesiredOutcome:       lo.ToPtr("Updated outcome again"),
				Revision:             lo.ToPtr("v1.1.0"),
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, revision bump",
			request: testclient.UpdateControlObjectiveInput{
				Status:       &enums.ObjectiveActiveStatus,
				RevisionBump: &models.Major,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "invalid revision",
			request: testclient.UpdateControlObjectiveInput{
				Revision: lo.ToPtr("1.1"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "revision, invalid semver value",
		},
		{
			name: "update not allowed, not permissions in same org",
			request: testclient.UpdateControlObjectiveInput{
				Status: &enums.ObjectiveActiveStatus,
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateControlObjectiveInput{
				DesiredOutcome: lo.ToPtr("update this"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateControlObjective(tc.ctx, controlObjective.ID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Check(t, resp != nil)

			if tc.request.DesiredOutcome != nil {
				assert.Check(t, is.Equal(*tc.request.DesiredOutcome, *resp.UpdateControlObjective.ControlObjective.DesiredOutcome))
			}

			if tc.request.Status != nil {
				assert.Check(t, is.Equal(*tc.request.Status, *resp.UpdateControlObjective.ControlObjective.Status))
			}

			if tc.request.Tags != nil {
				assert.DeepEqual(t, tc.request.Tags, resp.UpdateControlObjective.ControlObjective.Tags)
			}

			if tc.request.Revision != nil {
				assert.Check(t, is.Equal(*tc.request.Revision, *resp.UpdateControlObjective.ControlObjective.Revision))
			}

			if tc.request.RevisionBump == &models.Major {
				assert.Check(t, "v1.0.0" != *resp.UpdateControlObjective.ControlObjective.Revision)
			}

			if tc.request.Category != nil {
				assert.Check(t, is.Equal(*tc.request.Category, *resp.UpdateControlObjective.ControlObjective.Category))
			}

			if tc.request.Subcategory != nil {
				assert.Check(t, is.Equal(*tc.request.Subcategory, *resp.UpdateControlObjective.ControlObjective.Subcategory))
			}

			if tc.request.ControlObjectiveType != nil {
				assert.Check(t, is.Equal(*tc.request.ControlObjectiveType, *resp.UpdateControlObjective.ControlObjective.ControlObjectiveType))
			}

			if tc.request.Source != nil {
				assert.Check(t, is.Equal(*tc.request.Source, *resp.UpdateControlObjective.ControlObjective.Source))
			}

			if len(tc.request.AddViewerIDs) > 0 {
				assert.Check(t, is.Len(resp.UpdateControlObjective.ControlObjective.Viewers.Edges, 1))
				found := false
				for _, edge := range resp.UpdateControlObjective.ControlObjective.Viewers.Edges {
					if edge.Node.ID == tc.request.AddViewerIDs[0] {
						found = true
						break
					}
				}

				assert.Check(t, found)

				// ensure the user has access to the control objective now
				res, err := suite.Client.API.GetControlObjectiveByID(anotherViewUser.UserCtx, controlObjective.ID)
				assert.NilError(t, err)
				assert.Check(t, res != nil)
				assert.Check(t, is.Equal(controlObjective.ID, res.ControlObjective.ID))
			}
		})
	}
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, ID: program.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ControlObjectiveDeleteOne]{Client: suite.Client.DB.ControlObjective, ID: controlObjective.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, ID: groupMember.GroupID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteControlObjective(t *testing.T) {
	// create objects to be deleted
	controlObjective1 := (&th.ControlObjectiveBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	controlObjective2 := (&th.ControlObjectiveBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  controlObjective1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: controlObjective1.ID,
			client:     suite.Client.API,
			ctx:        th.SharedTestUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  controlObjective1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: controlObjective2.ID,
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
			resp, err := tc.client.DeleteControlObjective(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {

				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteControlObjective.DeletedID))
		})
	}
}
