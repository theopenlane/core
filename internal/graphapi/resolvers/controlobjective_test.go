package resolvers_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/utils/ulids"
)

func TestQueryControlObjective(t *testing.T) {
	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can create a ControlObjective
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)
	anonymousContext := createAnonymousTrustCenterContext(ulids.New().String(), testUser1.OrganizationID)

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
		{
			name:     "no access, anonymous user",
			client:   suite.client.api,
			ctx:      anonymousContext,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			// setup the control objective if it is not already created
			if tc.queryID == "" {
				resp, err := suite.client.api.CreateControlObjective(testUser1.UserCtx,
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

	(&Cleanup[*generated.ControlObjectiveDeleteOne]{client: suite.client.db.ControlObjective, IDs: controlObjectiveIDs}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: program.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryControlObjectives(t *testing.T) {
	// create multiple objects to be queried using testUser1
	co1 := (&ControlObjectiveBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	co2 := (&ControlObjectiveBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	userAnotherOrg := suite.userBuilder(context.Background(), t)

	// add control objective for the user to another org; this should not be returned for JWT auth, since it's
	// restricted to a single org. PAT auth would return it if both orgs are authorized on the token
	(&ControlObjectiveBuilder{client: suite.client}).MustNew(userAnotherOrg.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
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
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.ControlObjectives.Edges, tc.expectedResults))
		})
	}

	(&Cleanup[*generated.ControlObjectiveDeleteOne]{client: suite.client.db.ControlObjective, IDs: []string{co1.ID, co2.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateControlObjective(t *testing.T) {
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
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
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
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "add groups",
			request: testclient.CreateControlObjectiveInput{
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
			request: testclient.CreateControlObjectiveInput{
				Name:    "ControlObjective",
				OwnerID: &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "using api token",
			request: testclient.CreateControlObjectiveInput{
				Name: "ControlObjective",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateControlObjectiveInput{
				Name: "ControlObjective",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user now authorized, added to group with creator permissions",
			request: testclient.CreateControlObjectiveInput{
				Name: "ControlObjective",
			},
			addGroupToOrg: true,
			client:        suite.client.api,
			ctx:           viewOnlyUser.UserCtx,
		},
		{
			name: "user authorized, they were added to the program",
			request: testclient.CreateControlObjectiveInput{
				Name:       "ControlObjective",
				ProgramIDs: []string{program1.ID},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "user not authorized, user not authorized to one of the programs",
			request: testclient.CreateControlObjectiveInput{
				Name:       "ControlObjective",
				ProgramIDs: []string{program1.ID, program2.ID},
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "missing required name",
			request:     testclient.CreateControlObjectiveInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "user not authorized, no permissions to one of the programs",
			request: testclient.CreateControlObjectiveInput{
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
					testclient.UpdateOrganizationInput{
						AddControlObjectiveCreatorIDs: []string{groupMember.GroupID},
					}, nil)
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
			if tc.client == suite.client.apiWithToken {
				res, err := suite.client.api.GetControlObjectiveByID(testUser1.UserCtx, resp.CreateControlObjective.ControlObjective.ID)
				assert.NilError(t, err)
				assert.Assert(t, res != nil)
				assert.Check(t, is.Equal(resp.CreateControlObjective.ControlObjective.ID, res.ControlObjective.ID))
			}

			(&Cleanup[*generated.ControlObjectiveDeleteOne]{client: suite.client.db.ControlObjective, ID: resp.CreateControlObjective.ControlObjective.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}

	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, IDs: []string{program1.ID, program2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: programAnotherUser.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{blockedGroup.ID, viewerGroup.ID, groupMember.GroupID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateControlObjective(t *testing.T) {
	program := (&ProgramBuilder{client: suite.client, EditorIDs: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	controlObjective := (&ControlObjectiveBuilder{client: suite.client, ProgramID: program.ID}).MustNew(testUser1.UserCtx, t)

	// create another admin user and add them to the same organization and group as testUser1
	// this will allow us to test the group editor/viewer permissions
	anotherAdminUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser1.UserCtx, t, &anotherAdminUser, enums.RoleAdmin, testUser1.OrganizationID)

	groupMember := (&GroupMemberBuilder{client: suite.client, UserID: anotherAdminUser.ID}).MustNew(testUser1.UserCtx, t)

	// ensure the user does not currently have access to the control objective
	_, err := suite.client.api.GetControlObjectiveByID(anotherAdminUser.UserCtx, controlObjective.ID)
	assert.ErrorContains(t, err, notFoundErrorMsg)

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
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
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
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, revision bump",
			request: testclient.UpdateControlObjectiveInput{
				Status:       &enums.ObjectiveActiveStatus,
				RevisionBump: &models.Major,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "invalid revision",
			request: testclient.UpdateControlObjectiveInput{
				Revision: lo.ToPtr("1.1"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "revision, invalid semver value",
		},
		{
			name: "update not allowed, not permissions in same org",
			request: testclient.UpdateControlObjectiveInput{
				Status: &enums.ObjectiveActiveStatus,
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateControlObjectiveInput{
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
				res, err := suite.client.api.GetControlObjectiveByID(anotherAdminUser.UserCtx, controlObjective.ID)
				assert.NilError(t, err)
				assert.Check(t, res != nil)
				assert.Check(t, is.Equal(controlObjective.ID, res.ControlObjective.ID))
			}
		})
	}
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: program.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlObjectiveDeleteOne]{client: suite.client.db.ControlObjective, ID: controlObjective.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, ID: groupMember.GroupID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteControlObjective(t *testing.T) {
	// create objects to be deleted
	controlObjective1 := (&ControlObjectiveBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	controlObjective2 := (&ControlObjectiveBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: controlObjective1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  controlObjective1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: controlObjective2.ID,
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

				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteControlObjective.DeletedID))
		})
	}
}
