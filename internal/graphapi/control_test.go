package graphapi_test

import (
	"cmp"
	"context"
	"slices"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/core/pkg/testutils"
	"github.com/theopenlane/utils/ulids"
)

func TestQueryControl(t *testing.T) {
	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can create a control
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)

	controlIDs := []string{}
	// add test cases for querying the control
	testCases := []struct {
		name          string
		queryID       string
		programAccess bool // whether the user has access to the program
		client        *openlaneclient.OpenlaneClient
		ctx           context.Context
		errorMsg      string
	}{
		{
			name:          "happy path",
			programAccess: true,
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
		},
		{
			name:          "read only user, inherits access from the organization",
			programAccess: false,
			client:        suite.client.api,
			ctx:           viewOnlyUser.UserCtx,
		},
		{
			name:          "admin user, access to the program",
			programAccess: true,
			client:        suite.client.api,
			ctx:           adminUser.UserCtx,
		},
		{
			name:          "happy path using personal access token",
			programAccess: true,
			client:        suite.client.apiWithPAT,
			ctx:           context.Background(),
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
						RefCode:    "CTL-" + ulids.New().String(), // ensure unique ref code
						ProgramIDs: []string{program.ID},
					})

				assert.NilError(t, err)
				assert.Assert(t, resp != nil)

				tc.queryID = resp.CreateControl.Control.ID
				controlIDs = append(controlIDs, tc.queryID)
			}

			resp, err := tc.client.GetControlByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Check(t, resp != nil)

			assert.Equal(t, tc.queryID, resp.Control.ID)
			assert.Check(t, len(resp.Control.RefCode) != 0)

			if tc.programAccess {
				assert.Check(t, is.Len(resp.Control.Programs.Edges, 1))
				assert.Check(t, len(resp.Control.Programs.Edges[0].Node.ID) != 0)
			} else {
				assert.Check(t, is.Len(resp.Control.Programs.Edges, 0))
			}
		})
	}

	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: controlIDs}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: program.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryControls(t *testing.T) {
	// create multiple objects to be queried using testUser1
	controlsToCreate := int64(11)
	controlIDs := []string{}
	for range controlsToCreate { // set to 11 to ensure pagination is tested
		control := (&ControlBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
		controlIDs = append(controlIDs, control.ID)
	}

	userAnotherOrg := suite.userBuilder(context.Background(), t)

	// add a control for the user to another org; this should not be returned for JWT auth, since it's
	// restricted to a single org. PAT auth would return it if both orgs are authorized on the token
	controlAnotherOrg := (&ControlBuilder{client: suite.client}).MustNew(userAnotherOrg.UserCtx, t)

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
			name:            "happy path, admin user",
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
			first:           &controlsToCreate,
			ctx:             testUser1.UserCtx,
			expectedResults: testutils.MaxResultLimit,
		},
		{
			name:            "last set over max (10 in test)",
			client:          suite.client.api,
			last:            &controlsToCreate,
			ctx:             testUser1.UserCtx,
			expectedResults: testutils.MaxResultLimit,
		},
		{
			name:            "happy path, using read only user of the same org should inherit access from the org",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: testutils.MaxResultLimit,
		},
		{
			name:            "happy path, with api token",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: testutils.MaxResultLimit,
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
				assert.NilError(t, err)
				assert.Check(t, resp != nil)

				assert.Check(t, is.Len(resp.Controls.Edges, tc.expectedResults))
				assert.Check(t, is.Equal(controlsToCreate, resp.Controls.TotalCount))

				// if we are pulling the last, there won't be a next page, but there will be a previous page
				if tc.last != nil {
					assert.Check(t, resp.Controls.PageInfo.HasPreviousPage)
				} else {
					assert.Check(t, resp.Controls.PageInfo.HasNextPage)
				}

				return
			}

			resp, err := tc.client.GetAllControls(tc.ctx)
			assert.NilError(t, err)
			assert.Check(t, resp != nil)

			assert.Check(t, is.Len(resp.Controls.Edges, tc.expectedResults))

			if tc.expectedResults > 0 {
				assert.Check(t, is.Equal(int64(controlsToCreate), resp.Controls.TotalCount))
				assert.Check(t, resp.Controls.PageInfo.HasNextPage)
			} else {
				assert.Check(t, is.Len(resp.Controls.Edges, 0))
				assert.Equal(t, int64(0), resp.Controls.TotalCount)
				assert.Check(t, !resp.Controls.PageInfo.HasNextPage)
			}
		})
	}
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: controlIDs}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: controlAnotherOrg.ID}).MustDelete(userAnotherOrg.UserCtx, t)
}

func TestMutationCreateControl(t *testing.T) {
	program1 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	programAnotherUser := (&ProgramBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	ownerGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	delegateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can create a control associated with the program1
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program1.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)

	// create groups to be associated with the control
	blockedGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// create control implementation to be associated with the control
	controlImplementation := (&ControlImplementationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
				RefCode:            "A-2",
				Description:        lo.ToPtr("A description of the Control"),
				Status:             &enums.ControlStatusPreparing,
				Tags:               []string{"tag1", "tag2"},
				ControlType:        &enums.ControlTypeDetective,
				Category:           lo.ToPtr("Availability"),
				CategoryID:         lo.ToPtr("A"),
				Subcategory:        lo.ToPtr("Additional Criteria for Availability"),
				MappedCategories:   []string{"Govern", "Protect"},
				ControlQuestions:   []string{"What is the control question?"},
				ReferenceFramework: lo.ToPtr("MITBenchmark"),
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
				ControlOwnerID:           &ownerGroup.ID,
				DelegateID:               &delegateGroup.ID,
				Source:                   &enums.ControlSourceFramework,
				ProgramIDs:               []string{program1.ID, program2.ID}, // multiple programs
				ControlImplementationIDs: []string{controlImplementation.ID},
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
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Check(t, resp != nil)

			// check required fields
			assert.Check(t, len(resp.CreateControl.Control.ID) != 0)
			assert.Equal(t, tc.request.RefCode, resp.CreateControl.Control.RefCode)

			assert.Check(t, len(resp.CreateControl.Control.DisplayID) != 0)
			assert.Check(t, is.Contains(resp.CreateControl.Control.DisplayID, "CTL-"))

			assert.Check(t, len(resp.CreateControl.Control.RefCode) != 0)
			assert.Equal(t, tc.request.RefCode, resp.CreateControl.Control.RefCode)

			// ensure the program is set
			if len(tc.request.ProgramIDs) > 0 {
				assert.Check(t, is.Len(resp.CreateControl.Control.Programs.Edges, len(tc.request.ProgramIDs)))

				for i, p := range resp.CreateControl.Control.Programs.Edges {
					assert.Equal(t, tc.request.ProgramIDs[i], p.Node.ID)
				}
			} else {
				assert.Check(t, is.Len(resp.CreateControl.Control.Programs.Edges, 0))
			}

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.CreateControl.Control.Description)
			} else {
				assert.Equal(t, *resp.CreateControl.Control.Description, "")
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, *resp.CreateControl.Control.Status)
			} else {
				assert.Equal(t, enums.ControlStatusNotImplemented, *resp.CreateControl.Control.Status)
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
				assert.Equal(t, *resp.CreateControl.Control.Category, "")
			}

			if tc.request.CategoryID != nil {
				assert.Equal(t, *tc.request.CategoryID, *resp.CreateControl.Control.CategoryID)
			} else {
				assert.Equal(t, *resp.CreateControl.Control.CategoryID, "")
			}

			if tc.request.Subcategory != nil {
				assert.Equal(t, *tc.request.Subcategory, *resp.CreateControl.Control.Subcategory)
			} else {
				assert.Equal(t, *resp.CreateControl.Control.Subcategory, "")
			}

			if tc.request.MappedCategories != nil {
				assert.DeepEqual(t, tc.request.MappedCategories, resp.CreateControl.Control.MappedCategories)
			} else {
				assert.Check(t, is.Len(resp.CreateControl.Control.MappedCategories, 0))
			}

			if tc.request.ControlQuestions != nil {
				assert.DeepEqual(t, tc.request.ControlQuestions, resp.CreateControl.Control.ControlQuestions)
			} else {
				assert.Check(t, is.Len(resp.CreateControl.Control.ControlQuestions, 0))
			}

			if tc.request.AssessmentObjectives != nil {
				assert.Check(t, is.Len(resp.CreateControl.Control.AssessmentObjectives, len(tc.request.AssessmentObjectives)))
				assert.DeepEqual(t, tc.request.AssessmentObjectives, resp.CreateControl.Control.AssessmentObjectives)
			} else {
				assert.Check(t, is.Len(resp.CreateControl.Control.AssessmentObjectives, 0))
			}

			if tc.request.AssessmentMethods != nil {
				assert.Check(t, is.Len(resp.CreateControl.Control.AssessmentMethods, len(tc.request.AssessmentMethods)))
				assert.DeepEqual(t, tc.request.AssessmentMethods, resp.CreateControl.Control.AssessmentMethods)
			} else {
				assert.Check(t, is.Len(resp.CreateControl.Control.AssessmentMethods, 0))
			}

			if tc.request.ImplementationGuidance != nil {
				assert.Check(t, is.Len(resp.CreateControl.Control.ImplementationGuidance, len(tc.request.ImplementationGuidance)))
				assert.DeepEqual(t, tc.request.ImplementationGuidance, resp.CreateControl.Control.ImplementationGuidance)
			} else {
				assert.Check(t, is.Len(resp.CreateControl.Control.ImplementationGuidance, 0))
			}

			if tc.request.ExampleEvidence != nil {
				assert.Check(t, is.Len(resp.CreateControl.Control.ExampleEvidence, len(tc.request.ExampleEvidence)))
				assert.DeepEqual(t, tc.request.ExampleEvidence, resp.CreateControl.Control.ExampleEvidence)
			} else {
				assert.Check(t, is.Len(resp.CreateControl.Control.ExampleEvidence, 0))
			}

			if tc.request.References != nil {
				assert.Check(t, is.Len(resp.CreateControl.Control.References, len(tc.request.References)))
				assert.DeepEqual(t, tc.request.References, resp.CreateControl.Control.References)
			} else {
				assert.Check(t, is.Len(resp.CreateControl.Control.References, 0))
			}

			if tc.request.ControlOwnerID != nil {
				assert.Equal(t, *tc.request.ControlOwnerID, resp.CreateControl.Control.ControlOwner.ID)
			} else {
				assert.Check(t, resp.CreateControl.Control.ControlOwner == nil)
			}

			if tc.request.DelegateID != nil {
				assert.Equal(t, *tc.request.DelegateID, resp.CreateControl.Control.Delegate.ID)
			} else {
				assert.Check(t, resp.CreateControl.Control.Delegate == nil)
			}

			if len(tc.request.EditorIDs) > 0 {
				assert.Check(t, is.Len(resp.CreateControl.Control.Editors.Edges, 1))
				for _, edge := range resp.CreateControl.Control.Editors.Edges {
					assert.Equal(t, testUser1.GroupID, edge.Node.ID)
				}
			}

			if len(tc.request.BlockedGroupIDs) > 0 {
				assert.Check(t, is.Len(resp.CreateControl.Control.BlockedGroups.Edges, 1))
				for _, edge := range resp.CreateControl.Control.BlockedGroups.Edges {
					assert.Equal(t, blockedGroup.ID, edge.Node.ID)
				}
			}

			if tc.request.ControlImplementationIDs != nil {
				assert.Assert(t, is.Len(resp.CreateControl.Control.ControlImplementations.Edges, len(tc.request.ControlImplementationIDs)))
			}

			if tc.request.ReferenceFramework != nil {
				assert.Check(t, is.Equal(*tc.request.ReferenceFramework, *resp.CreateControl.Control.ReferenceFramework))
			} else {
				assert.Check(t, resp.CreateControl.Control.ReferenceFramework == nil)
			}

			// ensure the org owner has access to the control that was created by an api token
			if tc.client == suite.client.apiWithToken {
				res, err := suite.client.api.GetControlByID(testUser1.UserCtx, resp.CreateControl.Control.ID)
				assert.NilError(t, err)
				assert.Equal(t, resp.CreateControl.Control.ID, res.Control.ID)

				if tc.request.ControlImplementationIDs != nil {
					assert.Check(t, is.Len(res.Control.ControlImplementations.Edges, len(tc.request.ControlImplementationIDs)))
				}
			}

			// delete the created evidence, update for the token user cases
			if tc.ctx == context.Background() {
				tc.ctx = testUser1.UserCtx
			}

			(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: resp.CreateControl.Control.ID}).MustDelete(tc.ctx, t)
		})
	}

	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, IDs: []string{program1.ID, program2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: programAnotherUser.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.ControlImplementationDeleteOne]{client: suite.client.db.ControlImplementation, ID: controlImplementation.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{ownerGroup.ID, delegateGroup.ID, blockedGroup.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateControlsByClone(t *testing.T) {
	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	(&ProgramMemberBuilder{client: suite.client, ProgramID: program.ID, UserID: viewOnlyUser.ID}).MustNew(testUser1.UserCtx, t)

	programAnotherOrg := (&ProgramBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	publicStandard := (&StandardBuilder{client: suite.client, IsPublic: true}).MustNew(systemAdminUser.UserCtx, t)

	// create standard with controls to clone
	numControls := int64(20)
	controls := []*generated.Control{}
	controlIDs := make([]string, 0, numControls)
	subcontrols := []*generated.Subcontrol{}
	subcontrolIDs := []string{}
	for range numControls {
		control := (&ControlBuilder{client: suite.client, StandardID: publicStandard.ID, AllFields: true}).MustNew(systemAdminUser.UserCtx, t)
		controls = append(controls, control)
		controlIDs = append(controlIDs, control.ID)
		// give them all a subcontrol
		subcontrol := (&SubcontrolBuilder{client: suite.client, ControlID: control.ID}).MustNew(systemAdminUser.UserCtx, t)
		subcontrols = append(subcontrols, subcontrol)
		subcontrolIDs = append(subcontrolIDs, subcontrol.ID)
	}

	// ensure the standard exists and has the correct number of controls for the non-system admin user
	standard, err := suite.client.api.GetStandardByID(testUser2.UserCtx, publicStandard.ID)
	assert.NilError(t, err)
	assert.Assert(t, standard != nil)
	assert.Equal(t, standard.Standard.Controls.TotalCount, numControls)

	// create org owned control in custom standard
	orgStandard := (&StandardBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	orgOwnedControl := (&ControlBuilder{client: suite.client, AllFields: true, StandardID: orgStandard.ID}).MustNew(testUser1.UserCtx, t)

	// sort controls so they are consistent
	slices.SortFunc(controls, func(a, b *generated.Control) int {
		return cmp.Compare(a.RefCode, b.RefCode)
	})

	controlIDsToDelete := []string{}
	subcontrolIDsToDelete := []string{}
	testCases := []struct {
		name               string
		request            openlaneclient.CloneControlInput
		expectedControls   []*generated.Control
		client             *openlaneclient.OpenlaneClient
		ctx                context.Context
		expectedStandard   *string
		expectedNumProgram int
		expectedErr        string
	}{
		{
			name: "happy path, all controls under standard using standard id",
			request: openlaneclient.CloneControlInput{
				StandardID: &publicStandard.ID,
			},
			expectedStandard: lo.ToPtr(publicStandard.ShortName),
			expectedControls: controls,
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
		},
		{
			name: "happy path, all controls under standard",
			request: openlaneclient.CloneControlInput{
				ControlIDs: controlIDs,
			},
			expectedStandard: lo.ToPtr(publicStandard.ShortName),
			expectedControls: controls,
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
		},
		{
			name: "happy path, clone single control, should  be a no-op. because the control already exists",
			request: openlaneclient.CloneControlInput{
				ControlIDs: []string{controls[7].ID},
			},
			expectedControls: []*generated.Control{controls[7]},
			expectedStandard: lo.ToPtr(publicStandard.ShortName),
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
		},
		{
			name: "happy path, all controls under standard with program",
			request: openlaneclient.CloneControlInput{
				ControlIDs: controlIDs,
				ProgramID:  &program.ID,
			},
			expectedControls:   controls,
			expectedStandard:   lo.ToPtr(publicStandard.ShortName),
			expectedNumProgram: 1,
			client:             suite.client.api,
			ctx:                testUser1.UserCtx,
		},
		{
			name: "all controls under standard with program no access",
			request: openlaneclient.CloneControlInput{
				ControlIDs: controlIDs,
				ProgramID:  &programAnotherOrg.ID,
			},
			expectedControls: controls,
			expectedStandard: lo.ToPtr(publicStandard.ShortName),
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
			expectedStandard: nil,
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
		},
		{
			name: "happy path, clone control under org with program",
			request: openlaneclient.CloneControlInput{
				ControlIDs: []string{orgOwnedControl.ID},
				ProgramID:  &program.ID,
			},
			expectedStandard:   nil,
			expectedControls:   []*generated.Control{orgOwnedControl},
			expectedNumProgram: 1,
			client:             suite.client.api,
			ctx:                testUser1.UserCtx,
		},
		{
			name: "happy path, clone single control using personal access token",
			request: openlaneclient.CloneControlInput{
				ControlIDs: []string{controls[:1][0].ID},
				OwnerID:    &testUser1.OrganizationID,
			},
			expectedStandard:   lo.ToPtr(publicStandard.ShortName),
			expectedControls:   controls[:1],
			expectedNumProgram: 1, // control was cloned, so the previous program will still be here
			client:             suite.client.apiWithPAT,
			ctx:                context.Background(),
		},
		{
			name: "clone single control using personal access token, missing owner id",
			request: openlaneclient.CloneControlInput{
				ControlIDs: []string{controls[:1][0].ID},
			},
			expectedStandard: lo.ToPtr(publicStandard.ShortName),
			expectedControls: controls[:1],
			client:           suite.client.apiWithPAT,
			ctx:              context.Background(),
			expectedErr:      "owner_id is required",
		},
		{
			name: "happy path, clone single control using api token",
			request: openlaneclient.CloneControlInput{
				ControlIDs: []string{controls[:1][0].ID},
			},
			expectedStandard:   lo.ToPtr(publicStandard.ShortName),
			expectedControls:   controls[:1],
			expectedNumProgram: 0, // api token has no program access
			client:             suite.client.apiWithToken,
			ctx:                context.Background(),
		},
		{
			name: "clone control under org, no access to control",
			request: openlaneclient.CloneControlInput{
				ControlIDs: []string{orgOwnedControl.ID},
			},
			expectedStandard: lo.ToPtr("Custom"),
			expectedControls: []*generated.Control{orgOwnedControl},
			client:           suite.client.api,
			ctx:              testUser2.UserCtx,
			expectedErr:      notAuthorizedErrorMsg,
		},
		{
			name:             "clone control under org, empty request",
			request:          openlaneclient.CloneControlInput{},
			expectedStandard: lo.ToPtr("Custom"),
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
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				errors := parseClientError(t, err)
				for _, e := range errors {
					if tc.expectedErr == notAuthorizedErrorMsg {
						assertErrorCode(t, e, gqlerrors.UnauthorizedErrorCode)
					}
				}

				return
			}

			assert.NilError(t, err)
			assert.Check(t, resp != nil)

			assert.Check(t, is.Len(resp.CreateControlsByClone.Controls, len(tc.expectedControls)))

			// sort controls so they are consistent
			slices.SortFunc(resp.CreateControlsByClone.Controls, func(a, b *openlaneclient.CreateControlsByClone_CreateControlsByClone_Controls) int {
				return cmp.Compare(a.RefCode, b.RefCode)
			})

			for i, control := range resp.CreateControlsByClone.Controls {
				// check required fields
				assert.Check(t, len(control.ID) != 0)
				assert.Check(t, len(control.DisplayID) != 0)
				assert.Check(t, len(control.RefCode) != 0)

				// all cloned controls should have an owner
				assert.Check(t, control.OwnerID != nil)

				if tc.request.ProgramID != nil {
					assert.Check(t, is.Len(control.Programs.Edges, 1))
					assert.Equal(t, *tc.request.ProgramID, control.Programs.Edges[0].Node.ID)
				} else {
					assert.Check(t, is.Len(control.Programs.Edges, tc.expectedNumProgram))
				}

				// check the cloned control fields are set and match the original control
				assert.Check(t, is.Equal(tc.expectedControls[i].RefCode, control.RefCode))
				assert.Check(t, is.Equal(tc.expectedControls[i].ControlType, *control.ControlType))
				assert.Check(t, is.Equal(tc.expectedControls[i].Category, *control.Category))
				assert.Check(t, is.Equal(tc.expectedControls[i].CategoryID, *control.CategoryID))
				assert.Check(t, is.Equal(tc.expectedControls[i].Subcategory, *control.Subcategory))
				assert.Check(t, is.DeepEqual(tc.expectedControls[i].MappedCategories, control.MappedCategories))
				assert.Check(t, is.DeepEqual(tc.expectedControls[i].ControlQuestions, control.ControlQuestions))
				assert.Check(t, is.DeepEqual(tc.expectedControls[i].Tags, control.Tags))
				assert.Check(t, is.Equal(enums.ControlStatusNotImplemented, *control.Status))
				assert.Check(t, is.Equal(tc.expectedControls[i].ControlType, *control.ControlType))
				assert.Check(t, is.Equal(tc.expectedControls[i].Source, *control.Source))
				assert.Check(t, is.Equal(tc.expectedControls[i].StandardID, *control.StandardID))

				if tc.expectedStandard != nil {
					assert.Check(t, is.Equal(*tc.expectedStandard, *control.ReferenceFramework))
				} else {
					assert.Check(t, control.ReferenceFramework == nil)
				}

				for j, ao := range control.AssessmentObjectives {
					assert.Check(t, is.DeepEqual(tc.expectedControls[i].AssessmentObjectives[j], *ao))
				}

				for j, am := range control.AssessmentMethods {
					assert.Check(t, is.DeepEqual(tc.expectedControls[i].AssessmentMethods[j], *am))
				}

				for j, ig := range control.ImplementationGuidance {
					assert.Check(t, is.DeepEqual(tc.expectedControls[i].ImplementationGuidance[j], *ig))
				}

				for j, ref := range control.References {
					assert.Check(t, is.DeepEqual(tc.expectedControls[i].References[j], *ref))
				}

				for j, ee := range control.ExampleEvidence {
					assert.Check(t, is.DeepEqual(tc.expectedControls[i].ExampleEvidence[j], *ee))
				}

				// ensure the org owner has access to the control that was created by an api token
				if tc.client == suite.client.apiWithToken {
					res, err := suite.client.api.GetControlByID(testUser1.UserCtx, control.ID)
					assert.NilError(t, err)
					assert.Check(t, res != nil)
					assert.Check(t, is.Equal(control.ID, res.Control.ID))
				}

				// ensure view only user can see the control created by the admin user
				res, err := suite.client.api.GetControlByID(viewOnlyUser.UserCtx, control.ID)
				assert.NilError(t, err)
				assert.Check(t, res != nil)
				assert.Check(t, is.Equal(control.ID, res.Control.ID))

				// ensure a user outside my organization cannot get the control
				res, err = suite.client.api.GetControlByID(testUser2.UserCtx, control.ID)
				assert.Check(t, is.Nil(res))

				assert.ErrorContains(t, err, notFoundErrorMsg)

				// delete the created evidence, update for the token user cases
				if tc.ctx == context.Background() {
					tc.ctx = testUser1.UserCtx
				}

				// keep track of controls to delete, sometimes we clone existing controls that were created
				// so we don't want a duplicate delete
				if !slices.Contains(controlIDsToDelete, control.ID) {
					controlIDsToDelete = append(controlIDsToDelete, control.ID)
					if len(control.Subcontrols.Edges) > 0 && !slices.Contains(subcontrolIDsToDelete, control.Subcontrols.Edges[0].Node.ID) {
						subcontrolIDsToDelete = append(subcontrolIDsToDelete, control.Subcontrols.Edges[0].Node.ID)
					}
				}
			}
		})
	}

	// cleanup created controls and standards
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, IDs: subcontrolIDsToDelete}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: controlIDsToDelete}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: publicStandard.ID}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: orgStandard.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: controlIDs}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, IDs: subcontrolIDs}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: program.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateControl(t *testing.T) {
	program1 := (&ProgramBuilder{client: suite.client, EditorIDs: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client, EditorIDs: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	control := (&ControlBuilder{client: suite.client, ProgramID: program1.ID}).MustNew(testUser1.UserCtx, t)

	ownerGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	delegateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// create control implementation to be associated with the control
	controlImplementation := (&ControlImplementationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can update the control
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program1.ID, UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).MustNew(testUser1.UserCtx, t)

	// create another user and add them to the same organization and group as testUser1
	// this will allow us to test the group editor permissions
	anotherViewerUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser1.UserCtx, t, &anotherViewerUser, enums.RoleMember, testUser1.OrganizationID)

	groupMember := (&GroupMemberBuilder{client: suite.client, UserID: anotherViewerUser.ID}).MustNew(testUser1.UserCtx, t)

	// ensure the user does not currently have access to update the control
	res, err := suite.client.api.UpdateControl(anotherViewerUser.UserCtx, control.ID, openlaneclient.UpdateControlInput{
		Status: lo.ToPtr(enums.ControlStatusPreparing),
	})
	assert.ErrorContains(t, err, notAuthorizedErrorMsg)
	assert.Assert(t, is.Nil(res))

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
				AddEditorIDs:  []string{groupMember.GroupID},
				AddControlImplementationIDs: []string{
					controlImplementation.ID,
				},
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
				DelegateID:     &delegateGroup.ID,
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
			expectedErr: notAuthorizedErrorMsg,
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

				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.UpdateControl.Control.Description)
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, *resp.UpdateControl.Control.Status)
			}

			if tc.request.Tags != nil {
				assert.DeepEqual(t, tc.request.Tags, resp.UpdateControl.Control.Tags)
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
				assert.DeepEqual(t, tc.request.AppendMappedCategories, resp.UpdateControl.Control.MappedCategories)
			}

			if tc.request.AppendControlQuestions != nil {
				assert.DeepEqual(t, tc.request.AppendControlQuestions, resp.UpdateControl.Control.ControlQuestions)
			}

			if tc.request.AppendAssessmentObjectives != nil {
				assert.DeepEqual(t, tc.request.AppendAssessmentObjectives, resp.UpdateControl.Control.AssessmentObjectives)
			}

			if tc.request.AppendAssessmentMethods != nil {
				assert.DeepEqual(t, tc.request.AppendAssessmentMethods, resp.UpdateControl.Control.AssessmentMethods)
			}

			if tc.request.AppendImplementationGuidance != nil {
				assert.DeepEqual(t, tc.request.AppendImplementationGuidance, resp.UpdateControl.Control.ImplementationGuidance)
			}

			if tc.request.AppendExampleEvidence != nil {
				assert.DeepEqual(t, tc.request.AppendExampleEvidence, resp.UpdateControl.Control.ExampleEvidence)
			}

			if tc.request.ControlOwnerID != nil {
				assert.Assert(t, resp.UpdateControl.Control.ControlOwner != nil)
				assert.Equal(t, *tc.request.ControlOwnerID, resp.UpdateControl.Control.ControlOwner.ID)
			}

			if tc.request.DelegateID != nil {
				assert.Assert(t, resp.UpdateControl.Control.Delegate != nil)
				assert.Equal(t, *tc.request.DelegateID, resp.UpdateControl.Control.Delegate.ID)
			}

			if tc.request.AppendReferences != nil {
				assert.DeepEqual(t, tc.request.AppendReferences, resp.UpdateControl.Control.References)
			}

			if tc.request.ClearReferences != nil && *tc.request.ClearReferences {
				assert.Check(t, is.Len(resp.UpdateControl.Control.References, 0))
			}

			if tc.request.ClearMappedCategories != nil && *tc.request.ClearMappedCategories {
				assert.Check(t, is.Len(resp.UpdateControl.Control.MappedCategories, 0))
			}

			if tc.request.AddControlImplementationIDs != nil {
				assert.Assert(t, is.Len(resp.UpdateControl.Control.ControlImplementations.Edges, len(tc.request.AddControlImplementationIDs)))
			}

			// ensure the program is set
			if len(tc.request.AddProgramIDs) > 0 {
				assert.Assert(t, is.Len(resp.UpdateControl.Control.Programs.Edges, len(tc.request.AddProgramIDs)))
			}

			if len(tc.request.AddEditorIDs) > 0 {
				assert.Assert(t, is.Len(resp.UpdateControl.Control.Editors.Edges, 1))
				found := false
				for _, edge := range resp.UpdateControl.Control.Editors.Edges {
					if edge.Node.ID == tc.request.AddEditorIDs[0] {
						found = true
						break
					}
				}

				assert.Check(t, found)

				// ensure the user has access to the control now
				res, err := suite.client.api.UpdateControl(anotherViewerUser.UserCtx, control.ID, openlaneclient.UpdateControlInput{
					Tags: []string{"tag1"},
				})
				assert.NilError(t, err)
				assert.Assert(t, res != nil)
				assert.Equal(t, control.ID, res.UpdateControl.Control.ID)
				assert.Check(t, slices.Contains(res.UpdateControl.Control.Tags, "tag1"))
			}
		})
	}

	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, IDs: []string{program1.ID, program2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlImplementationDeleteOne]{client: suite.client.db.ControlImplementation, ID: controlImplementation.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{ownerGroup.ID, delegateGroup.ID, groupMember.GroupID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteControl(t *testing.T) {
	// create objects to be deleted
	control1 := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  control1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: control1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  control1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: control2.ID,
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

				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Equal(t, tc.idToDelete, resp.DeleteControl.DeletedID)
		})
	}
}
