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

func (suite *GraphTestSuite) TestQueryRisk() {
	t := suite.T()

	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can create a Risk
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)

	// add test cases for querying the Risk
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
			name:     "risk not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "risk not found, using not authorized user",
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			// setup the risk if it is not already created
			if tc.queryID == "" {
				resp, err := suite.client.api.CreateRisk(testUser1.UserCtx,
					openlaneclient.CreateRiskInput{
						Name:       "Risk",
						ProgramIDs: []string{program.ID},
					})

				require.NoError(t, err)
				require.NotNil(t, resp)

				tc.queryID = resp.CreateRisk.Risk.ID
			}

			resp, err := tc.client.GetRiskByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			require.NotEmpty(t, resp.Risk)

			assert.Equal(t, tc.queryID, resp.Risk.ID)
			assert.NotEmpty(t, resp.Risk.Name)

			require.Len(t, resp.Risk.Programs, 1)
			assert.NotEmpty(t, resp.Risk.Programs[0].ID)
		})
	}
}

func (suite *GraphTestSuite) TestQueryRisks() {
	t := suite.T()

	// create multiple objects to be queried using testUser1
	(&RiskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&RiskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			name:            "another user, no risks should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllRisks(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.Risks.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateRisk() {
	t := suite.T()

	program1 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	programAnotherUser := (&ProgramBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	// add adminUser to the program so that they can create a risk associated with the program1
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program1.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)

	// create groups to be associated with the risk
	blockedGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	viewerGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.CreateRiskInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateRiskInput{
				Name: "Risk",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateRiskInput{
				Name:          "Another Risk",
				Description:   lo.ToPtr("A description of the Risk"),
				Status:        lo.ToPtr("mitigated"),
				RiskType:      lo.ToPtr("operational"),
				BusinessCosts: lo.ToPtr("much money"),
				Impact:        &enums.RiskImpactHigh,
				Likelihood:    &enums.RiskLikelihoodHigh,
				Mitigation:    lo.ToPtr("did the thing"),
				Satisfies:     lo.ToPtr("controls"),
				Details:       map[string]interface{}{"stuff": "things"},
				ProgramIDs:    []string{program1.ID, program2.ID}, // multiple programs
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "add groups",
			request: openlaneclient.CreateRiskInput{
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
			request: openlaneclient.CreateRiskInput{
				Name:    "Risk",
				OwnerID: testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "using api token",
			request: openlaneclient.CreateRiskInput{
				Name: "Risk",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: openlaneclient.CreateRiskInput{
				Name: "Risk",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user authorized, they were added to the program",
			request: openlaneclient.CreateRiskInput{
				Name:       "Risk",
				ProgramIDs: []string{program1.ID},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "user authorized, user not authorized to one of the programs",
			request: openlaneclient.CreateRiskInput{
				Name:       "Risk",
				ProgramIDs: []string{program1.ID, program2.ID},
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "missing required name",
			request:     openlaneclient.CreateRiskInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "user not authorized, no permissions to one of the programs",
			request: openlaneclient.CreateRiskInput{
				Name:       "Risk",
				ProgramIDs: []string{programAnotherUser.ID, program1.ID},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateRisk(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// check required fields
			require.NotEmpty(t, resp.CreateRisk.Risk.ID)
			assert.Equal(t, tc.request.Name, resp.CreateRisk.Risk.Name)

			// ensure the program is set
			if len(tc.request.ProgramIDs) > 0 {
				require.NotEmpty(t, resp.CreateRisk.Risk.Programs)
				require.Len(t, resp.CreateRisk.Risk.Programs, len(tc.request.ProgramIDs))

				for i, p := range resp.CreateRisk.Risk.Programs {
					assert.Equal(t, tc.request.ProgramIDs[i], p.ID)
				}
			} else {
				assert.Empty(t, resp.CreateRisk.Risk.Programs)
			}

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.CreateRisk.Risk.Description)
			} else {
				assert.Empty(t, resp.CreateRisk.Risk.Description)
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, *resp.CreateRisk.Risk.Status)
			} else {
				assert.Empty(t, resp.CreateRisk.Risk.Status)
			}

			if tc.request.RiskType != nil {
				assert.Equal(t, *tc.request.RiskType, *resp.CreateRisk.Risk.RiskType)
			} else {
				assert.Empty(t, resp.CreateRisk.Risk.RiskType)
			}

			if tc.request.BusinessCosts != nil {
				assert.Equal(t, *tc.request.BusinessCosts, *resp.CreateRisk.Risk.BusinessCosts)
			} else {
				assert.Empty(t, resp.CreateRisk.Risk.BusinessCosts)
			}

			if tc.request.Impact != nil {
				assert.Equal(t, *tc.request.Impact, *resp.CreateRisk.Risk.Impact)
			} else {
				assert.Equal(t, enums.RiskImpactModerate, *resp.CreateRisk.Risk.Impact)
			}

			if tc.request.Likelihood != nil {
				assert.Equal(t, *tc.request.Likelihood, *resp.CreateRisk.Risk.Likelihood)
			} else {
				assert.Equal(t, enums.RiskLikelihoodMid, *resp.CreateRisk.Risk.Likelihood)
			}

			if tc.request.Mitigation != nil {
				assert.Equal(t, *tc.request.Mitigation, *resp.CreateRisk.Risk.Mitigation)
			} else {
				assert.Empty(t, resp.CreateRisk.Risk.Mitigation)
			}

			if tc.request.Satisfies != nil {
				assert.Equal(t, *tc.request.Satisfies, *resp.CreateRisk.Risk.Satisfies)
			} else {
				assert.Empty(t, resp.CreateRisk.Risk.Satisfies)
			}

			if tc.request.Details != nil {
				assert.Equal(t, tc.request.Details, resp.CreateRisk.Risk.Details)
			} else {
				assert.Empty(t, resp.CreateRisk.Risk.Details)
			}

			if len(tc.request.EditorIDs) > 0 {
				require.Len(t, resp.CreateRisk.Risk.Editors, 1)
				for _, edge := range resp.CreateRisk.Risk.Editors {
					assert.Equal(t, testUser1.GroupID, edge.ID)
				}
			}

			if len(tc.request.BlockedGroupIDs) > 0 {
				require.Len(t, resp.CreateRisk.Risk.BlockedGroups, 1)
				for _, edge := range resp.CreateRisk.Risk.BlockedGroups {
					assert.Equal(t, blockedGroup.ID, edge.ID)
				}
			}

			if len(tc.request.ViewerIDs) > 0 {
				require.Len(t, resp.CreateRisk.Risk.Viewers, 1)
				for _, edge := range resp.CreateRisk.Risk.Viewers {
					assert.Equal(t, viewerGroup.ID, edge.ID)
				}
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateRisk() {
	t := suite.T()

	program := (&ProgramBuilder{client: suite.client, EditorIDs: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	risk := (&RiskBuilder{client: suite.client, ProgramID: program.ID}).MustNew(testUser1.UserCtx, t)

	// create another admin user and add them to the same organization and group as testUser1
	// this will allow us to test the group editor/viewer permissions
	anotherAdminUser := suite.userBuilder(context.Background())
	suite.addUserToOrganization(&anotherAdminUser, enums.RoleAdmin, testUser1.OrganizationID)

	(&GroupMemberBuilder{client: suite.client, UserID: anotherAdminUser.ID, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)

	// ensure the user does not currently have access to the risk
	res, err := suite.client.api.GetRiskByID(anotherAdminUser.UserCtx, risk.ID)
	require.Error(t, err)
	require.Nil(t, res)

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateRiskInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: openlaneclient.UpdateRiskInput{
				Description:  lo.ToPtr("Updated description"),
				AddViewerIDs: []string{testUser1.GroupID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: openlaneclient.UpdateRiskInput{
				Satisfies:  lo.ToPtr("Updated controls"),
				Status:     lo.ToPtr("mitigated"),
				Tags:       []string{"tag1", "tag2"},
				Mitigation: lo.ToPtr("Updated mitigation"),
				Impact:     &enums.RiskImpactModerate,
				Likelihood: &enums.RiskLikelihoodLow,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, not permissions in same org",
			request: openlaneclient.UpdateRiskInput{
				Likelihood: &enums.RiskLikelihoodLow,
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: openlaneclient.UpdateRiskInput{
				Likelihood: &enums.RiskLikelihoodLow,
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateRisk(tc.ctx, risk.ID, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.UpdateRisk.Risk.Description)
			}

			if tc.request.Satisfies != nil {
				assert.Equal(t, *tc.request.Satisfies, *resp.UpdateRisk.Risk.Satisfies)
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, *resp.UpdateRisk.Risk.Status)
			}

			if tc.request.Tags != nil {
				assert.ElementsMatch(t, tc.request.Tags, resp.UpdateRisk.Risk.Tags)
			}

			if tc.request.Mitigation != nil {
				assert.Equal(t, *tc.request.Mitigation, *resp.UpdateRisk.Risk.Mitigation)
			}

			if tc.request.Impact != nil {
				assert.Equal(t, *tc.request.Impact, *resp.UpdateRisk.Risk.Impact)
			}

			if tc.request.Likelihood != nil {
				assert.Equal(t, *tc.request.Likelihood, *resp.UpdateRisk.Risk.Likelihood)
			}

			if tc.request.Details != nil {
				assert.Equal(t, tc.request.Details, resp.UpdateRisk.Risk.Details)
			}

			if len(tc.request.AddViewerIDs) > 0 {
				require.Len(t, resp.UpdateRisk.Risk.Viewers, 1)
				for _, edge := range resp.UpdateRisk.Risk.Viewers {
					assert.Equal(t, testUser1.GroupID, edge.ID)
				}

				// ensure the user has access to the risk now
				res, err := suite.client.api.GetRiskByID(anotherAdminUser.UserCtx, risk.ID)
				require.NoError(t, err)
				require.NotEmpty(t, res)
				assert.Equal(t, risk.ID, res.Risk.ID)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteRisk() {
	t := suite.T()

	// create objects to be deleted
	Risk1 := (&RiskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	Risk2 := (&RiskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  Risk1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: Risk1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  Risk1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: Risk2.ID,
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
			resp, err := tc.client.DeleteRisk(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteRisk.DeletedID)
		})
	}
}
