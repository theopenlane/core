package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/enums"
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
						Name:       "SC-1",
						ProgramIDs: []string{program.ID},
					})

				require.NoError(t, err)
				require.NotNil(t, control)

				resp, err := suite.client.api.CreateSubcontrol(testUser1.UserCtx,
					openlaneclient.CreateSubcontrolInput{
						Name:       "SC-1",
						ControlIDs: []string{control.CreateControl.Control.ID},
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
			assert.NotEmpty(t, resp.Subcontrol.Name)

			require.Len(t, resp.Subcontrol.Controls, 1)
			assert.NotEmpty(t, resp.Subcontrol.Controls[0].ID)
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

	// add adminUser to the program so that they can create a subcontrol
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)

	control1 := (&ControlBuilder{client: suite.client, ProgramID: program.ID}).MustNew(testUser1.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name                string
		request             openlaneclient.CreateSubcontrolInput
		createParentControl bool
		client              *openlaneclient.OpenlaneClient
		ctx                 context.Context
		expectedErr         string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateSubcontrolInput{
				Name:       "SC-1",
				ControlIDs: []string{control1.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateSubcontrolInput{
				Name:                           "Another Subcontrol",
				Description:                    lo.ToPtr("A description of the Subcontrol"),
				Status:                         lo.ToPtr("mitigated"),
				SubcontrolType:                 lo.ToPtr("operational"),
				Version:                        lo.ToPtr("1.0.0"),
				SubcontrolNumber:               lo.ToPtr("1.1"),
				Family:                         lo.ToPtr("AC"),
				Class:                          lo.ToPtr("AC-1"),
				Source:                         lo.ToPtr("NIST framework"),
				MappedFrameworks:               lo.ToPtr("NIST"),
				ImplementationEvidence:         lo.ToPtr("Evidence of implementation"),
				ImplementationStatus:           lo.ToPtr("Implemented"),
				ImplementationDate:             lo.ToPtr(time.Now().AddDate(0, 0, -1)),
				ImplementationVerification:     lo.ToPtr("Verification of implementation"),
				ImplementationVerificationDate: lo.ToPtr(time.Now().AddDate(0, 0, -1)),
				Details:                        map[string]interface{}{"stuff": "things"},
				ControlIDs:                     []string{control1.ID, control2.ID}, // multiple controls
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: openlaneclient.CreateSubcontrolInput{
				Name:       "Subcontrol",
				ControlIDs: []string{control1.ID},
				OwnerID:    testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "using pat, missing owner ID",
			request: openlaneclient.CreateSubcontrolInput{
				Name:       "SC-1",
				ControlIDs: []string{control1.ID},
			},
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			expectedErr: "organization_id is required",
		},
		{
			name: "user not authorized, not enough permissions",
			request: openlaneclient.CreateSubcontrolInput{
				Name:       "SC-1",
				ControlIDs: []string{control1.ID},
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user authorized, they have access to the parent control via the program",
			request: openlaneclient.CreateSubcontrolInput{
				Name:       "SC-1",
				ControlIDs: []string{control1.ID},
			},
			createParentControl: true, // create the parent control first
			client:              suite.client.api,
			ctx:                 adminUser.UserCtx,
		},
		{
			name: "user not authorized to one of the controls",
			request: openlaneclient.CreateSubcontrolInput{
				Name:       "SC-1",
				ControlIDs: []string{control1.ID, control2.ID},
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "missing required control ID",
			request: openlaneclient.CreateSubcontrolInput{
				Name: "SC-1",
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "missing required edge",
		},
		{
			name: "missing required name",
			request: openlaneclient.CreateSubcontrolInput{
				ControlIDs: []string{control1.ID},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.createParentControl {
				// create the control first
				control, err := suite.client.api.CreateControl(testUser1.UserCtx,
					openlaneclient.CreateControlInput{
						Name:       "SC",
						ProgramIDs: []string{program.ID},
					})

				require.NoError(t, err)
				require.NotNil(t, control)

				tc.request.ControlIDs = []string{control.CreateControl.Control.ID}
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
			assert.Equal(t, tc.request.Name, resp.CreateSubcontrol.Subcontrol.Name)

			// ensure the control is set
			if len(tc.request.ControlIDs) > 0 {
				require.NotEmpty(t, resp.CreateSubcontrol.Subcontrol.Controls)
				require.Len(t, resp.CreateSubcontrol.Subcontrol.Controls, len(tc.request.ControlIDs))

				for i, c := range resp.CreateSubcontrol.Subcontrol.Controls {
					assert.Equal(t, tc.request.ControlIDs[i], c.ID)
				}
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.Controls)
			}

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.CreateSubcontrol.Subcontrol.Description)
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.Description)
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, *resp.CreateSubcontrol.Subcontrol.Status)
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.Status)
			}

			if tc.request.SubcontrolType != nil {
				assert.Equal(t, *tc.request.SubcontrolType, *resp.CreateSubcontrol.Subcontrol.SubcontrolType)
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.SubcontrolType)
			}

			if tc.request.Version != nil {
				assert.Equal(t, *tc.request.Version, *resp.CreateSubcontrol.Subcontrol.Version)
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.Version)
			}

			if tc.request.SubcontrolNumber != nil {
				assert.Equal(t, *tc.request.SubcontrolNumber, *resp.CreateSubcontrol.Subcontrol.SubcontrolNumber)
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.SubcontrolNumber)
			}

			if tc.request.Family != nil {
				assert.Equal(t, *tc.request.Family, *resp.CreateSubcontrol.Subcontrol.Family)
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.Family)
			}

			if tc.request.Class != nil {
				assert.Equal(t, *tc.request.Class, *resp.CreateSubcontrol.Subcontrol.Class)
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.Class)
			}

			if tc.request.Source != nil {
				assert.Equal(t, *tc.request.Source, *resp.CreateSubcontrol.Subcontrol.Source)
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.Source)
			}

			if tc.request.MappedFrameworks != nil {
				assert.Equal(t, *tc.request.MappedFrameworks, *resp.CreateSubcontrol.Subcontrol.MappedFrameworks)
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.MappedFrameworks)
			}

			if tc.request.ImplementationEvidence != nil {
				assert.Equal(t, *tc.request.ImplementationEvidence, *resp.CreateSubcontrol.Subcontrol.ImplementationEvidence)
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.ImplementationEvidence)
			}

			if tc.request.ImplementationStatus != nil {
				assert.Equal(t, *tc.request.ImplementationStatus, *resp.CreateSubcontrol.Subcontrol.ImplementationStatus)
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.ImplementationStatus)
			}

			if tc.request.ImplementationDate != nil {
				assert.WithinDuration(t, *tc.request.ImplementationDate, *resp.CreateSubcontrol.Subcontrol.ImplementationDate, time.Second)
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.ImplementationDate)
			}

			if tc.request.ImplementationVerification != nil {
				assert.Equal(t, *tc.request.ImplementationVerification, *resp.CreateSubcontrol.Subcontrol.ImplementationVerification)
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.ImplementationVerification)
			}

			if tc.request.ImplementationVerificationDate != nil {
				assert.WithinDuration(t, *tc.request.ImplementationVerificationDate, *resp.CreateSubcontrol.Subcontrol.ImplementationVerificationDate, time.Second)
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.ImplementationVerificationDate)
			}

			if tc.request.Details != nil {
				assert.Equal(t, tc.request.Details, resp.CreateSubcontrol.Subcontrol.Details)
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.Details)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateSubcontrol() {
	t := suite.T()

	control1 := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
				Description:   lo.ToPtr("Updated description"),
				AddControlIDs: []string{control1.ID, control2.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: openlaneclient.UpdateSubcontrolInput{
				Status:                         lo.ToPtr("mitigated"),
				Tags:                           []string{"tag1", "tag2"},
				Version:                        lo.ToPtr("1.0.1"),
				SubcontrolNumber:               lo.ToPtr("1.2"),
				Family:                         lo.ToPtr("AB"),
				Class:                          lo.ToPtr("AB-2"),
				Source:                         lo.ToPtr("ISO27001"),
				MappedFrameworks:               lo.ToPtr("ISO"),
				ImplementationEvidence:         lo.ToPtr("Evidence of implementation"),
				ImplementationStatus:           lo.ToPtr("Implemented"),
				ImplementationDate:             lo.ToPtr(time.Now().AddDate(0, 0, -10)),
				ImplementationVerification:     lo.ToPtr("Verification of implementation"),
				ImplementationVerificationDate: lo.ToPtr(time.Now().AddDate(0, 0, -7)),
				RemoveControlIDs:               []string{control2.ID},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, cannot remove the only control",
			request: openlaneclient.UpdateSubcontrolInput{
				RemoveControlIDs: []string{control1.ID},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "subcontrol must have at least one control assigned",
		},
		{
			name: "update not allowed, not permissions in same org",
			request: openlaneclient.UpdateSubcontrolInput{
				MappedFrameworks: lo.ToPtr("SOC"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: openlaneclient.UpdateSubcontrolInput{
				MappedFrameworks: lo.ToPtr("SOC"),
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

			if tc.request.Version != nil {
				assert.Equal(t, *tc.request.Version, *resp.UpdateSubcontrol.Subcontrol.Version)
			}

			if tc.request.SubcontrolNumber != nil {
				assert.Equal(t, *tc.request.SubcontrolNumber, *resp.UpdateSubcontrol.Subcontrol.SubcontrolNumber)
			}

			if tc.request.Family != nil {
				assert.Equal(t, *tc.request.Family, *resp.UpdateSubcontrol.Subcontrol.Family)
			}

			if tc.request.Class != nil {
				assert.Equal(t, *tc.request.Class, *resp.UpdateSubcontrol.Subcontrol.Class)
			}

			if tc.request.Source != nil {
				assert.Equal(t, *tc.request.Source, *resp.UpdateSubcontrol.Subcontrol.Source)
			}

			if tc.request.MappedFrameworks != nil {
				assert.Equal(t, *tc.request.MappedFrameworks, *resp.UpdateSubcontrol.Subcontrol.MappedFrameworks)
			}

			if tc.request.ImplementationEvidence != nil {
				assert.Equal(t, *tc.request.ImplementationEvidence, *resp.UpdateSubcontrol.Subcontrol.ImplementationEvidence)
			}

			if tc.request.ImplementationStatus != nil {
				assert.Equal(t, *tc.request.ImplementationStatus, *resp.UpdateSubcontrol.Subcontrol.ImplementationStatus)
			}

			if tc.request.ImplementationDate != nil {
				assert.WithinDuration(t, *tc.request.ImplementationDate, *resp.UpdateSubcontrol.Subcontrol.ImplementationDate, time.Second)
			}

			if tc.request.ImplementationVerification != nil {
				assert.Equal(t, *tc.request.ImplementationVerification, *resp.UpdateSubcontrol.Subcontrol.ImplementationVerification)
			}

			if tc.request.ImplementationVerificationDate != nil {
				assert.WithinDuration(t, *tc.request.ImplementationVerificationDate, *resp.UpdateSubcontrol.Subcontrol.ImplementationVerificationDate, time.Second)
			}

			if tc.request.Details != nil {
				assert.Equal(t, tc.request.Details, resp.UpdateSubcontrol.Subcontrol.Details)
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
