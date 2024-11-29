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
				resp, err := suite.client.api.CreateSubcontrol(testUser1.UserCtx,
					openlaneclient.CreateSubcontrolInput{
						Name:       "Subcontrol",
						ProgramIDs: []string{program.ID},
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

			require.Len(t, resp.Subcontrol.Programs, 1)
			assert.NotEmpty(t, resp.Subcontrol.Programs[0].ID)
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

	program1 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	programAnotherUser := (&ProgramBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	// add adminUser to the program so that they can create a subcontrol associated with the program1
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program1.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)

	// create groups to be associated with the subcontrol
	blockedGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	viewerGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.CreateSubcontrolInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateSubcontrolInput{
				Name: "Subcontrol",
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
				ProgramIDs:                     []string{program1.ID, program2.ID}, // multiple programs
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "add groups",
			request: openlaneclient.CreateSubcontrolInput{
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
			request: openlaneclient.CreateSubcontrolInput{
				Name:    "Subcontrol",
				OwnerID: testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "using api token",
			request: openlaneclient.CreateSubcontrolInput{
				Name: "Subcontrol",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: openlaneclient.CreateSubcontrolInput{
				Name: "Subcontrol",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user authorized, they were added to the program",
			request: openlaneclient.CreateSubcontrolInput{
				Name:       "Subcontrol",
				ProgramIDs: []string{program1.ID},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "user authorized, user not authorized to one of the programs",
			request: openlaneclient.CreateSubcontrolInput{
				Name:       "Subcontrol",
				ProgramIDs: []string{program1.ID, program2.ID},
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "missing required name",
			request:     openlaneclient.CreateSubcontrolInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "user not authorized, no permissions to one of the programs",
			request: openlaneclient.CreateSubcontrolInput{
				Name:       "Subcontrol",
				ProgramIDs: []string{programAnotherUser.ID, program1.ID},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
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

			// ensure the program is set
			if len(tc.request.ProgramIDs) > 0 {
				require.NotEmpty(t, resp.CreateSubcontrol.Subcontrol.Programs)
				require.Len(t, resp.CreateSubcontrol.Subcontrol.Programs, len(tc.request.ProgramIDs))

				for i, p := range resp.CreateSubcontrol.Subcontrol.Programs {
					assert.Equal(t, tc.request.ProgramIDs[i], p.ID)
				}
			} else {
				assert.Empty(t, resp.CreateSubcontrol.Subcontrol.Programs)
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

			if len(tc.request.EditorIDs) > 0 {
				require.Len(t, resp.CreateSubcontrol.Subcontrol.Editors, 1)
				for _, edge := range resp.CreateSubcontrol.Subcontrol.Editors {
					assert.Equal(t, testUser1.GroupID, edge.ID)
				}
			}

			if len(tc.request.BlockedGroupIDs) > 0 {
				require.Len(t, resp.CreateSubcontrol.Subcontrol.BlockedGroups, 1)
				for _, edge := range resp.CreateSubcontrol.Subcontrol.BlockedGroups {
					assert.Equal(t, blockedGroup.ID, edge.ID)
				}
			}

			if len(tc.request.ViewerIDs) > 0 {
				require.Len(t, resp.CreateSubcontrol.Subcontrol.Viewers, 1)
				for _, edge := range resp.CreateSubcontrol.Subcontrol.Viewers {
					assert.Equal(t, viewerGroup.ID, edge.ID)
				}
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateSubcontrol() {
	t := suite.T()

	program1 := (&ProgramBuilder{client: suite.client, EditorIDs: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client, EditorIDs: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	subcontrol := (&SubcontrolBuilder{client: suite.client, ProgramID: program1.ID}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can update the subcontrol
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program1.ID, UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).MustNew(testUser1.UserCtx, t)

	// create another admin user and add them to the same organization and group as testUser1
	// this will allow us to test the group editor/viewer permissions
	anotherAdminUser := suite.userBuilder(context.Background())
	suite.addUserToOrganization(&anotherAdminUser, enums.RoleAdmin, testUser1.OrganizationID)

	(&GroupMemberBuilder{client: suite.client, UserID: anotherAdminUser.ID, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)

	// ensure the user does not currently have access to the subcontrol
	res, err := suite.client.api.GetSubcontrolByID(anotherAdminUser.UserCtx, subcontrol.ID)
	require.Error(t, err)
	require.Nil(t, res)

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
				AddProgramIDs: []string{program1.ID, program2.ID}, // add multiple programs
				AddViewerIDs:  []string{testUser1.GroupID},
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
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
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
			name: "update allowed, user added to one of the programs",
			request: openlaneclient.UpdateSubcontrolInput{
				MappedFrameworks: lo.ToPtr("SOC2"),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
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

			if len(tc.request.AddViewerIDs) > 0 {
				require.Len(t, resp.UpdateSubcontrol.Subcontrol.Viewers, 1)
				for _, edge := range resp.UpdateSubcontrol.Subcontrol.Viewers {
					assert.Equal(t, testUser1.GroupID, edge.ID)
				}

				// ensure the user has access to the subcontrol now
				res, err := suite.client.api.GetSubcontrolByID(anotherAdminUser.UserCtx, subcontrol.ID)
				require.NoError(t, err)
				require.NotEmpty(t, res)
				assert.Equal(t, subcontrol.ID, res.Subcontrol.ID)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteSubcontrol() {
	t := suite.T()

	// create objects to be deleted
	Subcontrol1 := (&SubcontrolBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	Subcontrol2 := (&SubcontrolBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  Subcontrol1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: Subcontrol1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  Subcontrol1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: Subcontrol2.ID,
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
