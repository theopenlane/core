package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/ulids"
)

func (suite *GraphTestSuite) TestQueryControl() {
	t := suite.T()

	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can create a control
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)

	// add test cases for querying the control
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
			name:     "control not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "control not found, using not authorized user",
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			// setup the control if it is not already created
			if tc.queryID == "" {
				resp, err := suite.client.api.CreateControl(testUser1.UserCtx,
					openlaneclient.CreateControlInput{
						Name:       "Control",
						ProgramIDs: []string{program.ID},
					})

				require.NoError(t, err)
				require.NotNil(t, resp)

				tc.queryID = resp.CreateControl.Control.ID
			}

			resp, err := tc.client.GetControlByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			require.NotEmpty(t, resp.Control)

			assert.Equal(t, tc.queryID, resp.Control.ID)
			assert.NotEmpty(t, resp.Control.Name)

			require.Len(t, resp.Control.Programs, 1)
			assert.NotEmpty(t, resp.Control.Programs[0].ID)
		})
	}
}

func (suite *GraphTestSuite) TestQueryControls() {
	t := suite.T()

	// create multiple objects to be queried using testUser1
	(&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			name:            "another user, no controls should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllControls(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.Controls.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateControl() {
	t := suite.T()

	program1 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	programAnotherUser := (&ProgramBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	// add adminUser to the program so that they can create a control associated with the program1
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program1.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)

	// create groups to be associated with the control
	blockedGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	viewerGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.CreateControlInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateControlInput{
				Name: "Control",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateControlInput{
				Name:             "Another Control",
				Description:      lo.ToPtr("A description of the Control"),
				Status:           lo.ToPtr("mitigated"),
				ControlType:      lo.ToPtr("operational"),
				Version:          lo.ToPtr("1.0.0"),
				ControlNumber:    lo.ToPtr("1.1"),
				Family:           lo.ToPtr("AC"),
				Class:            lo.ToPtr("AC-1"),
				Source:           lo.ToPtr("NIST framework"),
				MappedFrameworks: lo.ToPtr("NIST"),
				Satisfies:        lo.ToPtr("AC-1, AC-2"),
				Details:          map[string]interface{}{"stuff": "things"},
				ProgramIDs:       []string{program1.ID, program2.ID}, // multiple programs
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "add groups",
			request: openlaneclient.CreateControlInput{
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
			request: openlaneclient.CreateControlInput{
				Name:    "Control",
				OwnerID: testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "using pat with no owner id",
			request: openlaneclient.CreateControlInput{
				Name: "Control",
			},
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			expectedErr: "organization_id is required",
		},
		{
			name: "using api token",
			request: openlaneclient.CreateControlInput{
				Name: "Control",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: openlaneclient.CreateControlInput{
				Name: "Control",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user authorized, they were added to the program",
			request: openlaneclient.CreateControlInput{
				Name:       "Control",
				ProgramIDs: []string{program1.ID},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "user authorized, user not authorized to one of the programs",
			request: openlaneclient.CreateControlInput{
				Name:       "Control",
				ProgramIDs: []string{program1.ID, program2.ID},
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "missing required name",
			request:     openlaneclient.CreateControlInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "user not authorized, no permissions to one of the programs",
			request: openlaneclient.CreateControlInput{
				Name:       "Control",
				ProgramIDs: []string{programAnotherUser.ID, program1.ID},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateControl(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// check required fields
			require.NotEmpty(t, resp.CreateControl.Control.ID)
			assert.Equal(t, tc.request.Name, resp.CreateControl.Control.Name)

			// ensure the program is set
			if len(tc.request.ProgramIDs) > 0 {
				require.NotEmpty(t, resp.CreateControl.Control.Programs)
				require.Len(t, resp.CreateControl.Control.Programs, len(tc.request.ProgramIDs))

				for i, p := range resp.CreateControl.Control.Programs {
					assert.Equal(t, tc.request.ProgramIDs[i], p.ID)
				}
			} else {
				assert.Empty(t, resp.CreateControl.Control.Programs)
			}

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.CreateControl.Control.Description)
			} else {
				assert.Empty(t, resp.CreateControl.Control.Description)
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, *resp.CreateControl.Control.Status)
			} else {
				assert.Empty(t, resp.CreateControl.Control.Status)
			}

			if tc.request.ControlType != nil {
				assert.Equal(t, *tc.request.ControlType, *resp.CreateControl.Control.ControlType)
			} else {
				assert.Empty(t, resp.CreateControl.Control.ControlType)
			}

			if tc.request.Version != nil {
				assert.Equal(t, *tc.request.Version, *resp.CreateControl.Control.Version)
			} else {
				assert.Empty(t, resp.CreateControl.Control.Version)
			}

			if tc.request.ControlNumber != nil {
				assert.Equal(t, *tc.request.ControlNumber, *resp.CreateControl.Control.ControlNumber)
			} else {
				assert.Empty(t, resp.CreateControl.Control.ControlNumber)
			}

			if tc.request.Family != nil {
				assert.Equal(t, *tc.request.Family, *resp.CreateControl.Control.Family)
			} else {
				assert.Empty(t, resp.CreateControl.Control.Family)
			}

			if tc.request.Class != nil {
				assert.Equal(t, *tc.request.Class, *resp.CreateControl.Control.Class)
			} else {
				assert.Empty(t, resp.CreateControl.Control.Class)
			}

			if tc.request.Source != nil {
				assert.Equal(t, *tc.request.Source, *resp.CreateControl.Control.Source)
			} else {
				assert.Empty(t, resp.CreateControl.Control.Source)
			}

			if tc.request.MappedFrameworks != nil {
				assert.Equal(t, *tc.request.MappedFrameworks, *resp.CreateControl.Control.MappedFrameworks)
			} else {
				assert.Empty(t, resp.CreateControl.Control.MappedFrameworks)
			}

			if tc.request.Satisfies != nil {
				assert.Equal(t, *tc.request.Satisfies, *resp.CreateControl.Control.Satisfies)
			} else {
				assert.Empty(t, resp.CreateControl.Control.Satisfies)
			}

			if tc.request.Details != nil {
				assert.Equal(t, tc.request.Details, resp.CreateControl.Control.Details)
			} else {
				assert.Empty(t, resp.CreateControl.Control.Details)
			}

			if len(tc.request.EditorIDs) > 0 {
				require.Len(t, resp.CreateControl.Control.Editors, 1)
				for _, edge := range resp.CreateControl.Control.Editors {
					assert.Equal(t, testUser1.GroupID, edge.ID)
				}
			}

			if len(tc.request.BlockedGroupIDs) > 0 {
				require.Len(t, resp.CreateControl.Control.BlockedGroups, 1)
				for _, edge := range resp.CreateControl.Control.BlockedGroups {
					assert.Equal(t, blockedGroup.ID, edge.ID)
				}
			}

			if len(tc.request.ViewerIDs) > 0 {
				require.Len(t, resp.CreateControl.Control.Viewers, 1)
				for _, edge := range resp.CreateControl.Control.Viewers {
					assert.Equal(t, viewerGroup.ID, edge.ID)
				}
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateControl() {
	t := suite.T()

	program1 := (&ProgramBuilder{client: suite.client, EditorIDs: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client, EditorIDs: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	control := (&ControlBuilder{client: suite.client, ProgramID: program1.ID}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can update the control
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program1.ID, UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).MustNew(testUser1.UserCtx, t)

	// create another admin user and add them to the same organization and group as testUser1
	// this will allow us to test the group editor/viewer permissions
	anotherAdminUser := suite.userBuilder(context.Background())
	suite.addUserToOrganization(testUser1.UserCtx, &anotherAdminUser, enums.RoleAdmin, testUser1.OrganizationID)

	(&GroupMemberBuilder{client: suite.client, UserID: anotherAdminUser.ID, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)

	// ensure the user does not currently have access to the control
	res, err := suite.client.api.GetControlByID(anotherAdminUser.UserCtx, control.ID)
	require.Error(t, err)
	require.Nil(t, res)

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateControlInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: openlaneclient.UpdateControlInput{
				Description:   lo.ToPtr("Updated description"),
				AddProgramIDs: []string{program1.ID, program2.ID}, // add multiple programs
				AddViewerIDs:  []string{testUser1.GroupID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: openlaneclient.UpdateControlInput{
				Status:           lo.ToPtr("mitigated"),
				Tags:             []string{"tag1", "tag2"},
				Version:          lo.ToPtr("1.0.1"),
				ControlNumber:    lo.ToPtr("1.2"),
				Family:           lo.ToPtr("AB"),
				Class:            lo.ToPtr("AB-2"),
				Source:           lo.ToPtr("ISO27001"),
				MappedFrameworks: lo.ToPtr("ISO"),
				Satisfies:        lo.ToPtr("AB-2, AB-3"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, not permissions in same org",
			request: openlaneclient.UpdateControlInput{
				MappedFrameworks: lo.ToPtr("SOC"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "update allowed, user added to one of the programs",
			request: openlaneclient.UpdateControlInput{
				MappedFrameworks: lo.ToPtr("SOC2"),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "update not allowed, no permissions",
			request: openlaneclient.UpdateControlInput{
				MappedFrameworks: lo.ToPtr("SOC"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateControl(tc.ctx, control.ID, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.UpdateControl.Control.Description)
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, *resp.UpdateControl.Control.Status)
			}

			if tc.request.Tags != nil {
				assert.ElementsMatch(t, tc.request.Tags, resp.UpdateControl.Control.Tags)
			}

			if tc.request.Version != nil {
				assert.Equal(t, *tc.request.Version, *resp.UpdateControl.Control.Version)
			}

			if tc.request.ControlNumber != nil {
				assert.Equal(t, *tc.request.ControlNumber, *resp.UpdateControl.Control.ControlNumber)
			}

			if tc.request.Family != nil {
				assert.Equal(t, *tc.request.Family, *resp.UpdateControl.Control.Family)
			}

			if tc.request.Class != nil {
				assert.Equal(t, *tc.request.Class, *resp.UpdateControl.Control.Class)
			}

			if tc.request.Source != nil {
				assert.Equal(t, *tc.request.Source, *resp.UpdateControl.Control.Source)
			}

			if tc.request.MappedFrameworks != nil {
				assert.Equal(t, *tc.request.MappedFrameworks, *resp.UpdateControl.Control.MappedFrameworks)
			}

			if tc.request.Satisfies != nil {
				assert.Equal(t, *tc.request.Satisfies, *resp.UpdateControl.Control.Satisfies)
			}

			if tc.request.Details != nil {
				assert.Equal(t, tc.request.Details, resp.UpdateControl.Control.Details)
			}

			if len(tc.request.AddViewerIDs) > 0 {
				require.Len(t, resp.UpdateControl.Control.Viewers, 1)
				for _, edge := range resp.UpdateControl.Control.Viewers {
					assert.Equal(t, testUser1.GroupID, edge.ID)
				}

				// ensure the user has access to the control now
				res, err := suite.client.api.GetControlByID(anotherAdminUser.UserCtx, control.ID)
				require.NoError(t, err)
				require.NotEmpty(t, res)
				assert.Equal(t, control.ID, res.Control.ID)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteControl() {
	t := suite.T()

	// create objects to be deleted
	Control1 := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	Control2 := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  Control1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: Control1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  Control1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: Control2.ID,
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
			resp, err := tc.client.DeleteControl(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteControl.DeletedID)
		})
	}
}
