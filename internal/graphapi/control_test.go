package graphapi_test

import (
	"cmp"
	"context"
	"slices"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/internal/testutils"
	"github.com/theopenlane/core/pkg/objects/storage"
)

func TestQueryControl(t *testing.T) {
	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can create a control
	(&ProgramMemberBuilder{
		client: suite.client, ProgramID: program.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String(),
	}).
		MustNew(testUser1.UserCtx, t)
	anonymousContext := createAnonymousTrustCenterContext(ulids.New().String(), testUser1.OrganizationID)

	controlIDs := []string{}
	// add test cases for querying the control
	testCases := []struct {
		name          string
		queryID       string
		programAccess bool // whether the user has access to the program
		client        *testclient.TestClient
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
		{
			name:     "no access, anonymous user",
			client:   suite.client.api,
			ctx:      anonymousContext,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			// setup the control if it is not already created
			if tc.queryID == "" {
				resp, err := suite.client.api.CreateControl(testUser1.UserCtx,
					testclient.CreateControlInput{
						RefCode:    "CTL-" + ulids.New().String(), // ensure unique ref code
						ProgramIDs: []string{program.ID},
						ImplementationGuidance: []*models.ImplementationGuidance{
							{
								ReferenceID: "ref-id-2",
								Guidance: []string{
									"guidance 1",
									"guidance 2",
								},
							},
							{
								ReferenceID: "ref-id-1",
								Guidance: []string{
									"guidance 3",
									"guidance 4",
								},
							},
						},
					})

				assert.NilError(t, err)
				assert.Assert(t, resp != nil)

				tc.queryID = resp.CreateControl.Control.ID
				controlIDs = append(controlIDs, tc.queryID)
			}

			resp, err := tc.client.GetControlByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Check(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.Control.ID))
			assert.Check(t, len(resp.Control.RefCode) != 0)

			// ensure the implementation guidance is sorted
			assert.Check(t, resp.Control.ImplementationGuidance[0].ReferenceID == "ref-id-1")
			assert.Check(t, resp.Control.ImplementationGuidance[1].ReferenceID == "ref-id-2")

			if tc.programAccess {
				assert.Assert(t, resp.Control.Programs.Edges != nil)
				assert.Assert(t, is.Len(resp.Control.Programs.Edges, 1))
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
		client          *testclient.TestClient
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
				resp, err := tc.client.GetControls(tc.ctx, tc.first, tc.last, nil, nil, nil, nil)
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
				assert.Check(t, is.Equal(int64(0), resp.Controls.TotalCount))
				assert.Check(t, !resp.Controls.PageInfo.HasNextPage)
			}
		})
	}
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: controlIDs}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: controlAnotherOrg.ID}).MustDelete(userAnotherOrg.UserCtx, t)
}

func TestQueryControlsMultipleOrgCheck(t *testing.T) {
	// test to make sure we don't get cross org results back even if the user technically has access to them
	testUser := suite.userBuilder(context.Background(), t)

	// create controls for the test user in their org
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	testUserOriginalCtx := auth.NewTestContextWithOrgID(testUser.ID, testUser.OrganizationID)

	// create another org and a control in that org
	anotherOrg := (&OrganizationBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	testUserCtxUpdate := auth.NewTestContextWithOrgID(testUser.ID, anotherOrg.ID)

	controlAnotherOrg := (&ControlBuilder{client: suite.client}).MustNew(testUserCtxUpdate, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		first           *int64
		last            *int64
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             testUserOriginalCtx,
			expectedResults: 1,
		},
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             testUserCtxUpdate,
			expectedResults: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllControls(tc.ctx)
			assert.NilError(t, err)
			assert.Check(t, resp != nil)

			assert.Check(t, is.Len(resp.Controls.Edges, tc.expectedResults))
			assert.Check(t, is.Equal(int64(tc.expectedResults), resp.Controls.TotalCount))
		})
	}
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUserOriginalCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: controlAnotherOrg.ID}).MustDelete(testUserCtxUpdate, t)
}

func TestMutationCreateControl(t *testing.T) {
	program1 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	programAnotherUser := (&ProgramBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	ownerGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	delegateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can create a control associated with the program1
	(&ProgramMemberBuilder{
		client: suite.client, ProgramID: program1.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String(),
	}).
		MustNew(testUser1.UserCtx, t)

	// create groups to be associated with the control
	blockedGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// create control implementation to be associated with the control
	controlImplementation := (&ControlImplementationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	standard := (&StandardBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.CreateControlInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateControlInput{
				RefCode: "A-1",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateControlInput{
				RefCode:          "A-2",
				Description:      lo.ToPtr("A description of the Control"),
				Status:           &enums.ControlStatusPreparing,
				Tags:             []string{"tag1", "tag2"},
				Category:         lo.ToPtr("Availability"),
				CategoryID:       lo.ToPtr("A"),
				Subcategory:      lo.ToPtr("Additional Criteria for Availability"),
				MappedCategories: []string{"Govern", "Protect"},
				ControlQuestions: []string{"What is the control question?"},
				StandardID:       &standard.ID,
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
						ReferenceID: "ref-id-2",
						Guidance: []string{
							"guidance 1",
							"guidance 2",
						},
					},
					{
						ReferenceID: "ref-id-1",
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
			request: testclient.CreateControlInput{
				RefCode:         "A-3",
				EditorIDs:       []string{testUser1.GroupID},
				BlockedGroupIDs: []string{blockedGroup.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateControlInput{
				RefCode: "A-4",
				OwnerID: &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "using api token",
			request: testclient.CreateControlInput{
				RefCode: "A-5",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateControlInput{
				RefCode: "A-6",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user authorized, they were added to the program",
			request: testclient.CreateControlInput{
				RefCode:    "A-7",
				ProgramIDs: []string{program1.ID},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "user authorized, user not authorized to one of the programs",
			request: testclient.CreateControlInput{
				RefCode:    "A-8",
				ProgramIDs: []string{program1.ID, program2.ID},
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "missing required ref code",
			request: testclient.CreateControlInput{
				Description: lo.ToPtr("A description of the Control"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "user not authorized, no permissions to one of the programs",
			request: testclient.CreateControlInput{
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
				return
			}

			assert.NilError(t, err)
			assert.Check(t, resp != nil)

			// check required fields
			assert.Check(t, len(resp.CreateControl.Control.ID) != 0)
			assert.Check(t, is.Equal(tc.request.RefCode, resp.CreateControl.Control.RefCode))

			assert.Check(t, len(resp.CreateControl.Control.DisplayID) != 0)
			assert.Check(t, is.Contains(resp.CreateControl.Control.DisplayID, "CTL-"))

			assert.Check(t, len(resp.CreateControl.Control.RefCode) != 0)
			assert.Check(t, is.Equal(tc.request.RefCode, resp.CreateControl.Control.RefCode))

			// ensure the program is set
			if len(tc.request.ProgramIDs) > 0 {
				assert.Check(t, is.Len(resp.CreateControl.Control.Programs.Edges, len(tc.request.ProgramIDs)))

				for i, p := range resp.CreateControl.Control.Programs.Edges {
					assert.Check(t, is.Equal(tc.request.ProgramIDs[i], p.Node.ID))
				}
			} else {
				assert.Check(t, is.Len(resp.CreateControl.Control.Programs.Edges, 0))
			}

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.CreateControl.Control.Description))
			} else {
				assert.Check(t, is.Equal(*resp.CreateControl.Control.Description, ""))
			}

			if tc.request.Status != nil {
				assert.Check(t, is.Equal(*tc.request.Status, *resp.CreateControl.Control.Status))
			} else {
				assert.Check(t, is.Equal(enums.ControlStatusNotImplemented, *resp.CreateControl.Control.Status))
			}

			if tc.request.Source != nil {
				assert.Check(t, is.Equal(*tc.request.Source, *resp.CreateControl.Control.Source))
			} else {
				assert.Check(t, is.Equal(enums.ControlSourceUserDefined, *resp.CreateControl.Control.Source))
			}

			if tc.request.Category != nil {
				assert.Check(t, is.Equal(*tc.request.Category, *resp.CreateControl.Control.Category))
			} else {
				assert.Check(t, is.Equal(*resp.CreateControl.Control.Category, ""))
			}

			if tc.request.CategoryID != nil {
				assert.Check(t, is.Equal(*tc.request.CategoryID, *resp.CreateControl.Control.CategoryID))
			} else {
				assert.Check(t, is.Equal(*resp.CreateControl.Control.CategoryID, ""))
			}

			if tc.request.Subcategory != nil {
				assert.Check(t, is.Equal(*tc.request.Subcategory, *resp.CreateControl.Control.Subcategory))
			} else {
				assert.Check(t, is.Equal(*resp.CreateControl.Control.Subcategory, ""))
			}

			if tc.request.MappedCategories != nil {
				assert.Check(t, is.DeepEqual(tc.request.MappedCategories, resp.CreateControl.Control.MappedCategories))
			} else {
				assert.Check(t, is.Len(resp.CreateControl.Control.MappedCategories, 0))
			}

			if tc.request.ControlQuestions != nil {
				assert.Check(t, is.DeepEqual(tc.request.ControlQuestions, resp.CreateControl.Control.ControlQuestions))
			} else {
				assert.Check(t, is.Len(resp.CreateControl.Control.ControlQuestions, 0))
			}

			if tc.request.AssessmentObjectives != nil {
				assert.Check(t, is.Len(resp.CreateControl.Control.AssessmentObjectives, len(tc.request.AssessmentObjectives)))
				assert.Check(t, is.DeepEqual(tc.request.AssessmentObjectives, resp.CreateControl.Control.AssessmentObjectives))
			} else {
				assert.Check(t, is.Len(resp.CreateControl.Control.AssessmentObjectives, 0))
			}

			if tc.request.AssessmentMethods != nil {
				assert.Check(t, is.Len(resp.CreateControl.Control.AssessmentMethods, len(tc.request.AssessmentMethods)))
				assert.Check(t, is.DeepEqual(tc.request.AssessmentMethods, resp.CreateControl.Control.AssessmentMethods))
			} else {
				assert.Check(t, is.Len(resp.CreateControl.Control.AssessmentMethods, 0))
			}

			if tc.request.ImplementationGuidance != nil {
				assert.Check(t, is.Len(resp.CreateControl.Control.ImplementationGuidance, len(tc.request.ImplementationGuidance)))
				assert.Check(t, is.DeepEqual(tc.request.ImplementationGuidance, resp.CreateControl.Control.ImplementationGuidance))
			} else {
				assert.Check(t, is.Len(resp.CreateControl.Control.ImplementationGuidance, 0))
			}

			if tc.request.ExampleEvidence != nil {
				assert.Check(t, is.Len(resp.CreateControl.Control.ExampleEvidence, len(tc.request.ExampleEvidence)))
				assert.Check(t, is.DeepEqual(tc.request.ExampleEvidence, resp.CreateControl.Control.ExampleEvidence))
			} else {
				assert.Check(t, is.Len(resp.CreateControl.Control.ExampleEvidence, 0))
			}

			if tc.request.References != nil {
				assert.Check(t, is.Len(resp.CreateControl.Control.References, len(tc.request.References)))
				assert.Check(t, is.DeepEqual(tc.request.References, resp.CreateControl.Control.References))
			} else {
				assert.Check(t, is.Len(resp.CreateControl.Control.References, 0))
			}

			if tc.request.ControlOwnerID != nil {
				assert.Check(t, is.Equal(*tc.request.ControlOwnerID, resp.CreateControl.Control.ControlOwner.ID))
			} else {
				assert.Check(t, resp.CreateControl.Control.ControlOwner == nil)
			}

			if tc.request.DelegateID != nil {
				assert.Check(t, is.Equal(*tc.request.DelegateID, resp.CreateControl.Control.Delegate.ID))
			} else {
				assert.Check(t, resp.CreateControl.Control.Delegate == nil)
			}

			if len(tc.request.EditorIDs) > 0 {
				assert.Check(t, is.Len(resp.CreateControl.Control.Editors.Edges, 1))
				for _, edge := range resp.CreateControl.Control.Editors.Edges {
					assert.Check(t, is.Equal(testUser1.GroupID, edge.Node.ID))
				}
			}

			if len(tc.request.BlockedGroupIDs) > 0 {
				assert.Check(t, is.Len(resp.CreateControl.Control.BlockedGroups.Edges, 1))
				for _, edge := range resp.CreateControl.Control.BlockedGroups.Edges {
					assert.Check(t, is.Equal(blockedGroup.ID, edge.Node.ID))
				}
			}

			if tc.request.ControlImplementationIDs != nil {
				assert.Check(t, is.Len(resp.CreateControl.Control.ControlImplementations.Edges, len(tc.request.ControlImplementationIDs)))
			}

			if tc.request.StandardID != nil {
				assert.Check(t, is.Equal(*resp.CreateControl.Control.ReferenceFramework, standard.ShortName))
				assert.Check(t, is.Equal(*resp.CreateControl.Control.StandardID, standard.ID))
			} else {
				assert.Check(t, resp.CreateControl.Control.ReferenceFramework == nil)
			}

			// ensure the org owner has access to the control that was created by an api token
			if tc.client == suite.client.apiWithToken {
				res, err := suite.client.api.GetControlByID(testUser1.UserCtx, resp.CreateControl.Control.ID)
				assert.NilError(t, err)
				assert.Check(t, is.Equal(resp.CreateControl.Control.ID, res.Control.ID))

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
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: standard.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateControlsByClone(t *testing.T) {
	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	(&ProgramMemberBuilder{client: suite.client, ProgramID: program.ID, UserID: viewOnlyUser.ID}).MustNew(testUser1.UserCtx, t)

	programAnotherOrg := (&ProgramBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	publicStandard := (&StandardBuilder{client: suite.client, IsPublic: true}).MustNew(systemAdminUser.UserCtx, t)
	publicStandard2 := (&StandardBuilder{client: suite.client, IsPublic: true}).MustNew(systemAdminUser.UserCtx, t)

	// create standard with controls to clone
	numControls := int64(20)
	controls := []*generated.Control{}
	controlIDs := make([]string, 0, numControls)
	subcontrols := []*generated.Subcontrol{}
	subcontrolIDs := []string{}
	controls2 := []*generated.Control{}
	controlID2s := make([]string, 0, numControls)
	subcontrols2 := []*generated.Subcontrol{}
	subcontrolID2s := []string{}
	for range numControls {
		control := (&ControlBuilder{client: suite.client, StandardID: publicStandard.ID, AllFields: true}).MustNew(systemAdminUser.UserCtx, t)
		controls = append(controls, control)
		controlIDs = append(controlIDs, control.ID)
		// give them all a subcontrol
		subcontrol := (&SubcontrolBuilder{client: suite.client, ControlID: control.ID}).MustNew(systemAdminUser.UserCtx, t)
		subcontrols = append(subcontrols, subcontrol)
		subcontrolIDs = append(subcontrolIDs, subcontrol.ID)

		control2 := (&ControlBuilder{client: suite.client, StandardID: publicStandard2.ID, AllFields: true}).MustNew(systemAdminUser.UserCtx, t)
		controls2 = append(controls2, control2)
		controlID2s = append(controlID2s, control2.ID)
		// give them all a subcontrol
		subcontrol2 := (&SubcontrolBuilder{client: suite.client, ControlID: control2.ID}).MustNew(systemAdminUser.UserCtx, t)
		subcontrols2 = append(subcontrols2, subcontrol2)
		subcontrolID2s = append(subcontrolID2s, subcontrol2.ID)
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
	slices.SortFunc(controls2, func(a, b *generated.Control) int {
		return cmp.Compare(a.RefCode, b.RefCode)
	})

	controlIDsToDelete := []string{}
	subcontrolIDsToDelete := []string{}
	testCases := []struct {
		name               string
		request            testclient.CloneControlInput
		expectedControls   []*generated.Control
		client             *testclient.TestClient
		ctx                context.Context
		expectedStandard   *string
		expectedNumProgram int
		expectedErr        string
	}{
		{
			name: "happy path, all controls under standard using standard id",
			request: testclient.CloneControlInput{
				StandardID: &publicStandard.ID,
			},
			expectedStandard: lo.ToPtr(publicStandard.ShortName),
			expectedControls: controls,
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
		},
		{
			name: "happy path, all controls under standard using standard id and program set",
			request: testclient.CloneControlInput{
				StandardID: &publicStandard2.ID,
				ProgramID:  &program.ID,
			},
			expectedStandard: lo.ToPtr(publicStandard2.ShortName),
			expectedControls: controls2,
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
		},
		{
			name: "happy path, all controls under standard",
			request: testclient.CloneControlInput{
				ControlIDs: controlIDs,
			},
			expectedStandard: &publicStandard.ShortName,
			expectedControls: controls,
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
		},
		{
			name: "happy path, clone single control, should  be a no-op. because the control already exists",
			request: testclient.CloneControlInput{
				ControlIDs: []string{controls[7].ID},
			},
			expectedControls: []*generated.Control{controls[7]},
			expectedStandard: &publicStandard.ShortName,
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
		},
		{
			name: "happy path, all controls under standard with program",
			request: testclient.CloneControlInput{
				ControlIDs: controlIDs,
				ProgramID:  &program.ID,
			},
			expectedControls:   controls,
			expectedStandard:   &publicStandard.ShortName,
			expectedNumProgram: 1,
			client:             suite.client.api,
			ctx:                testUser1.UserCtx,
		},
		{
			name: "all controls under standard with program no access",
			request: testclient.CloneControlInput{
				ControlIDs: controlIDs,
				ProgramID:  &programAnotherOrg.ID,
			},
			expectedControls: controls,
			expectedStandard: &publicStandard.ShortName,
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			expectedErr:      notAuthorizedErrorMsg,
		},
		{
			name: "happy path, clone control under org",
			request: testclient.CloneControlInput{
				ControlIDs: []string{orgOwnedControl.ID},
			},
			expectedControls: []*generated.Control{orgOwnedControl},
			expectedStandard: &orgStandard.Name,
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
		},
		{
			name: "happy path, clone control under org with program",
			request: testclient.CloneControlInput{
				ControlIDs: []string{orgOwnedControl.ID},
				ProgramID:  &program.ID,
			},
			expectedStandard:   &orgStandard.Name,
			expectedControls:   []*generated.Control{orgOwnedControl},
			expectedNumProgram: 1,
			client:             suite.client.api,
			ctx:                testUser1.UserCtx,
		},
		{
			name: "happy path, clone single control using personal access token",
			request: testclient.CloneControlInput{
				ControlIDs: []string{controls[:1][0].ID},
				OwnerID:    &testUser1.OrganizationID,
			},
			expectedStandard:   &publicStandard.ShortName,
			expectedControls:   controls[:1],
			expectedNumProgram: 1, // control was cloned, so the previous program will still be here
			client:             suite.client.apiWithPAT,
			ctx:                context.Background(),
		},
		{
			name: "happy path, clone single control using api token",
			request: testclient.CloneControlInput{
				ControlIDs: []string{controls[:1][0].ID},
			},
			expectedStandard:   &publicStandard.ShortName,
			expectedControls:   controls[:1],
			expectedNumProgram: 0, // api token has no program access
			client:             suite.client.apiWithToken,
			ctx:                context.Background(),
		},
		{
			name: "clone control under org, no access to control",
			request: testclient.CloneControlInput{
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
			request:          testclient.CloneControlInput{},
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
			slices.SortFunc(resp.CreateControlsByClone.Controls, func(a, b *testclient.CreateControlsByClone_CreateControlsByClone_Controls) int {
				return cmp.Compare(a.RefCode, b.RefCode)
			})

			for i, control := range resp.CreateControlsByClone.Controls {
				// check required fields
				assert.Check(t, len(control.ID) != 0)
				assert.Check(t, len(control.DisplayID) != 0)
				assert.Check(t, len(control.RefCode) != 0)

				// title isn't required, but checking if its set because we set it on the builder and
				// want to ensure its cloned
				assert.Check(t, len(*control.Title) != 0)

				// all cloned controls should have an owner
				assert.Check(t, control.OwnerID != nil)

				if tc.request.ProgramID != nil {
					assert.Check(t, is.Len(control.Programs.Edges, 1))
					assert.Check(t, is.Equal(*tc.request.ProgramID, control.Programs.Edges[0].Node.ID))
				} else {
					assert.Check(t, is.Len(control.Programs.Edges, tc.expectedNumProgram))
				}

				// check the cloned control fields are set and match the original control
				assert.Check(t, is.Equal(tc.expectedControls[i].RefCode, control.RefCode))
				assert.Check(t, is.Equal(tc.expectedControls[i].Category, *control.Category))
				assert.Check(t, is.Equal(tc.expectedControls[i].CategoryID, *control.CategoryID))
				assert.Check(t, is.Equal(tc.expectedControls[i].Subcategory, *control.Subcategory))
				assert.Check(t, is.DeepEqual(tc.expectedControls[i].MappedCategories, control.MappedCategories))
				assert.Check(t, is.DeepEqual(tc.expectedControls[i].ControlQuestions, control.ControlQuestions))
				assert.Check(t, is.DeepEqual(tc.expectedControls[i].Tags, control.Tags))
				assert.Check(t, is.Equal(enums.ControlStatusNotImplemented, *control.Status))
				assert.Check(t, is.Equal(tc.expectedControls[i].Source, *control.Source))
				assert.Check(t, is.Equal(tc.expectedControls[i].StandardID, *control.StandardID))

				if tc.expectedStandard != nil {
					assert.Check(t, is.Equal(*tc.expectedStandard, *control.ReferenceFramework))
					assert.Check(t, control.ReferenceFrameworkRevision != nil)
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
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: orgStandard.ID}).MustDelete(testUser1.UserCtx, t)

	// now we can delete it and the controls under it will be deleted
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, IDs: []string{publicStandard.ID, publicStandard2.ID}}).MustDelete(systemAdminUser.UserCtx, t)

	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, IDs: subcontrolIDs}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, IDs: subcontrolID2s}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: program.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateControlsByCloneCSV(t *testing.T) {
	validFile, err := storage.NewUploadFile("testdata/uploads/clone.csv")
	assert.NilError(t, err)

	missingControlsFile, err := storage.NewUploadFile("testdata/uploads/all_missing_clone.csv")
	assert.NilError(t, err)

	updateControlsFile, err := storage.NewUploadFile("testdata/uploads/update.csv")
	assert.NilError(t, err)

	// create the standard and controls to be cloned
	standard := (&StandardBuilder{client: suite.client, IsPublic: true, Name: "MITB 1987"}).MustNew(systemAdminUser.UserCtx, t)
	control1 := (&ControlBuilder{client: suite.client, StandardID: standard.ID, RefCode: "AA-1", AllFields: true}).MustNew(systemAdminUser.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client, StandardID: standard.ID, RefCode: "AA-2", Aliases: []string{"AA 2", "ALIAS 2"}, AllFields: true}).MustNew(systemAdminUser.UserCtx, t)

	controlsToDelete := []string{}
	implementationsToDelete := []string{}
	testCases := []struct {
		name                  string
		fileInput             graphql.Upload
		client                *testclient.TestClient
		ctx                   context.Context
		expectedCountControls int
		expectedErr           string
	}{
		{
			name: "happy path, clone controls from csv",
			fileInput: graphql.Upload{
				File:        validFile.RawFile,
				Filename:    validFile.OriginalName,
				Size:        validFile.Size,
				ContentType: validFile.ContentType,
			},
			client:                suite.client.api,
			ctx:                   testUser1.UserCtx,
			expectedCountControls: 2,
		},
		{
			name: "update existing controls, no new controls cloned",
			fileInput: graphql.Upload{
				File:        updateControlsFile.RawFile,
				Filename:    updateControlsFile.OriginalName,
				Size:        updateControlsFile.Size,
				ContentType: updateControlsFile.ContentType,
			},
			client:                suite.client.api,
			ctx:                   testUser1.UserCtx,
			expectedCountControls: 2,
		},
		{
			name: "controls missing from system, no controls cloned",
			fileInput: graphql.Upload{
				File:        missingControlsFile.RawFile,
				Filename:    missingControlsFile.OriginalName,
				Size:        missingControlsFile.Size,
				ContentType: missingControlsFile.ContentType,
			},
			client:                suite.client.api,
			ctx:                   testUser1.UserCtx,
			expectedCountControls: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CloneBulkCSVControl(tc.ctx, tc.fileInput)
			assert.NilError(t, err)
			assert.Check(t, resp != nil)

			assert.Check(t, is.Len(resp.CloneBulkCSVControl.Controls, tc.expectedCountControls))

			// sort controls so they are consistent
			slices.SortFunc(resp.CloneBulkCSVControl.Controls, func(a, b *testclient.CloneBulkCSVControl_CloneBulkCSVControl_Controls) int {
				return cmp.Compare(a.RefCode, b.RefCode)
			})

			for _, control := range resp.CloneBulkCSVControl.Controls {
				assert.Check(t, len(control.ID) != 0)
				assert.Check(t, len(control.DisplayID) != 0)
				assert.Check(t, len(control.RefCode) != 0)
				assert.Check(t, len(*control.Title) != 0)

				switch control.RefCode {
				case "AA-1":
					if tc.fileInput.Filename == updateControlsFile.OriginalName {
						assert.Check(t, is.Equal(enums.ControlStatusApproved, *control.Status))
					} else {
						assert.Check(t, is.Equal(enums.ControlStatusPreparing, *control.Status))
					}
					assert.Check(t, is.Equal("INT-0001", *control.ReferenceID))
				case "AA-2":
					assert.Check(t, is.Equal(enums.ControlStatusApproved, *control.Status))
					assert.Check(t, is.Equal("INT-0002", *control.ReferenceID))

				}

				assert.Check(t, control.ImplementationGuidance != nil)

				switch tc.fileInput.Filename {
				case updateControlsFile.OriginalName:
					assert.Check(t, len(control.ControlImplementations.Edges) == 2)
				case validFile.OriginalName:
					assert.Check(t, len(control.ControlImplementations.Edges) == 1)
				}

				controlsToDelete = append(controlsToDelete, control.ID)
				implementationsToDelete = append(implementationsToDelete, control.ControlImplementations.Edges[0].Node.ID)
			}
		})
	}

	controlsToDelete = lo.Uniq(controlsToDelete)
	implementationsToDelete = lo.Uniq(implementationsToDelete)

	// cleanup controls
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: controlsToDelete}).MustDelete(testUser1.UserCtx, t)
	// cleanup control implementation
	(&Cleanup[*generated.ControlImplementationDeleteOne]{client: suite.client.db.ControlImplementation, IDs: implementationsToDelete}).MustDelete(testUser1.UserCtx, t)
	// cleanup created controls and standards
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control1.ID, control2.ID}}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, IDs: []string{standard.ID}}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationCreateControlsByCloneWithFilter(t *testing.T) {
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

	// sort controls so they are consistent
	slices.SortFunc(controls, func(a, b *generated.Control) int {
		return cmp.Compare(a.RefCode, b.RefCode)
	})

	controlIDsToDelete := []string{}
	subcontrolIDsToDelete := []string{}
	testCases := []struct {
		name                 string
		request              testclient.CloneControlInput
		expectedControlCount int
		client               *testclient.TestClient
		ctx                  context.Context
		expectedErr          string
	}{
		{
			name: "happy path, filter by ref codes",
			request: testclient.CloneControlInput{
				StandardID: &publicStandard.ID,
				RefCodes:   []string{controls[0].RefCode, controls[1].RefCode, controls[2].RefCode},
			},
			expectedControlCount: 3,
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
		},
		{
			name: "happy path, filter by ref codes and aliases",
			request: testclient.CloneControlInput{
				StandardID: &publicStandard.ID,
				RefCodes:   []string{controls[0].RefCode, controls[1].Aliases[1]},
			},
			expectedControlCount: 2,
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
		},
		{
			name: "happy path, filter by aliases",
			request: testclient.CloneControlInput{
				StandardID: &publicStandard.ID,
				RefCodes:   []string{controls[2].Aliases[1], controls[1].Aliases[0]},
			},
			expectedControlCount: 2,
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
		},
		{
			name: "happy path, filter by categories",
			request: testclient.CloneControlInput{
				StandardID: &publicStandard.ID,
				Categories: []string{controls[0].Category, controls[1].Category, controls[2].Category},
			},
			expectedControlCount: 3,
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Create By Clone With Filter "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateControlsByClone(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
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

			assert.Check(t, is.Len(resp.CreateControlsByClone.Controls, tc.expectedControlCount))

			// delete the created controls
			for _, control := range resp.CreateControlsByClone.Controls {
				(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(tc.ctx, t)
			}

			// delete the subcontrols
			for _, control := range resp.CreateControlsByClone.Controls {
				for _, sc := range control.Subcontrols.Edges {
					(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, ID: sc.Node.ID}).MustDelete(tc.ctx, t)
				}
			}
		})
	}

	// cleanup created controls and standards
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, IDs: subcontrolIDsToDelete}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: controlIDsToDelete}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, IDs: subcontrolIDs}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationUpdateFrameworkControl(t *testing.T) {
	publicStandard := (&StandardBuilder{client: suite.client, IsPublic: true}).MustNew(systemAdminUser.UserCtx, t)

	// create standard with controls to clone
	numControls := int64(1)
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

	// sort controls so they are consistent
	slices.SortFunc(controls, func(a, b *generated.Control) int {
		return cmp.Compare(a.RefCode, b.RefCode)
	})

	controlIDsToDelete := []string{}
	subcontrolIDsToDelete := []string{}

	resp, err := suite.client.api.CreateControlsByClone(testUser1.UserCtx, testclient.CloneControlInput{
		StandardID: &publicStandard.ID,
	})
	assert.NilError(t, err)
	assert.Check(t, resp != nil)

	for _, control := range resp.CreateControlsByClone.Controls {
		controlIDsToDelete = append(controlIDsToDelete, control.ID)
		if len(control.Subcontrols.Edges) > 0 {
			subcontrolIDsToDelete = append(subcontrolIDsToDelete, control.Subcontrols.Edges[0].Node.ID)
		}
	}

	testCases := []struct {
		name        string
		request     testclient.UpdateControlInput
		controlID   string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update allowed fields",
			request: testclient.UpdateControlInput{
				Status:           &enums.ControlStatusApproved,
				MappedCategories: []string{"New Category"},
			},
			controlID: resp.CreateControlsByClone.Controls[0].ID,
			client:    suite.client.api,
			ctx:       testUser1.UserCtx,
		},
		{
			name: "fail, unable to update ref code",
			request: testclient.UpdateControlInput{
				RefCode: lo.ToPtr("NEW-REF-CODE"),
			},
			controlID:   resp.CreateControlsByClone.Controls[0].ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: invalidInputErrorMsg,
		},
		{
			name: "fail, unable to update description",
			request: testclient.UpdateControlInput{
				Description: lo.ToPtr("New Description"),
			},
			controlID:   resp.CreateControlsByClone.Controls[0].ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: invalidInputErrorMsg,
		},
		{
			name: "fail, unable to update title",
			request: testclient.UpdateControlInput{
				Title: lo.ToPtr("New Title"),
			},
			controlID:   resp.CreateControlsByClone.Controls[0].ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: invalidInputErrorMsg,
		},
		{
			name: "fail, unable to update source",
			request: testclient.UpdateControlInput{
				Source: &enums.ControlSourceUserDefined,
			},
			controlID:   resp.CreateControlsByClone.Controls[0].ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: invalidInputErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update framework based control"+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateControl(tc.ctx, tc.controlID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Check(t, resp != nil)
		})
	}

	// cleanup created controls and standards
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, IDs: subcontrolIDsToDelete}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: controlIDsToDelete}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, IDs: subcontrolIDs}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationUpdateControl(t *testing.T) {
	program1 := (&ProgramBuilder{client: suite.client, EditorIDs: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client, EditorIDs: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	control := (&ControlBuilder{client: suite.client, ProgramID: program1.ID}).MustNew(testUser1.UserCtx, t)
	subcontrol := (&SubcontrolBuilder{client: suite.client, ControlID: control.ID}).MustNew(testUser1.UserCtx, t)

	controlAnotherOrg := (&ControlBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

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
	_, err := suite.client.api.UpdateControl(anotherViewerUser.UserCtx, control.ID, testclient.UpdateControlInput{
		Status: lo.ToPtr(enums.ControlStatusPreparing),
	})
	assert.ErrorContains(t, err, notAuthorizedErrorMsg)

	// create system owned control kind
	kind := (&CustomTypeEnumBuilder{client: suite.client, Name: "Detective", ObjectType: "control"}).MustNew(systemAdminUser.UserCtx, t)
	kindCustom := (&CustomTypeEnumBuilder{client: suite.client, Name: "Custom Control Kind", ObjectType: "control"}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.UpdateControlInput
		controlID   string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: testclient.UpdateControlInput{
				Description:     lo.ToPtr("Updated description"),
				ControlKindName: &kindCustom.Name,
				AddProgramIDs:   []string{program1.ID, program2.ID}, // add multiple programs (one already associated)
				AddEditorIDs:    []string{groupMember.GroupID},
				AddControlImplementationIDs: []string{
					controlImplementation.ID,
				},
			},
			controlID: control.ID,
			client:    suite.client.api,
			ctx:       testUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: testclient.UpdateControlInput{
				Status:          &enums.ControlStatusPreparing,
				Tags:            []string{"tag1", "tag2"},
				ControlKindName: &kind.Name,
				Category:        lo.ToPtr("Availability"),
				CategoryID:      lo.ToPtr("A"),
				Subcategory:     lo.ToPtr("Additional Criteria for Availability"),
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
			},
			controlID: control.ID,
			client:    suite.client.apiWithPAT,
			ctx:       context.Background(),
		},
		{
			name: "happy path, remove some things",
			request: testclient.UpdateControlInput{
				ClearReferences:       lo.ToPtr(true),
				ClearMappedCategories: lo.ToPtr(true),
			},
			controlID: control.ID,
			client:    suite.client.apiWithPAT,
			ctx:       context.Background(),
		},
		{
			name: "invalid custom control enum for control kind",
			request: testclient.UpdateControlInput{
				ControlKindName: lo.ToPtr("InvalidKind"),
			},
			controlID:   control.ID,
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			expectedErr: "value does not exist:",
		},
		{
			name: "update ref code to empty",
			request: testclient.UpdateControlInput{
				RefCode: lo.ToPtr(""),
			},
			controlID:   control.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "update not allowed, not permissions in same org",
			request: testclient.UpdateControlInput{
				Status: &enums.ControlStatusPreparing,
			},
			controlID:   control.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, cannot access another org's custom control kind",
			request: testclient.UpdateControlInput{
				ControlKindName: &kindCustom.Name,
			},
			controlID:   controlAnotherOrg.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: "value does not exist:",
		},
		{
			name: "update allowed, user added to one of the programs",
			request: testclient.UpdateControlInput{
				Status: &enums.ControlStatusPreparing,
			},
			controlID: control.ID,
			client:    suite.client.api,
			ctx:       adminUser.UserCtx,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateControlInput{
				Status: &enums.ControlStatusPreparing,
			},
			controlID:   control.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateControl(tc.ctx, tc.controlID, tc.request)
			if tc.expectedErr != "" {

				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.UpdateControl.Control.Description))
			}

			if tc.request.Status != nil {
				assert.Check(t, is.Equal(*tc.request.Status, *resp.UpdateControl.Control.Status))
			}

			if tc.request.Tags != nil {
				assert.Check(t, is.DeepEqual(tc.request.Tags, resp.UpdateControl.Control.Tags))
			}

			if tc.request.Source != nil {
				assert.Check(t, is.Equal(*tc.request.Source, *resp.UpdateControl.Control.Source))
			}

			if tc.request.ControlKindName != nil {
				assert.Check(t, resp.UpdateControl.Control.ControlKindName != nil)
				assert.Check(t, is.Equal(*tc.request.ControlKindName, *resp.UpdateControl.Control.ControlKindName))
			}

			if tc.request.Category != nil {
				assert.Check(t, is.Equal(*tc.request.Category, *resp.UpdateControl.Control.Category))
			}

			if tc.request.CategoryID != nil {
				assert.Check(t, is.Equal(*tc.request.CategoryID, *resp.UpdateControl.Control.CategoryID))
			}

			if tc.request.Subcategory != nil {
				assert.Check(t, is.Equal(*tc.request.Subcategory, *resp.UpdateControl.Control.Subcategory))
			}

			if tc.request.AppendMappedCategories != nil {
				assert.Check(t, is.DeepEqual(tc.request.AppendMappedCategories, resp.UpdateControl.Control.MappedCategories))
			}

			if tc.request.AppendControlQuestions != nil {
				assert.Check(t, is.DeepEqual(tc.request.AppendControlQuestions, resp.UpdateControl.Control.ControlQuestions))
			}

			if tc.request.AppendAssessmentObjectives != nil {
				assert.Check(t, is.DeepEqual(tc.request.AppendAssessmentObjectives, resp.UpdateControl.Control.AssessmentObjectives))
			}

			if tc.request.AppendAssessmentMethods != nil {
				assert.Check(t, is.DeepEqual(tc.request.AppendAssessmentMethods, resp.UpdateControl.Control.AssessmentMethods))
			}

			if tc.request.AppendImplementationGuidance != nil {
				assert.Check(t, is.DeepEqual(tc.request.AppendImplementationGuidance, resp.UpdateControl.Control.ImplementationGuidance))
			}

			if tc.request.AppendExampleEvidence != nil {
				assert.Check(t, is.DeepEqual(tc.request.AppendExampleEvidence, resp.UpdateControl.Control.ExampleEvidence))
			}

			if tc.request.ControlOwnerID != nil {
				assert.Check(t, resp.UpdateControl.Control.ControlOwner != nil)
				assert.Check(t, is.Equal(*tc.request.ControlOwnerID, resp.UpdateControl.Control.ControlOwner.ID))
			}

			if tc.request.DelegateID != nil {
				assert.Check(t, resp.UpdateControl.Control.Delegate != nil)
				assert.Check(t, is.Equal(*tc.request.DelegateID, resp.UpdateControl.Control.Delegate.ID))
			}

			if tc.request.AppendReferences != nil {
				assert.Check(t, is.DeepEqual(tc.request.AppendReferences, resp.UpdateControl.Control.References))
			}

			if tc.request.ClearReferences != nil && *tc.request.ClearReferences {
				assert.Check(t, is.Len(resp.UpdateControl.Control.References, 0))
			}

			if tc.request.ClearMappedCategories != nil && *tc.request.ClearMappedCategories {
				assert.Check(t, is.Len(resp.UpdateControl.Control.MappedCategories, 0))
			}

			if tc.request.AddControlImplementationIDs != nil {
				assert.Check(t, is.Len(resp.UpdateControl.Control.ControlImplementations.Edges, len(tc.request.AddControlImplementationIDs)))
			}

			if tc.request.ClearStandard != nil && *tc.request.ClearStandard {
				assert.Check(t, resp.UpdateControl.Control.ReferenceFramework == nil)
				assert.Check(t, resp.UpdateControl.Control.Subcontrols.Edges[0].Node.ReferenceFramework == nil)
			}

			// ensure the program is set
			if len(tc.request.AddProgramIDs) > 0 {
				assert.Check(t, is.Len(resp.UpdateControl.Control.Programs.Edges, len(tc.request.AddProgramIDs)))
			}

			if len(tc.request.AddEditorIDs) > 0 {
				assert.Check(t, is.Len(resp.UpdateControl.Control.Editors.Edges, 1))
				found := false
				for _, edge := range resp.UpdateControl.Control.Editors.Edges {
					if edge.Node.ID == tc.request.AddEditorIDs[0] {
						found = true
						break
					}
				}

				assert.Check(t, found)

				// ensure the user has access to the control now
				res, err := suite.client.api.UpdateControl(anotherViewerUser.UserCtx, control.ID, testclient.UpdateControlInput{
					Tags: []string{"tag1"},
				})
				assert.NilError(t, err)
				assert.Check(t, res != nil)
				assert.Check(t, is.Equal(control.ID, res.UpdateControl.Control.ID))
				assert.Check(t, slices.Contains(res.UpdateControl.Control.Tags, "tag1"))
			}
		})
	}
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, ID: subcontrol.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, IDs: []string{program1.ID, program2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlImplementationDeleteOne]{client: suite.client.db.ControlImplementation, ID: controlImplementation.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{ownerGroup.ID, delegateGroup.ID, groupMember.GroupID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: controlAnotherOrg.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.CustomTypeEnumDeleteOne]{client: suite.client.db.CustomTypeEnum, IDs: []string{kind.ID}}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.CustomTypeEnumDeleteOne]{client: suite.client.db.CustomTypeEnum, IDs: []string{kindCustom.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateControlDescription(t *testing.T) {
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	descriptionRichText := "<div class=\"slate-editor group/editor relative w-full cursor-text overflow-x-hidden break-words whitespace-pre-wrap select-text rounded-md ring-offset-background focus-visible:outline-hidden placeholder:text-muted-foreground/80 **:data-slate-placeholder:top-[auto_!important] **:data-slate-placeholder:text-muted-foreground/80 **:data-slate-placeholder:opacity-100! [&_strong]:font-bold\" data-slate-editor=\"true\" data-slate-node=\"value\"><div data-slate-node=\"element\" data-block-id=\"kK9tZ5Tllq\" data-slate-type=\"p\" data-slate-id=\"kK9tZ5Tllq\" class=\"slate-p m-0 px-0 py-1\" style=\"position:relative\"><span data-slate-node=\"text\"><span data-slate-leaf=\"true\"><span data-slate-string=\"true\">lets just have text here and see</span></span></span></div></div>"
	descriptionSlateJSON := []any{`
        {
          "children": [
            {
              "text": "lets just have text here and see but I added something else here so thats my fault"
            }
          ],
          "id": "kK9tZ5Tllq",
          "type": "p"
        }
      `}

	descriptionRichUpdatedText := "<div class=\"slate-editor group/editor relative w-full cursor-text overflow-x-hidden break-words whitespace-pre-wrap select-text rounded-md ring-offset-background focus-visible:outline-hidden placeholder:text-muted-foreground/80 **:data-slate-placeholder:top-[auto_!important] **:data-slate-placeholder:text-muted-foreground/80 **:data-slate-placeholder:opacity-100! [&_strong]:font-bold\" data-slate-editor=\"true\" data-slate-node=\"value\"><div data-slate-node=\"element\" data-block-id=\"kK9tZ5Tllq\" data-slate-type=\"p\" data-slate-id=\"kK9tZ5Tllq\" class=\"slate-p m-0 px-0 py-1\" style=\"position:relative\"><span data-slate-node=\"text\"><span data-slate-leaf=\"true\"><span data-slate-string=\"true\">lets just have text here and see with update</span></span></span></div></div>"

	descriptionRichTextForComments := "<div class=\"slate-editor group/editor relative w-full cursor-text overflow-x-hidden break-words whitespace-pre-wrap select-text rounded-md ring-offset-background focus-visible:outline-hidden placeholder:text-muted-foreground/80 **:data-slate-placeholder:top-[auto_!important] **:data-slate-placeholder:text-muted-foreground/80 **:data-slate-placeholder:opacity-100! [&_strong]:font-bold\" data-slate-editor=\"true\" data-slate-node=\"value\"><div data-slate-node=\"element\" data-block-id=\"b_bwtnb9l8\" data-slate-id=\"b_bwtnb9l8\" data-slate-type=\"p\" class=\"slate-p m-0 px-0 py-1\" style=\"position:relative\"><span data-slate-node=\"text\"><span data-slate-leaf=\"true\"><span data-slate-string=\"true\">asfsadfsd</span></span></span></div><div data-slate-node=\"element\" data-block-id=\"lqbGHj_l70\" data-slate-id=\"lqbGHj_l70\" data-slate-type=\"p\" class=\"slate-p m-0 px-0 py-1\" style=\"position:relative\"><span data-slate-node=\"text\"><span data-slate-leaf=\"true\"><span data-slate-string=\"true\"></span></span></span></div><div data-slate-node=\"element\" data-block-id=\"qfPeKFLe13\" data-slate-id=\"qfPeKFLe13\" data-slate-type=\"p\" class=\"slate-p m-0 px-0 py-1\" style=\"position:relative\"><span data-slate-node=\"text\"><span data-slate-leaf=\"true\"><span data-slate-string=\"true\">for a comment</span></span></span></div></div>"

	jsonWithComments := []any{
		`{
              "children": [
                {
                  "text": "asfsadfsd"
                }
              ],
              "id": "b_bwtnb9l8",
              "type": "p"
            }`,
		`{
              "children": [
                {
                  "text": ""
                }
              ],
              "id": "lqbGHj_l70",
              "type": "p"
            }`,
		`{
              "children": [
                {
                  "text": "for a "
                },
                {
                  "comment": true,
                  "comment_MDHGnHfbfTfX-amk1Gugp": true,
                  "text": "comment"
                }
              ],
              "id": "qfPeKFLe13",
              "type": "p"
            }`,
	}

	testCases := []struct {
		name         string
		request      testclient.UpdateControlInput
		controlID    string
		client       *testclient.TestClient
		ctx          context.Context
		expectedErr  string
		expectedJSON []any
	}{
		{
			name: "happy path, update both fields, it should return description JSON provided (even if description text matches the JSON)",
			request: testclient.UpdateControlInput{
				Description:     &descriptionRichText,
				DescriptionJSON: descriptionSlateJSON,
			},
			controlID:    control.ID,
			client:       suite.client.api,
			ctx:          testUser1.UserCtx,
			expectedJSON: descriptionSlateJSON,
		},
		{
			name: "happy path, update description rich text, don't provide description JSON",
			request: testclient.UpdateControlInput{
				Description: &descriptionRichUpdatedText,
			},
			controlID: control.ID,
			client:    suite.client.api,
			ctx:       testUser1.UserCtx,
		},
		{
			name: "complex slate JSON with comments and multiple blocks",
			request: testclient.UpdateControlInput{
				Description: lo.ToPtr(`<div class="slate-editor group/editor relative w-full cursor-text overflow-x-hidden break-words whitespace-pre-wrap select-text rounded-md ring-offset-background focus-visible:outline-hidden placeholder:text-muted-foreground/80 **:data-slate-placeholder:top-[auto_!important] **:data-slate-placeholder:text-muted-foreground/80 **:data-slate-placeholder:opacity-100! [&_strong]:font-bold" data-slate-editor="true" data-slate-node="value"><div data-slate-node="element" data-block-id="b_bwtnb9l8" data-slate-id="b_bwtnb9l8" data-slate-type="p" class="slate-p m-0 px-0 py-1" style="position:relative"><span data-slate-node="text"><span data-slate-leaf="true"><span data-slate-string="true">asfsadfsd</span></span></span></div><div data-slate-node="element" data-block-id="lqbGHj_l70" data-slate-id="lqbGHj_l70" data-slate-type="p" class="slate-p m-0 px-0 py-1" style="position:relative"><span data-slate-node="text"><span data-slate-leaf="true"><span data-slate-string="true"></span></span></span></div><div data-slate-node="element" data-block-id="qfPeKFLe13" data-slate-id="qfPeKFLe13" data-slate-type="p" class="slate-p m-0 px-0 py-1" style="position:relative"><span data-slate-node="text"><span data-slate-leaf="true"><span data-slate-string="true">for a comment with update</span></span></span></div></div>`),
			},
			controlID: control.ID,
			client:    suite.client.api,
			ctx:       testUser1.UserCtx,
		},
		{
			name: "complex with comments, provide JSON",
			request: testclient.UpdateControlInput{
				Description:     &descriptionRichTextForComments,
				DescriptionJSON: jsonWithComments,
			},
			controlID:    control.ID,
			client:       suite.client.api,
			ctx:          testUser1.UserCtx,
			expectedJSON: jsonWithComments,
		},
		{
			name: "complex with comments, json not provided, should error",
			request: testclient.UpdateControlInput{
				Description: &descriptionRichTextForComments,
			},
			controlID:   control.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "text contains comments",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateControl(tc.ctx, tc.controlID, tc.request)
			if tc.expectedErr != "" {

				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(*tc.request.Description, *resp.UpdateControl.Control.Description))

			if tc.expectedJSON == nil {
				assert.Check(t, resp.UpdateControl.Control.DescriptionJSON == nil)
			} else {
				assert.Check(t, is.DeepEqual(tc.expectedJSON, resp.UpdateControl.Control.DescriptionJSON))
			}
		})
	}

	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUser1.UserCtx, t)

}

func TestMutationDeleteControl(t *testing.T) {
	// create objects to be deleted
	control1 := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	controlSystem := (&ControlBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
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
			name:        "not authorized, delete system owned control",
			idToDelete:  controlSystem.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
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
			name:       "happy path, delete system owned control",
			idToDelete: controlSystem.ID,
			client:     suite.client.api,
			ctx:        systemAdminUser.UserCtx,
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
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteControl.DeletedID))
		})
	}
}

func TestMutationDeleteBulkControl(t *testing.T) {
	testUser := suite.userBuilder(context.Background(), t)
	anotherUser := suite.userBuilder(context.Background(), t)

	// create objects to be deleted
	control1 := (&ControlBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	control3 := (&ControlBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	controlAnotherUser := (&ControlBuilder{client: suite.client}).MustNew(anotherUser.UserCtx, t)

	testCases := []struct {
		name                 string
		idsToDelete          []string
		client               *testclient.TestClient
		ctx                  context.Context
		expectedErr          string
		expectedDeletedCount int
	}{
		{
			name:                 "happy path, delete multiple controls",
			idsToDelete:          []string{control1.ID, control2.ID, control3.ID},
			client:               suite.client.api,
			ctx:                  testUser.UserCtx,
			expectedDeletedCount: 3,
		},
		{
			name:                 "not authorized, delete controls from another user",
			idsToDelete:          []string{control1.ID, controlAnotherUser.ID},
			client:               suite.client.api,
			ctx:                  anotherUser.UserCtx,
			expectedDeletedCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("Bulk Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteBulkControl(tc.ctx, tc.idsToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Len(resp.DeleteBulkControl.DeletedIDs, tc.expectedDeletedCount))

			// verify that the returned IDs match the ones that were actually deleted
			for _, deletedID := range resp.DeleteBulkControl.DeletedIDs {
				found := false
				for _, expectedID := range tc.idsToDelete {
					if expectedID == deletedID {
						found = true
						break
					}
				}
				assert.Check(t, found, "Deleted ID %s should be in the original request", deletedID)
			}

			// verify that the controls are actually deleted by trying to query them
			for _, deletedID := range resp.DeleteBulkControl.DeletedIDs {
				_, err := tc.client.GetControlByID(tc.ctx, deletedID)
				assert.ErrorContains(t, err, notFoundErrorMsg)
			}
		})
	}
}

func TestQueryControlCategories(t *testing.T) {
	newUser := suite.userBuilder(context.Background(), t)
	newUser2 := suite.userBuilder(context.Background(), t)

	// create controls with categories and subcategories
	control1 := (&ControlBuilder{client: suite.client, AllFields: true}).MustNew(newUser.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client, AllFields: true}).MustNew(newUser.UserCtx, t)

	// create one without a category
	control3 := (&ControlBuilder{client: suite.client}).MustNew(newUser.UserCtx, t)

	// create one with a duplicate category
	control4 := (&ControlBuilder{client: suite.client, Category: control1.Category}).MustNew(newUser.UserCtx, t)

	// create a subcontrol with a different category
	subcontrol := (&SubcontrolBuilder{client: suite.client, ControlID: control1.ID, Category: "New Category"}).MustNew(newUser.UserCtx, t)

	testCases := []struct {
		name           string
		client         *testclient.TestClient
		ctx            context.Context
		expectedErr    string
		expectedResult []string
	}{
		{
			name:           "happy path, get control categories",
			client:         suite.client.api,
			ctx:            newUser.UserCtx,
			expectedResult: []string{control1.Category, control2.Category, subcontrol.Category},
		},
		{
			name:           "no controls, no results",
			client:         suite.client.api,
			ctx:            newUser2.UserCtx,
			expectedResult: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run("Query Control "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetControlCategories(tc.ctx)
			if tc.expectedErr != "" {

				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// sort the categories so they are consistent
			slices.Sort(tc.expectedResult)

			for _, expected := range tc.expectedResult {
				found := false
				for _, category := range resp.ControlCategories {
					if category == expected {
						found = true
						break
					}

					// check for empty categories
					assert.Check(t, category != "")
				}

				assert.Check(t, found)
			}
		})
	}

	// cleanup created controls
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control1.ID, control2.ID, control3.ID, control4.ID}}).
		MustDelete(newUser.UserCtx, t)
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, ID: subcontrol.ID}).MustDelete(newUser.UserCtx, t)
}

// TestQueryControlSubcategories tests the query for control subcategories
// Note: this test will pull all categories, even if the controls weren't created in this test, or in this organization (E.g. public standards)
// will affect the results of this test
// never try to run this in parallel with other tests that create controls
// or standards, or that have controls linked to them
func TestQueryControlSubcategories(t *testing.T) {
	// create controls with categories and subcategories
	control1 := (&ControlBuilder{client: suite.client, AllFields: true}).MustNew(testUser1.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client, AllFields: true}).MustNew(testUser1.UserCtx, t)

	// create one without a subcategory
	control3 := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// create one with a duplicate subcategory
	control4 := (&ControlBuilder{client: suite.client, Subcategory: control1.Subcategory}).MustNew(testUser1.UserCtx, t)

	// create a subcontrol with a different category
	subcontrol := (&SubcontrolBuilder{client: suite.client, ControlID: control1.ID, Subcategory: "New Subcategory"}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name           string
		client         *testclient.TestClient
		ctx            context.Context
		expectedErr    string
		expectedResult []string
	}{
		{
			name:           "happy path, get control subcategories",
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
			expectedResult: []string{control1.Subcategory, control2.Subcategory, subcontrol.Subcategory},
		},
		{
			name:           "no controls, no results",
			client:         suite.client.api,
			ctx:            testUser2.UserCtx,
			expectedResult: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run("Query Control Subcategories "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetControlSubcategories(tc.ctx)
			if tc.expectedErr != "" {

				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// sort the categories so they are consistent
			slices.Sort(tc.expectedResult)

			for _, expected := range tc.expectedResult {
				found := false
				for _, category := range resp.ControlSubcategories {
					if category == expected {
						found = true
						break
					}

					// check for empty categories
					assert.Check(t, category != "")
				}

				assert.Check(t, found)
			}
		})
	}

	// cleanup created controls
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control1.ID, control2.ID, control3.ID, control4.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, ID: subcontrol.ID}).MustDelete(testUser1.UserCtx, t)
}

// TestQueryControlCategoriesByFramework tests the query for control subcategories by framework
// Note: this test will pull all categories, even if the controls weren't created in this test, or in this organization (E.g. public standards)
// will affect the results of this test
// never try to run this in parallel with other tests that create controls
// or standards, or that have controls linked to them
func TestQueryControlCategoriesByFramework(t *testing.T) {
	customFramework := "Custom"

	newUser := suite.userBuilder(context.Background(), t)
	newUser2 := suite.userBuilder(context.Background(), t)

	// create controls with categories and subcategories
	control1 := (&ControlBuilder{client: suite.client, AllFields: true}).MustNew(newUser.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client, AllFields: true}).MustNew(newUser.UserCtx, t)

	// create one without a category
	control3 := (&ControlBuilder{client: suite.client}).MustNew(newUser.UserCtx, t)

	// create one with a duplicate category
	control4 := (&ControlBuilder{client: suite.client, Category: control1.Category}).MustNew(newUser.UserCtx, t)

	// create one with a different framework
	standard := (&StandardBuilder{client: suite.client}).MustNew(newUser.UserCtx, t)
	control5 := (&ControlBuilder{client: suite.client, Category: "Meow" + ulids.New().String(), StandardID: standard.ID}).MustNew(newUser.UserCtx, t)
	// create another with the another category
	control6 := (&ControlBuilder{client: suite.client, Category: "Woof" + ulids.New().String(), StandardID: standard.ID}).MustNew(newUser.UserCtx, t)
	// create one with with duplicate category
	control7 := (&ControlBuilder{client: suite.client, Category: control5.Category, StandardID: standard.ID}).MustNew(newUser.UserCtx, t)

	// add default where to limit overlap between tests
	defaultWhere := &testclient.ControlWhereInput{
		OwnerIDIn: []string{newUser.OrganizationID},
	}

	testCases := []struct {
		name           string
		client         *testclient.TestClient
		where          *testclient.ControlWhereInput
		ctx            context.Context
		expectedErr    string
		expectedResult []*testclient.GetControlCategoriesWithFramework_ControlCategoriesByFramework
	}{
		{
			name:   "happy path, get control categories",
			where:  defaultWhere,
			client: suite.client.api,
			ctx:    newUser.UserCtx,
			expectedResult: []*testclient.GetControlCategoriesWithFramework_ControlCategoriesByFramework{
				{
					Node: testclient.GetControlCategoriesWithFramework_ControlCategoriesByFramework_Node{
						Name:               control1.Category,
						ReferenceFramework: &customFramework,
					},
				},
				{
					Node: testclient.GetControlCategoriesWithFramework_ControlCategoriesByFramework_Node{
						Name:               control2.Category,
						ReferenceFramework: &customFramework,
					},
				},
				{
					Node: testclient.GetControlCategoriesWithFramework_ControlCategoriesByFramework_Node{
						Name:               control5.Category,
						ReferenceFramework: &standard.ShortName,
					},
				},
				{
					Node: testclient.GetControlCategoriesWithFramework_ControlCategoriesByFramework_Node{
						Name:               control6.Category,
						ReferenceFramework: &standard.ShortName,
					},
				},
			},
		},
		{
			name:   "filter by standard, two results expected",
			client: suite.client.api,
			ctx:    newUser.UserCtx,
			where: &testclient.ControlWhereInput{
				StandardID: &standard.ID,
			},
			expectedResult: []*testclient.GetControlCategoriesWithFramework_ControlCategoriesByFramework{
				{
					Node: testclient.GetControlCategoriesWithFramework_ControlCategoriesByFramework_Node{
						Name:               control5.Category,
						ReferenceFramework: &standard.ShortName,
					},
				},
				{
					Node: testclient.GetControlCategoriesWithFramework_ControlCategoriesByFramework_Node{
						Name:               control6.Category,
						ReferenceFramework: &standard.ShortName,
					},
				},
			},
		},
		{
			name: "no controls, no results",
			where: &testclient.ControlWhereInput{
				OwnerIDIn: []string{newUser2.OrganizationID},
			},
			client:         suite.client.api,
			ctx:            newUser2.UserCtx,
			expectedResult: []*testclient.GetControlCategoriesWithFramework_ControlCategoriesByFramework{},
		},
	}

	for _, tc := range testCases {
		t.Run("Get Categories "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetControlCategoriesWithFramework(tc.ctx, tc.where)
			if tc.expectedErr != "" {

				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.ControlCategoriesByFramework, len(tc.expectedResult)))

			// sort the categories so they are consistent
			slices.SortFunc(tc.expectedResult, func(a, b *testclient.GetControlCategoriesWithFramework_ControlCategoriesByFramework) int {
				return cmp.Compare(a.Node.Name, b.Node.Name)
			})

			for _, category := range tc.expectedResult {
				foundCat := false
				for _, respCat := range resp.ControlCategoriesByFramework {
					// check for empty categories
					assert.Check(t, category.Node.Name != "")
					// ensure the reference framework is set
					assert.Check(t, category.Node.ReferenceFramework != nil)

					if category.Node.Name == respCat.Node.Name && category.Node.ReferenceFramework != nil && respCat.Node.ReferenceFramework != nil &&
						*category.Node.ReferenceFramework == *respCat.Node.ReferenceFramework {
						foundCat = true
						break
					}
				}
				assert.Check(t, foundCat, "Expected category %s to be in the response", category.Node.Name)
			}
		})
	}

	// cleanup created controls
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control1.ID, control2.ID, control3.ID, control4.ID, control5.ID, control6.ID, control7.ID}}).MustDelete(newUser.UserCtx, t)
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: standard.ID}).MustDelete(newUser.UserCtx, t)
}

func TestQueryControlSubcategoriesByFramework(t *testing.T) {
	customFramework := "Custom"

	testUser := suite.userBuilder(context.Background(), t)
	testUserAnother := suite.userBuilder(context.Background(), t)

	// create controls with categories and subcategories
	control1 := (&ControlBuilder{client: suite.client, AllFields: true}).MustNew(testUser.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client, AllFields: true}).MustNew(testUser.UserCtx, t)

	// create one without a subcategory
	control3 := (&ControlBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	// create one with a duplicate subcategory
	control4 := (&ControlBuilder{client: suite.client, Subcategory: control1.Subcategory}).MustNew(testUser.UserCtx, t)

	// create one with a different framework
	standard := (&StandardBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	control5 := (&ControlBuilder{client: suite.client, AllFields: true, StandardID: standard.ID}).MustNew(testUser.UserCtx, t)

	// add default where to limit overlap between tests
	defaultWhere := &testclient.ControlWhereInput{
		OwnerIDIn: []string{testUser.OrganizationID},
	}

	testCases := []struct {
		name           string
		client         *testclient.TestClient
		where          *testclient.ControlWhereInput
		ctx            context.Context
		expectedErr    string
		expectedResult []*testclient.GetControlSubcategoriesWithFramework_ControlSubcategoriesByFramework
	}{
		{
			name:   "happy path, get control subcategories",
			where:  defaultWhere,
			client: suite.client.api,
			ctx:    testUser.UserCtx,
			expectedResult: []*testclient.GetControlSubcategoriesWithFramework_ControlSubcategoriesByFramework{
				{
					Node: testclient.GetControlSubcategoriesWithFramework_ControlSubcategoriesByFramework_Node{
						Name:               control1.Subcategory,
						ReferenceFramework: &customFramework,
					},
				},
				{
					Node: testclient.GetControlSubcategoriesWithFramework_ControlSubcategoriesByFramework_Node{
						Name:               control2.Subcategory,
						ReferenceFramework: &customFramework,
					},
				},
				{
					Node: testclient.GetControlSubcategoriesWithFramework_ControlSubcategoriesByFramework_Node{
						Name:               control5.Subcategory,
						ReferenceFramework: &standard.ShortName,
					},
				},
			},
		},
		{
			name:   "filter by standard, one result expected",
			client: suite.client.api,
			ctx:    testUser.UserCtx,
			where: &testclient.ControlWhereInput{
				StandardID: &standard.ID,
			},
			expectedResult: []*testclient.GetControlSubcategoriesWithFramework_ControlSubcategoriesByFramework{
				{
					Node: testclient.GetControlSubcategoriesWithFramework_ControlSubcategoriesByFramework_Node{
						Name:               control5.Subcategory,
						ReferenceFramework: &standard.ShortName,
					},
				},
			},
		},
		{
			name: "no controls, no results",
			where: &testclient.ControlWhereInput{
				OwnerIDIn: []string{testUserAnother.OrganizationID},
			},
			client:         suite.client.api,
			ctx:            testUserAnother.UserCtx,
			expectedResult: []*testclient.GetControlSubcategoriesWithFramework_ControlSubcategoriesByFramework{},
		},
	}

	for _, tc := range testCases {
		t.Run("Get Subcategories "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetControlSubcategoriesWithFramework(tc.ctx, tc.where)
			if tc.expectedErr != "" {

				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// sort the categories so they are consistent
			slices.SortFunc(tc.expectedResult, func(a, b *testclient.GetControlSubcategoriesWithFramework_ControlSubcategoriesByFramework) int {
				return cmp.Compare(a.Node.Name, b.Node.Name)
			})

			for _, category := range tc.expectedResult {
				foundCat := false
				for _, respCat := range resp.ControlSubcategoriesByFramework {
					if category.Node.Name == respCat.Node.Name && category.Node.ReferenceFramework != nil && respCat.Node.ReferenceFramework != nil &&
						*category.Node.ReferenceFramework == *respCat.Node.ReferenceFramework {
						foundCat = true
						break
					}
				}
				assert.Check(t, foundCat, "Expected subcategory %s to be in the response", category.Node.Name)
			}

			for _, category := range resp.ControlSubcategoriesByFramework {
				// check for empty categories
				assert.Check(t, category.Node.Name != "")
			}
		})
	}

	// cleanup created controls
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control1.ID, control2.ID, control3.ID, control4.ID, control5.ID}}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: standard.ID}).MustDelete(testUser.UserCtx, t)
}

func TestQueryControlGroupsByCategory(t *testing.T) {
	user1 := suite.userBuilder(context.Background(), t)

	// create controls with categories and subcategories
	control1 := (&ControlBuilder{client: suite.client, AllFields: true}).MustNew(user1.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client, AllFields: true}).MustNew(user1.UserCtx, t)

	// create one without a category
	control3 := (&ControlBuilder{client: suite.client}).MustNew(user1.UserCtx, t)

	// create one with a duplicate category
	control4 := (&ControlBuilder{client: suite.client, Category: control1.Category}).MustNew(user1.UserCtx, t)

	// create one with a different framework
	standard := (&StandardBuilder{client: suite.client}).MustNew(user1.UserCtx, t)

	// make sure categories are unique across tests or you'll have a bad time
	// with test conflicts
	category := "Meow-" + ulids.New().String()
	control5 := (&ControlBuilder{client: suite.client, Category: category, StandardID: standard.ID}).MustNew(user1.UserCtx, t)
	// create another with the another category
	control6 := (&ControlBuilder{client: suite.client, Category: "Woof-" + ulids.New().String(), StandardID: standard.ID}).MustNew(user1.UserCtx, t)
	// create one with with duplicate category
	control7 := (&ControlBuilder{client: suite.client, Category: category, StandardID: standard.ID}).MustNew(user1.UserCtx, t)

	// create another without a category to test multiple controls in "No Category"
	control8 := (&ControlBuilder{client: suite.client}).MustNew(user1.UserCtx, t)

	// add default where to limit overlap between tests
	defaultWhere := &testclient.ControlWhereInput{
		OwnerIDIn: []string{user1.OrganizationID},
	}

	testCases := []struct {
		name               string
		client             *testclient.TestClient
		first              *int64
		after              *string
		where              *testclient.ControlWhereInput
		category           *string
		ctx                context.Context
		expectedErr        string
		expectedCategories map[string]struct{}
	}{
		{
			name:   "happy path, get control categories",
			where:  defaultWhere,
			client: suite.client.api,
			ctx:    user1.UserCtx,
			expectedCategories: map[string]struct{}{
				control1.Category: {},
				control2.Category: {},
				control5.Category: {},
				control6.Category: {},
				"No Category":     {},
			},
		},
		{
			name:   "happy path, get control categories",
			where:  defaultWhere,
			client: suite.client.api,
			ctx:    user1.UserCtx,
			first:  lo.ToPtr(int64(1)), // test pagination
		},
		{
			name:   "filter by standard, two results expected",
			client: suite.client.api,
			ctx:    user1.UserCtx,
			where: &testclient.ControlWhereInput{
				StandardID: &standard.ID,
			},
			expectedCategories: map[string]struct{}{
				control5.Category: {},
				control6.Category: {},
			},
		},
		{
			name:     "happy path, get controls in No Category",
			where:    defaultWhere,
			client:   suite.client.api,
			ctx:      user1.UserCtx,
			category: lo.ToPtr("No Category"),
			expectedCategories: map[string]struct{}{
				"No Category": {},
			},
		},
		{
			name:   "no controls, no results",
			where:  defaultWhere,
			client: suite.client.api,
			ctx:    testUser2.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Get Controls By Categories "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetControlsGroupByCategory(tc.ctx, tc.first, nil, tc.after, nil, tc.where, nil, tc.category)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.expectedCategories != nil {
				for cat := range tc.expectedCategories {
					foundCat := false
					for _, edge := range resp.ControlsGroupByCategory.Edges {
						if edge.Node.Category == cat {
							foundCat = true

							if cat == "No Category" && tc.category != nil && *tc.category == "No Category" {
								assert.Check(t, edge.Node.Controls.TotalCount == 2, "No Category group should contain exactly 2 controls (control3 and control8)")
							}
							break
						}
					}
					assert.Check(t, foundCat, "Expected category %s to be in the response", cat)
				}
			}
		})
	}

	// cleanup created controls first, then standard
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control1.ID, control2.ID, control3.ID, control4.ID, control5.ID, control6.ID, control7.ID, control8.ID}}).MustDelete(user1.UserCtx, t)
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, ID: standard.ID}).MustDelete(user1.UserCtx, t)
}

func TestMutationUpdateBulkControl(t *testing.T) {
	program1 := (&ProgramBuilder{client: suite.client, EditorIDs: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client, EditorIDs: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)

	control1 := (&ControlBuilder{client: suite.client, ProgramID: program1.ID}).MustNew(testUser1.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client, ProgramID: program1.ID}).MustNew(testUser1.UserCtx, t)
	control3 := (&ControlBuilder{client: suite.client, ProgramID: program1.ID}).MustNew(testUser1.UserCtx, t)

	subcontrol1 := (&SubcontrolBuilder{client: suite.client, ControlID: control1.ID}).MustNew(testUser1.UserCtx, t)
	subcontrol2 := (&SubcontrolBuilder{client: suite.client, ControlID: control2.ID}).MustNew(testUser1.UserCtx, t)

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

	controlAnotherUser := (&ControlBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	// ensure the user does not currently have access to update the control
	res, err := suite.client.api.UpdateBulkControl(testUser2.UserCtx, []string{control1.ID}, testclient.UpdateControlInput{
		Status: lo.ToPtr(enums.ControlStatusPreparing),
	})

	assert.Assert(t, is.Nil(err))
	// make sure nothing was updated
	assert.Equal(t, len(res.UpdateBulkControl.Controls), 0)

	testCases := []struct {
		name                 string
		ids                  []string
		input                testclient.UpdateControlInput
		client               *testclient.TestClient
		ctx                  context.Context
		expectedErr          string
		expectedUpdatedCount int
	}{
		{
			name: "happy path, update status on multiple controls",
			ids:  []string{control1.ID, control2.ID, control3.ID},
			input: testclient.UpdateControlInput{
				Status: &enums.ControlStatusPreparing,
			},
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
			expectedUpdatedCount: 3,
		},
		{
			name: "happy path, clear operations and editor permissions",
			ids:  []string{control1.ID, control2.ID},
			input: testclient.UpdateControlInput{
				ClearReferences:       lo.ToPtr(true),
				ClearMappedCategories: lo.ToPtr(true),
				AddEditorIDs:          []string{groupMember.GroupID},
			},
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
			expectedUpdatedCount: 2,
		},
		{
			name:        "empty ids array",
			ids:         []string{},
			input:       testclient.UpdateControlInput{Description: lo.ToPtr("test")},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "ids is required",
		},
		{
			name: "mixed success and failure - some controls not authorized",
			ids:  []string{control1.ID, controlAnotherUser.ID}, // second control should fail authorization
			input: testclient.UpdateControlInput{
				Status: &enums.ControlStatusPreparing,
			},
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
			expectedUpdatedCount: 1, // only control1 should be updated
		},
		{
			name: "update allowed, user added to one of the programs",
			ids:  []string{control1.ID, control2.ID},
			input: testclient.UpdateControlInput{
				Status: &enums.ControlStatusApproved,
			},
			client:               suite.client.api,
			ctx:                  adminUser.UserCtx,
			expectedUpdatedCount: 2,
		},
		{
			name: "update not allowed, no permissions to controls",
			ids:  []string{control1.ID},
			input: testclient.UpdateControlInput{
				Status: &enums.ControlStatusPreparing,
			},
			client:               suite.client.api,
			ctx:                  testUser2.UserCtx,
			expectedUpdatedCount: 0, // should not find any controls to update
		},
		{
			name: "update control type and category on multiple controls",
			ids:  []string{control1.ID, control2.ID, control3.ID},
			input: testclient.UpdateControlInput{
				Category:       lo.ToPtr("Availability"),
				ControlOwnerID: &ownerGroup.ID,
			},
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
			expectedUpdatedCount: 3,
		},
		{
			name: "add programs and control implementations to multiple controls",
			ids:  []string{control1.ID, control2.ID},
			input: testclient.UpdateControlInput{
				AddProgramIDs:               []string{program2.ID},
				AddControlImplementationIDs: []string{controlImplementation.ID},
				Tags:                        []string{"bulk", "update"},
			},
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
			expectedUpdatedCount: 2,
		},
	}

	for _, tc := range testCases {
		t.Run("Bulk Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateBulkControl(tc.ctx, tc.ids, tc.input)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.UpdateBulkControl.Controls, tc.expectedUpdatedCount))
			assert.Check(t, is.Len(resp.UpdateBulkControl.UpdatedIDs, tc.expectedUpdatedCount))

			// verify all returned controls have the expected values from tc.input
			for _, control := range resp.UpdateBulkControl.Controls {
				if tc.input.Description != nil {
					assert.Check(t, is.Equal(*tc.input.Description, *control.Description))
				}

				if tc.input.Status != nil {
					assert.Check(t, is.Equal(*tc.input.Status, *control.Status))
				}

				if tc.input.Tags != nil {
					assert.Check(t, is.DeepEqual(tc.input.Tags, control.Tags))
				}

				if tc.input.Category != nil {
					assert.Check(t, is.Equal(*tc.input.Category, *control.Category))
				}

				if tc.input.ControlOwnerID != nil {
					assert.Check(t, control.ControlOwner != nil)
					assert.Check(t, is.Equal(*tc.input.ControlOwnerID, control.ControlOwner.ID))
				}

				if tc.input.DelegateID != nil {
					assert.Check(t, control.Delegate != nil)
					assert.Check(t, is.Equal(*tc.input.DelegateID, control.Delegate.ID))
				}

				if tc.input.AppendReferences != nil {
					assert.Check(t, is.DeepEqual(tc.input.AppendReferences, control.References))
				}

				if tc.input.ClearReferences != nil && *tc.input.ClearReferences {
					assert.Check(t, is.Len(control.References, 0))
				}

				if tc.input.AppendMappedCategories != nil {
					assert.Check(t, is.DeepEqual(tc.input.AppendMappedCategories, control.MappedCategories))
				}

				if tc.input.ClearMappedCategories != nil && *tc.input.ClearMappedCategories {
					assert.Check(t, is.Len(control.MappedCategories, 0))
				}

				if tc.input.AppendControlQuestions != nil {
					assert.Check(t, is.DeepEqual(tc.input.AppendControlQuestions, control.ControlQuestions))
				}

				if tc.input.AppendAssessmentObjectives != nil {
					assert.Check(t, is.DeepEqual(tc.input.AppendAssessmentObjectives, control.AssessmentObjectives))
				}

				if tc.input.AddControlImplementationIDs != nil {
					assert.Check(t, is.Len(control.ControlImplementations.Edges, len(tc.input.AddControlImplementationIDs)))
				}

				// ensure the program is set
				if len(tc.input.AddProgramIDs) > 0 {
					foundPrograms := 0
					for _, programID := range tc.input.AddProgramIDs {
						for _, edge := range control.Programs.Edges {
							if edge.Node.ID == programID {
								foundPrograms++
								break
							}
						}
					}
					assert.Check(t, foundPrograms > 0)
				}

				if len(tc.input.AddEditorIDs) > 0 {
					found := false
					for _, edge := range control.Editors.Edges {
						for _, editorID := range tc.input.AddEditorIDs {
							if edge.Node.ID == editorID {
								found = true
								break
							}
						}
					}
					assert.Check(t, found)

					// ensure the user has access to the control now
					res, err := suite.client.api.UpdateControl(anotherViewerUser.UserCtx, control.ID, testclient.UpdateControlInput{
						Tags: []string{"bulk-test-tag"},
					})
					assert.NilError(t, err)
					assert.Check(t, res != nil)
					assert.Check(t, is.Equal(control.ID, res.UpdateControl.Control.ID))
					assert.Check(t, slices.Contains(res.UpdateControl.Control.Tags, "bulk-test-tag"))
				}
			}

			// verify that the returned IDs match the ones that were actually updated
			for _, updatedID := range resp.UpdateBulkControl.UpdatedIDs {
				found := false
				for _, expectedID := range tc.ids {
					if expectedID == updatedID {
						found = true
						break
					}
				}
				assert.Check(t, found, "Updated ID %s should be in the original request", updatedID)
			}
		})
	}

	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, IDs: []string{subcontrol1.ID, subcontrol2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control1.ID, control2.ID, control3.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: controlAnotherUser.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, IDs: []string{program1.ID, program2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlImplementationDeleteOne]{client: suite.client.db.ControlImplementation, ID: controlImplementation.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{ownerGroup.ID, delegateGroup.ID, groupMember.GroupID}}).MustDelete(testUser1.UserCtx, t)
}
