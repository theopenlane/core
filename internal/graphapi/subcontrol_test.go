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

func (suite *GraphTestSuite) TestQuerySubcontrol() {
	t := suite.T()

	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can create a subcontrol
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)

	// add test cases for querying the subcontrol
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
			name:     "read only user, same org, no access to the parent control",
			client:   suite.client.api,
			ctx:      viewOnlyUser.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:   "admin user, access to the parent control via the program",
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "happy path using personal access token",
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:     "subcontrol not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "subcontrol not found, using not authorized user",
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			// setup the subcontrol if it is not already created
			if tc.queryID == "" {
				// create the control first
				control, err := suite.client.api.CreateControl(testUser1.UserCtx,
					openlaneclient.CreateControlInput{
						RefCode:    "SC-1",
						ProgramIDs: []string{program.ID},
					})

				require.NoError(t, err)
				require.NotNil(t, control)

				resp, err := suite.client.api.CreateSubcontrol(testUser1.UserCtx,
					openlaneclient.CreateSubcontrolInput{
						RefCode:   "SC-1",
						ControlID: control.CreateControl.Control.ID,
					})

				require.NoError(t, err)
				require.NotNil(t, resp)

				tc.queryID = resp.CreateSubcontrol.Subcontrol.ID
			}

			resp, err := tc.client.GetSubcontrolByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			require.NotEmpty(t, resp.Subcontrol)

			assert.Equal(t, tc.queryID, resp.Subcontrol.ID)
			assert.NotEmpty(t, resp.Subcontrol.RefCode)

			assert.NotEmpty(t, resp.Subcontrol.ControlID)
		})
	}
}

func (suite *GraphTestSuite) TestQuerySubcontrols() {
	t := suite.T()

	// create multiple objects to be queried using testUser1
	(&SubcontrolBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&SubcontrolBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			name:            "another user, no subcontrols should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllSubcontrols(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.Subcontrols.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateSubcontrol() {
	t := suite.T()

	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	ownerGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	anotherOwnerGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	delegateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can create a subcontrol
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)

	control1 := (&ControlBuilder{client: suite.client, ProgramID: program.ID}).MustNew(testUser1.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	controlWithOwner := (&ControlBuilder{client: suite.client, ProgramID: program.ID,
		OwnerID: ownerGroup.ID}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name                string
		request             openlaneclient.CreateSubcontrolInput
		createParentControl bool
		client              *openlaneclient.OpenlaneClient
		ctx                 context.Context
		expectedErr         string
	}{
		{
			name: "missing required ref code",
			request: openlaneclient.CreateSubcontrolInput{
				Description: lo.ToPtr("A description of the Subcontrol"),
				ControlID:   control1.ID,
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateSubcontrolInput{
				RefCode:   "SC-1",
				ControlID: control1.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, parent control has owner",
			request: openlaneclient.CreateSubcontrolInput{
				RefCode:   "SC-2",
				ControlID: controlWithOwner.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, parent control has owner, subcontrol should override it",
			request: openlaneclient.CreateSubcontrolInput{
				RefCode:   "SC-2",
				ControlID: controlWithOwner.ID,
				OwnerID:   &anotherOwnerGroup.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateSubcontrolInput{
				RefCode:     "Another Subcontrol",
				Description: lo.ToPtr("A description of the Subcontrol"),
				Status:      &enums.ControlStatusPreparing,
				Tags:        []string{"tag1", "tag2"},
				Source:      &enums.ControlSourceFramework,
				ControlType: &enums.ControlTypeDetective,
				Category:    lo.ToPtr("Availability"),
				Subcategory: lo.ToPtr("Availability-1"),
				MappedCategories: []string{
					"Category1",
					"Category2",
				},
				ControlQuestions: []string{
					"Question 1",
					"Question 2",
				},
				AssessmentObjectives: []*models.AssessmentObjective{
					{
						Class:     "Class 1",
						ID:        "ID-1",
						Objective: "Objective 1",
					},
				},
				AssessmentMethods: []*models.AssessmentMethod{
					{
						Type:   "Examine",
						ID:     "ID-2",
						Method: "Method 1",
					},
				},
				ImplementationGuidance: []*models.ImplementationGuidance{
					{
						ReferenceID: "cc-1",
						Guidance: []string{
							"Step 1",
							"Step 2",
						},
					},
				},
				ControlID: control2.ID,
				ExampleEvidence: []*models.ExampleEvidence{
					{
						DocumentationType: "Policy",
						Description:       "Create a policy",
					},
				},
				DelegateID:     &delegateGroup.ID,
				ControlOwnerID: &ownerGroup.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: openlaneclient.CreateSubcontrolInput{
				RefCode:   "Subcontrol",
				ControlID: control1.ID,
				OwnerID:   &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "using pat, missing owner ID",
			request: openlaneclient.CreateSubcontrolInput{
				RefCode:   "SC-1",
				ControlID: control1.ID,
			},
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			expectedErr: "owner_id is required",
		},
		{
			name: "user not authorized, not enough permissions",
			request: openlaneclient.CreateSubcontrolInput{
				RefCode:   "SC-1",
				ControlID: control1.ID,
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user authorized, they have access to the parent control via the program",
			request: openlaneclient.CreateSubcontrolInput{
				RefCode:   "SC-1",
				ControlID: control1.ID,
			},
			createParentControl: true, // create the parent control first
			client:              suite.client.api,
			ctx:                 adminUser.UserCtx,
		},
		{
			name: "user not authorized to one of the controls",
			request: openlaneclient.CreateSubcontrolInput{
				RefCode:   "SC-1",
				ControlID: control2.ID,
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "missing required control ID",
			request: openlaneclient.CreateSubcontrolInput{
				RefCode: "SC-1",
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "validator failed for field",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.createParentControl {
				// create the control first
				control, err := suite.client.api.CreateControl(testUser1.UserCtx,
					openlaneclient.CreateControlInput{
						RefCode:    "SC",
						ProgramIDs: []string{program.ID},
					})

				require.NoError(t, err)
				require.NotNil(t, control)

				tc.request.ControlID = control.CreateControl.Control.ID
			}

			resp, err := tc.client.CreateSubcontrol(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// check required fields
			require.NotEmpty(t, resp.CreateSubcontrol.Subcontrol.ID)
			assert.Equal(t, tc.request.RefCode, resp.CreateSubcontrol.Subcontrol.RefCode)

			assert.NotEmpty(t, resp.CreateSubcontrol.Subcontrol.DisplayID)
			assert.Contains(t, resp.CreateSubcontrol.Subcontrol.DisplayID, "SCL-")

			assert.NotEmpty(t, resp.CreateSubcontrol.Subcontrol.RefCode)
			assert.Equal(t, tc.request.RefCode, resp.CreateSubcontrol.Subcontrol.RefCode)

			// ensure the control is set
			require.NotEmpty(t, resp.CreateSubcontrol.Subcontrol.ControlID)
			assert.Equal(t, tc.request.ControlID, resp.CreateSubcontrol.Subcontrol.ControlID)

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.CreateSubcontrol.Subcontrol.Description)
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.Description)
			}

			assert.Equal(t, enums.ControlStatusPreparing, *resp.CreateSubcontrol.Subcontrol.Status)

			if tc.request.Source != nil {
				assert.Equal(t, *tc.request.Source, *resp.CreateSubcontrol.Subcontrol.Source)
			} else {
				assert.Equal(t, enums.ControlSourceUserDefined, *resp.CreateSubcontrol.Subcontrol.Source)
			}

			if tc.request.ControlOwnerID != nil {
				require.NotNil(t, resp.CreateSubcontrol.Subcontrol.ControlOwner)
				assert.Equal(t, *tc.request.ControlOwnerID, resp.CreateSubcontrol.Subcontrol.ControlOwner.ID)
			} else if tc.request.ControlID == controlWithOwner.ID {
				// it should inherit the owner from the parent control if it was set
				require.NotNil(t, resp.CreateSubcontrol.Subcontrol.ControlOwner)
				assert.Equal(t, controlWithOwner.OwnerID, resp.CreateSubcontrol.Subcontrol.ControlOwner.ID)
			} else {
				assert.Nil(t, resp.CreateSubcontrol.Subcontrol.ControlOwner)
			}

			if tc.request.DelegateID != nil {
				require.NotNil(t, resp.CreateSubcontrol.Subcontrol.Delegate)
				assert.Equal(t, *tc.request.DelegateID, resp.CreateSubcontrol.Subcontrol.Delegate.ID)
			} else {
				assert.Nil(t, resp.CreateSubcontrol.Subcontrol.Delegate)
			}

			// ensure the org owner has access to the subcontrol that was created by an api token
			if tc.client == suite.client.apiWithToken {
				res, err := suite.client.api.GetSubcontrolByID(testUser1.UserCtx, resp.CreateSubcontrol.Subcontrol.ID)
				require.NoError(t, err)
				require.NotEmpty(t, res)
				assert.Equal(t, resp.CreateSubcontrol.Subcontrol.ID, res.Subcontrol.ID)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateSubcontrol() {
	t := suite.T()

	control1 := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	subcontrol := (&SubcontrolBuilder{client: suite.client, ControlID: control1.ID}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateSubcontrolInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: openlaneclient.UpdateSubcontrolInput{
				Description: lo.ToPtr("Updated description"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: openlaneclient.UpdateSubcontrolInput{
				Status: &enums.ControlStatusPreparing,
				Tags:   []string{"tag1", "tag2"},
				Source: lo.ToPtr(enums.ControlSourceFramework),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update ref code to empty",
			request: openlaneclient.UpdateSubcontrolInput{
				RefCode: lo.ToPtr(""),
			},
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			expectedErr: "value is less than the required length",
		},
		{
			name: "update not allowed, not permissions in same org",
			request: openlaneclient.UpdateSubcontrolInput{
				MappedCategories: []string{"Category1", "Category2"},
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: openlaneclient.UpdateSubcontrolInput{
				MappedCategories: []string{"Category1", "Category2"},
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateSubcontrol(tc.ctx, subcontrol.ID, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.UpdateSubcontrol.Subcontrol.Description)
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, *resp.UpdateSubcontrol.Subcontrol.Status)
			}

			if tc.request.Tags != nil {
				assert.ElementsMatch(t, tc.request.Tags, resp.UpdateSubcontrol.Subcontrol.Tags)
			}

			if tc.request.Source != nil {
				assert.Equal(t, *tc.request.Source, *resp.UpdateSubcontrol.Subcontrol.Source)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteSubcontrol() {
	t := suite.T()

	// create objects to be deleted
	subcontrol1 := (&SubcontrolBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	subcontrol2 := (&SubcontrolBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  subcontrol1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: subcontrol1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  subcontrol1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: subcontrol2.ID,
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
			resp, err := tc.client.DeleteSubcontrol(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteSubcontrol.DeletedID)
		})
	}
}
