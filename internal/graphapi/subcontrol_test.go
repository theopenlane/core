package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestQuerySubcontrol(t *testing.T) {
	program := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// add adminUser to the program so that they can create a subcontrol
	(&th.ProgramMemberBuilder{Client: suite.Client, ProgramID: program.ID,
		UserID: th.SharedAdminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(th.SharedTestUser1.UserCtx, t)

	createdControlIDs := []string{}
	createdSubcontrolIDs := []string{}
	// add test cases for querying the subcontrol
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
			name:   "read only user, same org, access to the parent control via organization",
			client: suite.Client.API,
			ctx:    th.SharedViewOnlyUser.UserCtx,
		},
		{
			name:   "admin user, access to the parent control via the program",
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name:   "happy path using personal access token",
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name:     "subcontrol not found, invalid ID",
			queryID:  "invalid",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "subcontrol not found, using not authorized user",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			// setup the subcontrol if it is not already created
			if tc.queryID == "" {
				// create the control first
				control, err := suite.Client.API.CreateControl(th.SharedTestUser1.UserCtx,
					testclient.CreateControlInput{
						RefCode:    "SC-" + ulids.New().String(),
						ProgramIDs: []string{program.ID},
					})

				assert.NilError(t, err)
				assert.Assert(t, control != nil)

				createdControlIDs = append(createdControlIDs, control.CreateControl.Control.ID)

				resp, err := suite.Client.API.CreateSubcontrol(th.SharedTestUser1.UserCtx,
					testclient.CreateSubcontrolInput{
						RefCode:   "SC-1" + ulids.New().String(),
						ControlID: control.CreateControl.Control.ID,
					})

				assert.NilError(t, err)
				assert.Assert(t, resp != nil)

				tc.queryID = resp.CreateSubcontrol.Subcontrol.ID
				createdSubcontrolIDs = append(createdSubcontrolIDs, resp.CreateSubcontrol.Subcontrol.ID)
			}

			resp, err := tc.client.GetSubcontrolByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.Subcontrol.ID))
			assert.Check(t, len(resp.Subcontrol.RefCode) != 0)

			assert.Check(t, len(resp.Subcontrol.ControlID) != 0)
		})
	}

	// cleanup the program
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, ID: program.ID}).
		MustDelete(th.SharedTestUser1.UserCtx, t)

	// cleanup the controls
	(&th.Cleanup[*generated.SubcontrolDeleteOne]{Client: suite.Client.DB.Subcontrol, IDs: createdSubcontrolIDs}).
		MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, IDs: createdControlIDs}).
		MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQuerySubcontrols(t *testing.T) {
	// create multiple objects to be queried using th.SharedTestUser1
	sc1 := (&th.SubcontrolBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	sc2 := (&th.SubcontrolBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			name:            "happy path, using read only user of the same org",
			client:          suite.Client.API,
			ctx:             th.SharedViewOnlyUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, api token with org access",
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
			name:            "another user, no subcontrols should be returned",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllSubcontrols(tc.ctx)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Subcontrols.Edges, tc.expectedResults), "expected %d, got %d", tc.expectedResults, len(resp.Subcontrols.Edges))
		})
	}

	// cleanup the controls
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, IDs: []string{sc1.ControlID, sc2.ControlID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	// cleanup the subcontrols
	(&th.Cleanup[*generated.SubcontrolDeleteOne]{Client: suite.Client.DB.Subcontrol, IDs: []string{sc1.ID, sc2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationCreateSubcontrol(t *testing.T) {
	program := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	ownerGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	anotherOwnerGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	delegateGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// add adminUser to the program
	(&th.ProgramMemberBuilder{Client: suite.Client, ProgramID: program.ID,
		UserID: th.SharedAdminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(th.SharedTestUser1.UserCtx, t)

	control1 := (&th.ControlBuilder{Client: suite.Client, ProgramID: program.ID}).MustNew(th.SharedTestUser1.UserCtx, t)
	control2 := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	controlWithOwner := (&th.ControlBuilder{Client: suite.Client, ProgramID: program.ID,
		ControlOwnerID: ownerGroup.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name                 string
		request              testclient.CreateSubcontrolInput
		createParentControl  bool
		client               *testclient.TestClient
		ctx                  context.Context
		expectedRefFramework *string
		expectedErr          string
	}{
		{
			name: "missing required ref code",
			request: testclient.CreateSubcontrolInput{
				Description: lo.ToPtr("A description of the Subcontrol"),
				ControlID:   control1.ID,
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "happy path, minimal input",
			request: testclient.CreateSubcontrolInput{
				RefCode:   "SC-1",
				ControlID: control1.ID,
			},
			client:               suite.Client.API,
			expectedRefFramework: control1.ReferenceFramework,
			ctx:                  th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, parent control has owner",
			request: testclient.CreateSubcontrolInput{
				RefCode:   "SC-2-1",
				ControlID: controlWithOwner.ID,
			},
			client:               suite.Client.API,
			expectedRefFramework: controlWithOwner.ReferenceFramework,
			ctx:                  th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, parent control has owner, subcontrol should override it",
			request: testclient.CreateSubcontrolInput{
				RefCode:        "SC-2-2",
				ControlID:      controlWithOwner.ID,
				ControlOwnerID: &anotherOwnerGroup.ID,
			},
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser1.UserCtx,
			expectedRefFramework: controlWithOwner.ReferenceFramework,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateSubcontrolInput{
				RefCode:     "Another Subcontrol",
				Description: lo.ToPtr("A description of the Subcontrol"),
				Status:      &enums.ControlStatusPreparing,
				Tags:        []string{"tag1", "tag2"},
				Source:      &enums.ControlSourceFramework,
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
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser1.UserCtx,
			expectedRefFramework: control2.ReferenceFramework,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateSubcontrolInput{
				RefCode:   "Subcontrol",
				ControlID: control1.ID,
				OwnerID:   &th.SharedTestUser1.OrganizationID,
			},
			client:               suite.Client.APIWithPAT,
			ctx:                  context.Background(),
			expectedRefFramework: control1.ReferenceFramework,
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateSubcontrolInput{
				RefCode:   "SC-1",
				ControlID: control1.ID,
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "user authorized, they have access to the parent control via the program",
			request: testclient.CreateSubcontrolInput{
				RefCode:   "SC-1",
				ControlID: control1.ID,
			},
			createParentControl:  true, // create the parent control first
			client:               suite.Client.API,
			ctx:                  th.SharedAdminUser.UserCtx,
			expectedRefFramework: control1.ReferenceFramework,
		},
		{
			name: "user not authorized to edit one of the controls",
			request: testclient.CreateSubcontrolInput{
				RefCode:   "SC-1",
				ControlID: control2.ID,
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "missing required control ID",
			request: testclient.CreateSubcontrolInput{
				RefCode: "SC-1",
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "validator failed for field",
		},
		{
			name: "invalid control ID",
			request: testclient.CreateSubcontrolInput{
				RefCode:   "SC-1",
				ControlID: "invalid-control-id",
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.createParentControl {
				// create the control first
				control, err := suite.Client.API.CreateControl(th.SharedTestUser1.UserCtx,
					testclient.CreateControlInput{
						RefCode:    "SC",
						ProgramIDs: []string{program.ID},
					})

				assert.NilError(t, err)
				assert.Assert(t, control != nil)

				tc.request.ControlID = control.CreateControl.Control.ID
			}

			resp, err := tc.client.CreateSubcontrol(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// check required fields
			assert.Check(t, resp.CreateSubcontrol.Subcontrol.ID != "")
			assert.Check(t, is.Equal(tc.request.RefCode, resp.CreateSubcontrol.Subcontrol.RefCode))
			assert.Check(t, is.Contains(resp.CreateSubcontrol.Subcontrol.DisplayID, "SCL-"))
			assert.Check(t, is.Equal(tc.request.RefCode, resp.CreateSubcontrol.Subcontrol.RefCode))

			assert.Equal(t, tc.expectedRefFramework, resp.CreateSubcontrol.Subcontrol.ReferenceFramework)

			// ensure the control is set
			assert.Check(t, is.Equal(tc.request.ControlID, resp.CreateSubcontrol.Subcontrol.ControlID))

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.CreateSubcontrol.Subcontrol.Description))
			} else {
				assert.Check(t, is.Equal(*resp.CreateSubcontrol.Subcontrol.Description, ""))
			}

			if tc.request.Status != nil {
				assert.Check(t, is.Equal(*tc.request.Status, *resp.CreateSubcontrol.Subcontrol.Status))
			} else {
				assert.Check(t, is.Equal(enums.ControlStatusNotImplemented, *resp.CreateSubcontrol.Subcontrol.Status))
			}

			if tc.request.Source != nil {
				assert.Check(t, is.Equal(*tc.request.Source, *resp.CreateSubcontrol.Subcontrol.Source))
			} else {
				assert.Check(t, is.Equal(enums.ControlSourceUserDefined, *resp.CreateSubcontrol.Subcontrol.Source))
			}

			if tc.request.ControlOwnerID != nil {
				assert.Assert(t, resp.CreateSubcontrol.Subcontrol.ControlOwner != nil)
				assert.Check(t, is.Equal(*tc.request.ControlOwnerID, resp.CreateSubcontrol.Subcontrol.ControlOwner.ID))
			} else if tc.request.ControlID == controlWithOwner.ID {
				// it should inherit the owner from the parent control if it was set
				assert.Assert(t, resp.CreateSubcontrol.Subcontrol.ControlOwner != nil)
				assert.Check(t, is.Equal(*controlWithOwner.ControlOwnerID, resp.CreateSubcontrol.Subcontrol.ControlOwner.ID))
			} else {
				assert.Check(t, is.Nil(resp.CreateSubcontrol.Subcontrol.ControlOwner))
			}

			if tc.request.DelegateID != nil {
				assert.Assert(t, resp.CreateSubcontrol.Subcontrol.Delegate != nil)
				assert.Check(t, is.Equal(*tc.request.DelegateID, resp.CreateSubcontrol.Subcontrol.Delegate.ID))
			} else {
				assert.Check(t, is.Nil(resp.CreateSubcontrol.Subcontrol.Delegate))
			}

			// ensure the org owner has access to the subcontrol that was created by an api token
			if tc.client == suite.Client.APIWithToken {
				res, err := suite.Client.API.GetSubcontrolByID(th.SharedTestUser1.UserCtx, resp.CreateSubcontrol.Subcontrol.ID)
				assert.NilError(t, err)
				assert.Assert(t, res != nil)
				assert.Check(t, is.Equal(resp.CreateSubcontrol.Subcontrol.ID, res.Subcontrol.ID))
			}
		})
	}

	// cleanup the program
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, ID: program.ID}).
		MustDelete(th.SharedTestUser1.UserCtx, t)
	// cleanup the controls
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, IDs: []string{control1.ID, control2.ID, controlWithOwner.ID}}).
		MustDelete(th.SharedTestUser1.UserCtx, t)
	// cleanup the groups
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, IDs: []string{ownerGroup.ID, anotherOwnerGroup.ID, delegateGroup.ID}}).
		MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationUpdateSubcontrol(t *testing.T) {
	control1 := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	subcontrol := (&th.SubcontrolBuilder{Client: suite.Client, ControlID: control1.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	ownerGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	delegateGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	kind := (&th.CustomTypeEnumBuilder{Client: suite.Client, ObjectType: "control"}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.UpdateSubcontrolInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: testclient.UpdateSubcontrolInput{
				Description: lo.ToPtr("Updated description"),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: testclient.UpdateSubcontrolInput{
				Status:             &enums.ControlStatusPreparing,
				Tags:               []string{"tag1", "tag2"},
				Source:             lo.ToPtr(enums.ControlSourceUserDefined),
				ControlOwnerID:     &ownerGroup.ID,
				DelegateID:         &delegateGroup.ID,
				SubcontrolKindName: &kind.Name,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update ref code to empty",
			request: testclient.UpdateSubcontrolInput{
				RefCode: lo.ToPtr(""),
			},
			client:      suite.Client.APIWithPAT,
			ctx:         context.Background(),
			expectedErr: "value is less than the required length",
		},
		{
			name: "update not allowed, no permissions in same org",
			request: testclient.UpdateSubcontrolInput{
				MappedCategories: []string{"Category1", "Category2"},
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateSubcontrolInput{
				MappedCategories: []string{"Category1", "Category2"},
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateSubcontrol(tc.ctx, subcontrol.ID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.UpdateSubcontrol.Subcontrol.Description))
			}

			if tc.request.Status != nil {
				assert.Check(t, is.Equal(*tc.request.Status, *resp.UpdateSubcontrol.Subcontrol.Status))
			}

			if tc.request.Tags != nil {
				assert.Check(t, is.DeepEqual(tc.request.Tags, resp.UpdateSubcontrol.Subcontrol.Tags))
			}

			if tc.request.Source != nil {
				assert.Check(t, is.Equal(*tc.request.Source, *resp.UpdateSubcontrol.Subcontrol.Source))
			}

			if tc.request.ControlOwnerID != nil {
				assert.Assert(t, resp.UpdateSubcontrol.Subcontrol.ControlOwner != nil)
				assert.Check(t, is.Equal(*tc.request.ControlOwnerID, resp.UpdateSubcontrol.Subcontrol.ControlOwner.ID))
			} else {
				assert.Check(t, is.Nil(resp.UpdateSubcontrol.Subcontrol.ControlOwner))
			}

			if tc.request.DelegateID != nil {
				assert.Assert(t, resp.UpdateSubcontrol.Subcontrol.Delegate != nil)
				assert.Check(t, is.Equal(*tc.request.DelegateID, resp.UpdateSubcontrol.Subcontrol.Delegate.ID))
			} else {
				assert.Check(t, is.Nil(resp.UpdateSubcontrol.Subcontrol.Delegate))
			}

			if tc.request.SubcontrolKindName != nil {
				assert.Assert(t, resp.UpdateSubcontrol.Subcontrol.SubcontrolKindName != nil)
				assert.Check(t, is.Equal(*tc.request.SubcontrolKindName, *resp.UpdateSubcontrol.Subcontrol.SubcontrolKindName))
			}
		})
	}

	// cleanup the subcontrol
	(&th.Cleanup[*generated.SubcontrolDeleteOne]{Client: suite.Client.DB.Subcontrol, ID: subcontrol.ID}).
		MustDelete(th.SharedTestUser1.UserCtx, t)
	// cleanup the control
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, IDs: []string{control1.ID}}).
		MustDelete(th.SharedTestUser1.UserCtx, t)
	// cleanup the groups
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, IDs: []string{ownerGroup.ID, delegateGroup.ID}}).
		MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteSubcontrol(t *testing.T) {
	// create objects to be deleted
	subcontrol1 := (&th.SubcontrolBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	subcontrol2 := (&th.SubcontrolBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  subcontrol1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: subcontrol1.ID,
			client:     suite.Client.API,
			ctx:        th.SharedTestUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  subcontrol1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: subcontrol2.ID,
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
			resp, err := tc.client.DeleteSubcontrol(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {

				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteSubcontrol.DeletedID))
		})
	}

	// cleanup the controls
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, IDs: []string{subcontrol1.ControlID, subcontrol2.ControlID}}).
		MustDelete(th.SharedTestUser1.UserCtx, t)
}
