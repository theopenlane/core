package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/core/pkg/testutils"
	"github.com/theopenlane/iam/auth"
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
						RefCode:    "CC-1.1",
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
			assert.NotEmpty(t, resp.Control.RefCode)

			require.Len(t, resp.Control.Programs.Edges, 1)
			assert.NotEmpty(t, resp.Control.Programs.Edges[0].Node.ID)
		})
	}
}

func (suite *GraphTestSuite) TestQueryControls() {
	t := suite.T()

	// create multiple objects to be queried using testUser1
	controlsToCreate := int64(11)
	for range controlsToCreate { // set to 11 to ensure pagination is tested
		(&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	}

	org := (&OrganizationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	userCtxAnotherOrg := auth.NewTestContextWithOrgID(testUser1.ID, org.ID)

	// add a control for the user to another org; this should not be returned for JWT auth, since it's
	// restricted to a single org. PAT auth would return it if both orgs are authorized on the token
	(&ControlBuilder{client: suite.client}).MustNew(userCtxAnotherOrg, t)

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
		first           *int64
		last            *int64
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: testutils.MaxResultLimit,
		},
		{
			name:            "happy path, with first set",
			client:          suite.client.api,
			first:           lo.ToPtr(int64(5)),
			ctx:             testUser1.UserCtx,
			expectedResults: 5,
		},
		{
			name:            "happy path, with last set",
			client:          suite.client.api,
			first:           lo.ToPtr(int64(3)),
			ctx:             testUser1.UserCtx,
			expectedResults: 3,
		},
		{
			name:            "first set over max (10 in test)",
			client:          suite.client.api,
			first:           lo.ToPtr(int64(11)),
			ctx:             testUser1.UserCtx,
			expectedResults: testutils.MaxResultLimit,
		},
		{
			name:            "last set over max (10 in test)",
			client:          suite.client.api,
			last:            lo.ToPtr(int64(11)),
			ctx:             testUser1.UserCtx,
			expectedResults: testutils.MaxResultLimit,
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
			expectedResults: testutils.MaxResultLimit,
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
			if tc.first != nil || tc.last != nil {
				resp, err := tc.client.GetControls(tc.ctx, tc.first, tc.last, nil)
				require.NoError(t, err)
				require.NotNil(t, resp)

				assert.Len(t, resp.Controls.Edges, tc.expectedResults)
				assert.Equal(t, int64(11), resp.Controls.TotalCount)

				// if we are pulling the last, there won't be a next page, but there will be a previous page
				if tc.last != nil {
					assert.True(t, resp.Controls.PageInfo.HasPreviousPage)
				} else {
					assert.True(t, resp.Controls.PageInfo.HasNextPage)
				}

				return
			}

			resp, err := tc.client.GetAllControls(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.Controls.Edges, tc.expectedResults)

			if tc.expectedResults > 0 {
				assert.Equal(t, int64(controlsToCreate), resp.Controls.TotalCount)
				assert.True(t, resp.Controls.PageInfo.HasNextPage)
			} else {
				assert.Equal(t, 0, len(resp.Controls.Edges))
				assert.Equal(t, int64(0), resp.Controls.TotalCount)
				assert.False(t, resp.Controls.PageInfo.HasNextPage)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateControl() {
	t := suite.T()

	program1 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	programAnotherUser := (&ProgramBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	ownerGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	deleteGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
				RefCode: "A-1",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateControlInput{
				RefCode:          "A-2",
				Description:      lo.ToPtr("A description of the Control"),
				Status:           &enums.ControlStatusPreparing,
				Tags:             []string{"tag1", "tag2"},
				ControlType:      &enums.ControlTypeDetective,
				Category:         lo.ToPtr("Availability"),
				CategoryID:       lo.ToPtr("A"),
				Subcategory:      lo.ToPtr("Additional Criteria for Availability"),
				MappedCategories: []string{"Govern", "Protect"},
				ControlQuestions: []string{"What is the control question?"},
				AssessmentObjectives: []*models.AssessmentObjective{
					{
						Class:     "class",
						ID:        "id",
						Objective: "objective for the control",
					},
				},
				AssessmentMethods: []*models.AssessmentMethod{
					{
						ID:     "id",
						Type:   "Examine",
						Method: "method of assessment for the control",
					},
				},
				ImplementationGuidance: []*models.ImplementationGuidance{
					{
						ReferenceID: "ref-id",
						Guidance: []string{
							"guidance 1",
							"guidance 2",
						},
					},
				},
				ExampleEvidence: []*models.ExampleEvidence{
					{
						DocumentationType: "policy",
						Description:       "description of the example evidence",
					},
				},
				References: []*models.Reference{
					{
						Name: "name of ref",
						URL:  "https://example.com",
					},
				},
				ControlOwnerID: &ownerGroup.ID,
				DelegateID:     &deleteGroup.ID,
				Source:         &enums.ControlSourceFramework,
				ProgramIDs:     []string{program1.ID, program2.ID}, // multiple programs
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "add groups",
			request: openlaneclient.CreateControlInput{
				RefCode:         "A-3",
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
				RefCode: "A-4",
				OwnerID: &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "using pat with no owner id",
			request: openlaneclient.CreateControlInput{
				RefCode: "A-4",
			},
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			expectedErr: "owner_id is required",
		},
		{
			name: "using api token",
			request: openlaneclient.CreateControlInput{
				RefCode: "A-5",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: openlaneclient.CreateControlInput{
				RefCode: "A-6",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user authorized, they were added to the program",
			request: openlaneclient.CreateControlInput{
				RefCode:    "A-7",
				ProgramIDs: []string{program1.ID},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "user authorized, user not authorized to one of the programs",
			request: openlaneclient.CreateControlInput{
				RefCode:    "A-8",
				ProgramIDs: []string{program1.ID, program2.ID},
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "missing required ref code",
			request: openlaneclient.CreateControlInput{
				Description: lo.ToPtr("A description of the Control"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "user not authorized, no permissions to one of the programs",
			request: openlaneclient.CreateControlInput{
				RefCode:    "A-9",
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
			assert.Equal(t, tc.request.RefCode, resp.CreateControl.Control.RefCode)

			assert.NotEmpty(t, resp.CreateControl.Control.DisplayID)
			assert.Contains(t, resp.CreateControl.Control.DisplayID, "CTL-")

			assert.NotEmpty(t, resp.CreateControl.Control.RefCode)
			assert.Equal(t, tc.request.RefCode, resp.CreateControl.Control.RefCode)

			// ensure the program is set
			if len(tc.request.ProgramIDs) > 0 {
				require.NotEmpty(t, resp.CreateControl.Control.Programs)
				require.Len(t, resp.CreateControl.Control.Programs.Edges, len(tc.request.ProgramIDs))

				for i, p := range resp.CreateControl.Control.Programs.Edges {
					assert.Equal(t, tc.request.ProgramIDs[i], p.Node.ID)
				}
			} else {
				assert.Empty(t, resp.CreateControl.Control.Programs.Edges)
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
				assert.Equal(t, enums.ControlTypePreventative, *resp.CreateControl.Control.ControlType) // default value
			}

			if tc.request.Source != nil {
				assert.Equal(t, *tc.request.Source, *resp.CreateControl.Control.Source)
			} else {
				assert.Equal(t, enums.ControlSourceUserDefined, *resp.CreateControl.Control.Source)
			}

			if tc.request.Category != nil {
				assert.Equal(t, *tc.request.Category, *resp.CreateControl.Control.Category)
			} else {
				assert.Empty(t, resp.CreateControl.Control.Category)
			}

			if tc.request.CategoryID != nil {
				assert.Equal(t, *tc.request.CategoryID, *resp.CreateControl.Control.CategoryID)
			} else {
				assert.Empty(t, resp.CreateControl.Control.CategoryID)
			}

			if tc.request.Subcategory != nil {
				assert.Equal(t, *tc.request.Subcategory, *resp.CreateControl.Control.Subcategory)
			} else {
				assert.Empty(t, resp.CreateControl.Control.Subcategory)
			}

			if tc.request.MappedCategories != nil {
				assert.ElementsMatch(t, tc.request.MappedCategories, resp.CreateControl.Control.MappedCategories)
			} else {
				assert.Empty(t, resp.CreateControl.Control.MappedCategories)
			}

			if tc.request.ControlQuestions != nil {
				assert.ElementsMatch(t, tc.request.ControlQuestions, resp.CreateControl.Control.ControlQuestions)
			} else {
				assert.Empty(t, resp.CreateControl.Control.ControlQuestions)
			}

			if tc.request.AssessmentObjectives != nil {
				require.Len(t, resp.CreateControl.Control.AssessmentObjectives, len(tc.request.AssessmentObjectives))
				assert.ElementsMatch(t, tc.request.AssessmentObjectives, resp.CreateControl.Control.AssessmentObjectives)
			} else {
				assert.Empty(t, resp.CreateControl.Control.AssessmentObjectives)
			}

			if tc.request.AssessmentMethods != nil {
				require.Len(t, resp.CreateControl.Control.AssessmentMethods, len(tc.request.AssessmentMethods))
				assert.ElementsMatch(t, tc.request.AssessmentMethods, resp.CreateControl.Control.AssessmentMethods)
			} else {
				assert.Empty(t, resp.CreateControl.Control.AssessmentMethods)
			}

			if tc.request.ImplementationGuidance != nil {
				require.Len(t, resp.CreateControl.Control.ImplementationGuidance, len(tc.request.ImplementationGuidance))
				assert.ElementsMatch(t, tc.request.ImplementationGuidance, resp.CreateControl.Control.ImplementationGuidance)
			} else {
				assert.Empty(t, resp.CreateControl.Control.ImplementationGuidance)
			}

			if tc.request.ExampleEvidence != nil {
				require.Len(t, resp.CreateControl.Control.ExampleEvidence, len(tc.request.ExampleEvidence))
				assert.ElementsMatch(t, tc.request.ExampleEvidence, resp.CreateControl.Control.ExampleEvidence)
			} else {
				assert.Empty(t, resp.CreateControl.Control.ExampleEvidence)
			}

			if tc.request.References != nil {
				require.Len(t, resp.CreateControl.Control.References, len(tc.request.References))
				assert.ElementsMatch(t, tc.request.References, resp.CreateControl.Control.References)
			} else {
				assert.Empty(t, resp.CreateControl.Control.References)
			}

			if tc.request.ControlOwnerID != nil {
				require.NotEmpty(t, resp.CreateControl.Control.ControlOwner)
				assert.Equal(t, *tc.request.ControlOwnerID, resp.CreateControl.Control.ControlOwner.ID)
			} else {
				assert.Empty(t, resp.CreateControl.Control.ControlOwner)
			}

			if tc.request.DelegateID != nil {
				require.NotEmpty(t, resp.CreateControl.Control.Delegate)
				assert.Equal(t, *tc.request.DelegateID, resp.CreateControl.Control.Delegate.ID)
			} else {
				assert.Empty(t, resp.CreateControl.Control.Delegate)
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

			// ensure the org owner has access to the control that was created by an api token
			if tc.client == suite.client.apiWithToken {
				res, err := suite.client.api.GetControlByID(testUser1.UserCtx, resp.CreateControl.Control.ID)
				require.NoError(t, err)
				require.NotEmpty(t, res)
				assert.Equal(t, resp.CreateControl.Control.ID, res.Control.ID)
			}

		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateControlsByClone() {
	t := suite.T()

	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program.ID, UserID: viewOnlyUser.ID}).MustNew(testUser1.UserCtx, t)

	programAnotherOrg := (&ProgramBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	publicStandard := (&StandardBuilder{client: suite.client, IsPublic: true}).MustNew(systemAdminUser.UserCtx, t)

	// create standard with controls to clone
	numControls := int64(20)
	controls := []*generated.Control{}
	controlIDs := make([]string, 0, numControls)
	for range numControls {
		control := (&ControlBuilder{client: suite.client, StandardID: publicStandard.ID, AllFields: true}).MustNew(systemAdminUser.UserCtx, t)
		controls = append(controls, control)
		controlIDs = append(controlIDs, control.ID)
	}

	// ensure the standard exists and has the correct number of controls for the non-system admin user
	standard, err := suite.client.api.GetStandardByID(testUser2.UserCtx, publicStandard.ID)
	require.NoError(t, err)
	require.NotNil(t, standard)
	assert.Equal(t, standard.Standard.Controls.TotalCount, numControls)

	// create org owned control
	orgOwnedControl := (&ControlBuilder{client: suite.client, AllFields: true}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name                 string
		request              openlaneclient.CloneControlInput
		expectedControls     []*generated.Control
		client               *openlaneclient.OpenlaneClient
		ctx                  context.Context
		expectNoAccessViewer bool
		expectedErr          string
	}{
		{
			name: "happy path, clone single control",
			request: openlaneclient.CloneControlInput{
				ControlIDs: []string{controlIDs[0]},
			},
			expectedControls: controls[:1],
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
		},
		{
			name: "happy path, all controls under standard",
			request: openlaneclient.CloneControlInput{
				ControlIDs: controlIDs,
			},
			expectedControls: controls,
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
		},
		{
			name: "happy path, all controls under standard with program",
			request: openlaneclient.CloneControlInput{
				ControlIDs: controlIDs,
				ProgramID:  &program.ID,
			},
			expectedControls: controls,
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
		},
		{
			name: "all controls under standard with program no access",
			request: openlaneclient.CloneControlInput{
				ControlIDs: controlIDs,
				ProgramID:  &programAnotherOrg.ID,
			},
			expectedControls: controls,
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			expectedErr:      notAuthorizedErrorMsg,
		},
		{
			name: "happy path, clone control under org",
			request: openlaneclient.CloneControlInput{
				ControlIDs: []string{orgOwnedControl.ID},
			},
			expectedControls: []*generated.Control{orgOwnedControl},
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			// created directly under organization with no program, view only user should not have access
			expectNoAccessViewer: true,
		},
		{
			name: "happy path, clone control under org with program",
			request: openlaneclient.CloneControlInput{
				ControlIDs: []string{orgOwnedControl.ID},
				ProgramID:  &program.ID,
			},
			expectedControls: []*generated.Control{orgOwnedControl},
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
		},
		{
			name: "happy path, clone single control using personal access token",
			request: openlaneclient.CloneControlInput{
				ControlIDs: []string{controlIDs[0]},
				OwnerID:    &testUser1.OrganizationID,
			},
			expectedControls: controls[:1],
			client:           suite.client.apiWithPAT,
			ctx:              context.Background(),
		},
		{
			name: "clone single control using personal access token, missing owner id",
			request: openlaneclient.CloneControlInput{
				ControlIDs: []string{controlIDs[0]},
			},
			expectedControls: controls[:1],
			client:           suite.client.apiWithPAT,
			ctx:              context.Background(),
			expectedErr:      "owner_id is required",
		},
		{
			name: "happy path, clone single control using api token",
			request: openlaneclient.CloneControlInput{
				ControlIDs: []string{controlIDs[0]},
			},
			expectedControls: controls[:1],
			client:           suite.client.apiWithToken,
			ctx:              context.Background(),
		},
		{
			name: "clone control under org, no access to control",
			request: openlaneclient.CloneControlInput{
				ControlIDs: []string{orgOwnedControl.ID},
			},
			expectedControls: []*generated.Control{orgOwnedControl},
			client:           suite.client.api,
			ctx:              testUser2.UserCtx,
			expectedErr:      notAuthorizedErrorMsg,
		},
		{
			name:             "clone control under org, empty request",
			request:          openlaneclient.CloneControlInput{},
			expectedControls: []*generated.Control{orgOwnedControl},
			client:           suite.client.api,
			ctx:              testUser2.UserCtx,
			expectedErr:      notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateControlsByClone(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			for i, control := range resp.CreateControlsByClone.Controls {
				// check required fields
				require.NotEmpty(t, control.ID)
				require.NotEmpty(t, control.DisplayID)
				require.NotEmpty(t, control.RefCode)

				// all cloned controls should have an owner
				assert.NotEmpty(t, control.OwnerID)

				if tc.request.ProgramID != nil {
					require.NotEmpty(t, control.Programs)
					require.Len(t, control.Programs.Edges, 1)
					assert.Equal(t, *tc.request.ProgramID, control.Programs.Edges[0].Node.ID)
				} else {
					assert.Empty(t, control.Programs.Edges)
				}

				// check the cloned control fields are set and match the original control
				assert.Equal(t, tc.expectedControls[i].RefCode, control.RefCode)
				assert.Equal(t, tc.expectedControls[i].ControlType, *control.ControlType)
				assert.Equal(t, tc.expectedControls[i].Category, *control.Category)
				assert.Equal(t, tc.expectedControls[i].CategoryID, *control.CategoryID)
				assert.Equal(t, tc.expectedControls[i].Subcategory, *control.Subcategory)
				assert.Equal(t, tc.expectedControls[i].MappedCategories, control.MappedCategories)
				assert.Equal(t, tc.expectedControls[i].ControlQuestions, control.ControlQuestions)
				assert.Equal(t, tc.expectedControls[i].Tags, control.Tags)
				assert.Equal(t, tc.expectedControls[i].Status, *control.Status)
				assert.Equal(t, tc.expectedControls[i].ControlType, *control.ControlType)
				assert.Equal(t, tc.expectedControls[i].Source, *control.Source)
				assert.Equal(t, tc.expectedControls[i].StandardID, *control.StandardID)

				for j, ao := range control.AssessmentObjectives {
					assert.Equal(t, tc.expectedControls[i].AssessmentObjectives[j], *ao)
				}

				for j, am := range control.AssessmentMethods {
					assert.Equal(t, tc.expectedControls[i].AssessmentMethods[j], *am)
				}

				for j, ig := range control.ImplementationGuidance {
					assert.Equal(t, tc.expectedControls[i].ImplementationGuidance[j], *ig)
				}

				for j, ref := range control.References {
					assert.Equal(t, tc.expectedControls[i].References[j], *ref)
				}

				for j, ee := range control.ExampleEvidence {
					assert.Equal(t, tc.expectedControls[i].ExampleEvidence[j], *ee)
				}

				// ensure the org owner has access to the control that was created by an api token
				if tc.client == suite.client.apiWithToken {
					res, err := suite.client.api.GetControlByID(testUser1.UserCtx, control.ID)
					require.NoError(t, err)
					require.NotEmpty(t, res)
					assert.Equal(t, control.ID, res.Control.ID)
				}

				// ensure view only user can see the control created by the admin user
				// TODO (sfunk): verify its okay users without access to the program can see a control
				// from a public standard
				res, err := suite.client.api.GetControlByID(viewOnlyUser.UserCtx, control.ID)
				if tc.expectNoAccessViewer {
					require.Error(t, err)
					assert.ErrorContains(t, err, notFoundErrorMsg)
					assert.Nil(t, res)
				} else {
					require.NoError(t, err)
					require.NotEmpty(t, res)
					assert.Equal(t, control.ID, res.Control.ID)
				}

				// ensure a user outside my organization cannot get the control
				res, err = suite.client.api.GetControlByID(testUser2.UserCtx, control.ID)
				require.Nil(t, res)
				require.Error(t, err)
				assert.ErrorContains(t, err, notFoundErrorMsg)
			}
		})
	}

	// cleanup created controls and standards
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: publicStandard.ID}).MustDelete(testUser1.UserCtx, suite)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: orgOwnedControl.ID}).MustDelete(testUser1.UserCtx, suite)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: controlIDs}).MustDelete(testUser1.UserCtx, suite)
}

func (suite *GraphTestSuite) TestMutationUpdateControl() {
	t := suite.T()

	program1 := (&ProgramBuilder{client: suite.client, EditorIDs: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client, EditorIDs: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	control := (&ControlBuilder{client: suite.client, ProgramID: program1.ID}).MustNew(testUser1.UserCtx, t)

	ownerGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	deleteGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can update the control
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program1.ID, UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).MustNew(testUser1.UserCtx, t)

	// create another admin user and add them to the same organization and group as testUser1
	// this will allow us to test the group editor/viewer permissions
	anotherAdminUser := suite.userBuilder(context.Background())
	suite.addUserToOrganization(testUser1.UserCtx, &anotherAdminUser, enums.RoleAdmin, testUser1.OrganizationID)

	groupMember := (&GroupMemberBuilder{client: suite.client, UserID: anotherAdminUser.ID}).MustNew(testUser1.UserCtx, t)

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
				AddProgramIDs: []string{program1.ID, program2.ID}, // add multiple programs (one already associated)
				AddViewerIDs:  []string{groupMember.GroupID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: openlaneclient.UpdateControlInput{
				Status:      &enums.ControlStatusPreparing,
				Tags:        []string{"tag1", "tag2"},
				ControlType: &enums.ControlTypeDetective,
				Category:    lo.ToPtr("Availability"),
				CategoryID:  lo.ToPtr("A"),
				Subcategory: lo.ToPtr("Additional Criteria for Availability"),
				AppendReferences: []*models.Reference{
					{
						Name: "name of ref",
						URL:  "https://example.com",
					},
				},
				AppendMappedCategories: []string{"Govern", "Protect"},
				AppendControlQuestions: []string{"What is the control question?"},
				AppendAssessmentObjectives: []*models.AssessmentObjective{
					{
						Class:     "class",
						ID:        "id",
						Objective: "objective for the control",
					},
				},
				AppendAssessmentMethods: []*models.AssessmentMethod{
					{
						ID:     "id",
						Type:   "Examine",
						Method: "method of assessment for the control",
					},
				},
				AppendImplementationGuidance: []*models.ImplementationGuidance{
					{
						ReferenceID: "ref-id",
						Guidance: []string{
							"guidance 1",
							"guidance 2",
						},
					},
				},
				AppendExampleEvidence: []*models.ExampleEvidence{
					{
						DocumentationType: "policy",
						Description:       "description of the example evidence",
					},
				},
				ControlOwnerID: &ownerGroup.ID,
				DelegateID:     &deleteGroup.ID,
				Source:         &enums.ControlSourceFramework,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, remove some things",
			request: openlaneclient.UpdateControlInput{
				ClearReferences:       lo.ToPtr(true),
				ClearMappedCategories: lo.ToPtr(true),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update ref code to empty",
			request: openlaneclient.UpdateControlInput{
				RefCode: lo.ToPtr(""),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "update not allowed, not permissions in same org",
			request: openlaneclient.UpdateControlInput{
				Status: &enums.ControlStatusPreparing,
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "update allowed, user added to one of the programs",
			request: openlaneclient.UpdateControlInput{
				Status: &enums.ControlStatusPreparing,
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "update not allowed, no permissions",
			request: openlaneclient.UpdateControlInput{
				Status: &enums.ControlStatusPreparing,
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

			if tc.request.Source != nil {
				assert.Equal(t, *tc.request.Source, *resp.UpdateControl.Control.Source)
			}

			if tc.request.ControlType != nil {
				assert.Equal(t, *tc.request.ControlType, *resp.UpdateControl.Control.ControlType)
			}

			if tc.request.Category != nil {
				assert.Equal(t, *tc.request.Category, *resp.UpdateControl.Control.Category)
			}

			if tc.request.CategoryID != nil {
				assert.Equal(t, *tc.request.CategoryID, *resp.UpdateControl.Control.CategoryID)
			}

			if tc.request.Subcategory != nil {
				assert.Equal(t, *tc.request.Subcategory, *resp.UpdateControl.Control.Subcategory)
			}

			if tc.request.AppendMappedCategories != nil {
				assert.ElementsMatch(t, tc.request.AppendMappedCategories, resp.UpdateControl.Control.MappedCategories)
			}

			if tc.request.AppendControlQuestions != nil {
				assert.ElementsMatch(t, tc.request.AppendControlQuestions, resp.UpdateControl.Control.ControlQuestions)
			}

			if tc.request.AppendAssessmentObjectives != nil {
				assert.ElementsMatch(t, tc.request.AppendAssessmentObjectives, resp.UpdateControl.Control.AssessmentObjectives)
			}

			if tc.request.AppendAssessmentMethods != nil {
				assert.ElementsMatch(t, tc.request.AppendAssessmentMethods, resp.UpdateControl.Control.AssessmentMethods)
			}

			if tc.request.AppendImplementationGuidance != nil {
				assert.ElementsMatch(t, tc.request.AppendImplementationGuidance, resp.UpdateControl.Control.ImplementationGuidance)
			}

			if tc.request.AppendExampleEvidence != nil {
				assert.ElementsMatch(t, tc.request.AppendExampleEvidence, resp.UpdateControl.Control.ExampleEvidence)
			}

			if tc.request.ControlOwnerID != nil {
				require.NotNil(t, resp.UpdateControl.Control.ControlOwner)
				assert.Equal(t, *tc.request.ControlOwnerID, resp.UpdateControl.Control.ControlOwner.ID)
			}

			if tc.request.DelegateID != nil {
				require.NotNil(t, resp.UpdateControl.Control.Delegate)
				assert.Equal(t, *tc.request.DelegateID, resp.UpdateControl.Control.Delegate.ID)
			}

			if tc.request.AppendReferences != nil {
				assert.ElementsMatch(t, tc.request.AppendReferences, resp.UpdateControl.Control.References)
			}

			if tc.request.ClearReferences != nil && *tc.request.ClearReferences {
				assert.Empty(t, resp.UpdateControl.Control.References)
			}

			if tc.request.ClearMappedCategories != nil && *tc.request.ClearMappedCategories {
				assert.Empty(t, resp.UpdateControl.Control.MappedCategories)
			}

			// ensure the program is set
			if len(tc.request.AddProgramIDs) > 0 {
				require.NotEmpty(t, resp.UpdateControl.Control.Programs)
				require.Len(t, resp.UpdateControl.Control.Programs.Edges, len(tc.request.AddProgramIDs))
			}

			if len(tc.request.AddViewerIDs) > 0 {
				require.Len(t, resp.UpdateControl.Control.Viewers, 1)
				found := false
				for _, edge := range resp.UpdateControl.Control.Viewers {
					if edge.ID == tc.request.AddViewerIDs[0] {
						found = true
						break
					}
				}

				assert.True(t, found)

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
