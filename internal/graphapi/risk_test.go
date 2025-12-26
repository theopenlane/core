package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
)

func TestQueryRisk(t *testing.T) {
	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can create a Risk
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)
	anonymousContext := createAnonymousTrustCenterContext(ulids.New().String(), testUser1.OrganizationID)

	riskIDs := []string{}
	// add test cases for querying the Risk
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
		{
			name:     "no access, anonymous user",
			client:   suite.client.api,
			ctx:      anonymousContext,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			// setup the risk if it is not already created
			if tc.queryID == "" {
				resp, err := suite.client.api.CreateRisk(testUser1.UserCtx,
					testclient.CreateRiskInput{
						Name:       "Risk",
						ProgramIDs: []string{program.ID},
					})

				assert.NilError(t, err)
				assert.Assert(t, resp != nil)

				tc.queryID = resp.CreateRisk.Risk.ID
				riskIDs = append(riskIDs, tc.queryID)
			}

			resp, err := tc.client.GetRiskByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.Risk.ID))
			assert.Check(t, len(resp.Risk.Name) != 0)

			assert.Assert(t, is.Len(resp.Risk.Programs.Edges, 1))
			assert.Check(t, len(resp.Risk.Programs.Edges[0].Node.ID) != 0)
		})
	}

	// cleanup
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: program.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.RiskDeleteOne]{client: suite.client.db.Risk, IDs: riskIDs}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryRisks(t *testing.T) {
	// create multiple objects to be queried using testUser1
	risk1 := (&RiskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	risk2 := (&RiskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			name:            "another user, no risks should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllRisks(tc.ctx, nil, nil, nil, nil, nil)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Risks.Edges, tc.expectedResults))
		})
	}

	// cleanup
	(&Cleanup[*generated.RiskDeleteOne]{client: suite.client.db.Risk, IDs: []string{risk1.ID, risk2.ID}}).
		MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateRisk(t *testing.T) {
	program1 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	programAnotherUser := (&ProgramBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	// group to be used for checking access, defaulting to a read only user
	groupMember := (&GroupMemberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	groupMemberCtx := auth.NewTestContextWithOrgID(groupMember.UserID, groupMember.Edges.OrgMembership.OrganizationID)

	// add adminUser to the program so that they can create a risk associated with the program1
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program1.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)

	// create groups to be associated with the risk
	blockedGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	viewerGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	stakeholderGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	delegateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name          string
		request       testclient.CreateRiskInput
		addGroupToOrg bool
		client        *testclient.TestClient
		ctx           context.Context
		expectedErr   string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateRiskInput{
				Name: "Risk",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateRiskInput{
				Name:          "Another Risk",
				Details:       lo.ToPtr("details of the Risk"),
				Status:        &enums.RiskMitigated,
				RiskType:      lo.ToPtr("operational"),
				BusinessCosts: lo.ToPtr("much money"),
				Impact:        &enums.RiskImpactLow,
				Likelihood:    &enums.RiskLikelihoodHigh,
				Mitigation:    lo.ToPtr("did the thing"),
				Score:         lo.ToPtr(int64(5)),
				ProgramIDs:    []string{program1.ID, program2.ID}, // multiple programs
				StakeholderID: &stakeholderGroup.ID,
				DelegateID:    &delegateGroup.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "add groups",
			request: testclient.CreateRiskInput{
				Name:            "Test Risk",
				EditorIDs:       []string{testUser1.GroupID},
				BlockedGroupIDs: []string{blockedGroup.ID},
				ViewerIDs:       []string{viewerGroup.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateRiskInput{
				Name:    "Risk",
				OwnerID: &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "using api token",
			request: testclient.CreateRiskInput{
				Name: "Risk",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateRiskInput{
				Name: "Risk",
			},
			client:      suite.client.api,
			ctx:         groupMemberCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user now authorized, added group to org",
			request: testclient.CreateRiskInput{
				Name: "Risk",
			},
			addGroupToOrg: true,
			client:        suite.client.api,
			ctx:           groupMemberCtx,
		},
		{
			name: "user authorized, they were added to the program",
			request: testclient.CreateRiskInput{
				Name:       "Risk",
				ProgramIDs: []string{program1.ID},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "user authorized, user not authorized to one of the programs",
			request: testclient.CreateRiskInput{
				Name:       "Risk",
				ProgramIDs: []string{program1.ID, program2.ID},
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "missing required name",
			request:     testclient.CreateRiskInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "user not authorized, no permissions to one of the programs",
			request: testclient.CreateRiskInput{
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
			if tc.addGroupToOrg {
				_, err := suite.client.api.UpdateOrganization(testUser1.UserCtx, testUser1.OrganizationID,
					testclient.UpdateOrganizationInput{
						AddRiskCreatorIDs: []string{groupMember.GroupID},
					}, nil)
				assert.NilError(t, err)
			}

			resp, err := tc.client.CreateRisk(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// check required fields
			assert.Assert(t, len(resp.CreateRisk.Risk.ID) != 0)
			assert.Check(t, is.Equal(tc.request.Name, resp.CreateRisk.Risk.Name))

			assert.Check(t, len(resp.CreateRisk.Risk.DisplayID) != 0)
			assert.Check(t, is.Contains(resp.CreateRisk.Risk.DisplayID, "RSK-"))

			// ensure the program is set
			if len(tc.request.ProgramIDs) > 0 {
				assert.Check(t, is.Len(resp.CreateRisk.Risk.Programs.Edges, len(tc.request.ProgramIDs)))

				for i, p := range resp.CreateRisk.Risk.Programs.Edges {
					assert.Check(t, is.Equal(tc.request.ProgramIDs[i], p.Node.ID))
				}
			} else {
				assert.Check(t, is.Len(resp.CreateRisk.Risk.Programs.Edges, 0))
			}

			if tc.request.Status != nil {
				assert.Check(t, is.Equal(*tc.request.Status, *resp.CreateRisk.Risk.Status))
			} else {
				assert.Check(t, is.Equal(enums.RiskIdentified, *resp.CreateRisk.Risk.Status))
			}

			if tc.request.RiskType != nil {
				assert.Check(t, is.Equal(*tc.request.RiskType, *resp.CreateRisk.Risk.RiskType))
			} else {
				assert.Check(t, is.Len(*resp.CreateRisk.Risk.RiskType, 0))
			}

			if tc.request.BusinessCosts != nil {
				assert.Check(t, is.Equal(*tc.request.BusinessCosts, *resp.CreateRisk.Risk.BusinessCosts))
			} else {
				assert.Check(t, is.Len(*resp.CreateRisk.Risk.BusinessCosts, 0))
			}

			if tc.request.Impact != nil {
				assert.Check(t, is.Equal(*tc.request.Impact, *resp.CreateRisk.Risk.Impact))
			} else {
				assert.Check(t, is.Equal(enums.RiskImpactModerate, *resp.CreateRisk.Risk.Impact))
			}

			if tc.request.Likelihood != nil {
				assert.Check(t, is.Equal(*tc.request.Likelihood, *resp.CreateRisk.Risk.Likelihood))
			} else {
				assert.Check(t, is.Equal(enums.RiskLikelihoodMid, *resp.CreateRisk.Risk.Likelihood))
			}

			if tc.request.Mitigation != nil {
				assert.Check(t, is.Equal(*tc.request.Mitigation, *resp.CreateRisk.Risk.Mitigation))
			} else {
				assert.Check(t, is.Len(*resp.CreateRisk.Risk.Mitigation, 0))
			}

			if tc.request.Score != nil {
				assert.Check(t, is.Equal(*tc.request.Score, *resp.CreateRisk.Risk.Score))
			} else {
				assert.Check(t, is.Equal(*resp.CreateRisk.Risk.Score, int64(0)))
			}

			if tc.request.Details != nil {
				assert.Check(t, is.Equal(*tc.request.Details, *resp.CreateRisk.Risk.Details))
			} else {
				assert.Check(t, is.Len(*resp.CreateRisk.Risk.Details, 0))
			}

			if len(tc.request.EditorIDs) > 0 {
				assert.Assert(t, is.Len(resp.CreateRisk.Risk.Editors.Edges, 1))
				for _, edge := range resp.CreateRisk.Risk.Editors.Edges {
					assert.Check(t, is.Equal(testUser1.GroupID, edge.Node.ID))
				}
			}

			if len(tc.request.BlockedGroupIDs) > 0 {
				assert.Assert(t, is.Len(resp.CreateRisk.Risk.BlockedGroups.Edges, 1))
				for _, edge := range resp.CreateRisk.Risk.BlockedGroups.Edges {
					assert.Check(t, is.Equal(blockedGroup.ID, edge.Node.ID))
				}
			}

			if len(tc.request.ViewerIDs) > 0 {
				assert.Assert(t, is.Len(resp.CreateRisk.Risk.Viewers.Edges, 1))
				for _, edge := range resp.CreateRisk.Risk.Viewers.Edges {
					assert.Check(t, is.Equal(viewerGroup.ID, edge.Node.ID))
				}
			}

			if tc.request.StakeholderID != nil {
				assert.Check(t, is.Equal(*tc.request.StakeholderID, resp.CreateRisk.Risk.Stakeholder.ID))
			} else {
				assert.Check(t, is.Nil(resp.CreateRisk.Risk.Stakeholder))
			}

			if tc.request.DelegateID != nil {
				assert.Check(t, is.Equal(*tc.request.DelegateID, resp.CreateRisk.Risk.Delegate.ID))
			} else {
				assert.Check(t, is.Nil(resp.CreateRisk.Risk.Delegate))
			}

			// ensure the org owner has access to the risk that was created by an api token
			if tc.client == suite.client.apiWithToken {
				res, err := suite.client.api.GetRiskByID(testUser1.UserCtx, resp.CreateRisk.Risk.ID)
				assert.NilError(t, err)
				assert.Assert(t, res != nil)
				assert.Check(t, is.Equal(resp.CreateRisk.Risk.ID, res.Risk.ID))
			}
		})
	}

	// cleanup
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, IDs: []string{program1.ID, program2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: programAnotherUser.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{blockedGroup.ID, viewerGroup.ID, groupMember.GroupID, stakeholderGroup.ID, delegateGroup.ID}}).MustDelete(testUser1.UserCtx, t)

}

func TestMutationUpdateRisk(t *testing.T) {
	program := (&ProgramBuilder{client: suite.client, EditorIDs: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)
	risk := (&RiskBuilder{client: suite.client, ProgramID: program.ID}).MustNew(testUser1.UserCtx, t)

	// create another admin user and add them to the same organization and group as testUser1
	// this will allow us to test the group editor/viewer permissions
	anotherAdminUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser1.UserCtx, t, &anotherAdminUser, enums.RoleAdmin, testUser1.OrganizationID)

	groupMember := (&GroupMemberBuilder{client: suite.client, UserID: anotherAdminUser.ID}).MustNew(testUser1.UserCtx, t)

	stakeholderGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	delegateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	anotherStakeholderGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// ensure the user does not currently have access to the risk
	_, err := suite.client.api.GetRiskByID(anotherAdminUser.UserCtx, risk.ID)
	assert.ErrorContains(t, err, notFoundErrorMsg)

	testCases := []struct {
		name        string
		request     testclient.UpdateRiskInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: testclient.UpdateRiskInput{
				Details:       lo.ToPtr("Updated details"),
				AddViewerIDs:  []string{groupMember.GroupID},
				StakeholderID: &stakeholderGroup.ID,
				DelegateID:    &delegateGroup.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: testclient.UpdateRiskInput{
				Status:        &enums.RiskArchived,
				Tags:          []string{"tag1", "tag2"},
				Mitigation:    lo.ToPtr("Updated mitigation"),
				Impact:        &enums.RiskImpactModerate,
				Likelihood:    &enums.RiskLikelihoodLow,
				StakeholderID: &anotherStakeholderGroup.ID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, not permissions in same org",
			request: testclient.UpdateRiskInput{
				Likelihood: &enums.RiskLikelihoodLow,
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateRiskInput{
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
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.Status != nil {
				assert.Check(t, is.Equal(*tc.request.Status, *resp.UpdateRisk.Risk.Status))
			}

			if tc.request.Tags != nil {
				assert.DeepEqual(t, tc.request.Tags, resp.UpdateRisk.Risk.Tags)
			}

			if tc.request.Mitigation != nil {
				assert.Check(t, is.Equal(*tc.request.Mitigation, *resp.UpdateRisk.Risk.Mitigation))
			}

			if tc.request.Impact != nil {
				assert.Check(t, is.Equal(*tc.request.Impact, *resp.UpdateRisk.Risk.Impact))
			}

			if tc.request.Likelihood != nil {
				assert.Check(t, is.Equal(*tc.request.Likelihood, *resp.UpdateRisk.Risk.Likelihood))
			}

			if tc.request.Details != nil {
				assert.Check(t, is.DeepEqual(tc.request.Details, resp.UpdateRisk.Risk.Details))
			}

			if tc.request.Score != nil {
				assert.Check(t, is.Equal(*tc.request.Score, *resp.UpdateRisk.Risk.Score))
			}

			if len(tc.request.AddViewerIDs) > 0 {
				assert.Assert(t, is.Len(resp.UpdateRisk.Risk.Viewers.Edges, 1))
				found := false
				for _, edge := range resp.UpdateRisk.Risk.Viewers.Edges {
					if edge.Node.ID == tc.request.AddViewerIDs[0] {
						found = true
						break
					}
				}

				assert.Check(t, found)

				// ensure the user has access to the risk now
				res, err := suite.client.api.GetRiskByID(anotherAdminUser.UserCtx, risk.ID)
				assert.NilError(t, err)
				assert.Assert(t, res != nil)
				assert.Check(t, is.Equal(risk.ID, res.Risk.ID))
			}
		})
	}

	// cleanup
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: program.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.RiskDeleteOne]{client: suite.client.db.Risk, ID: risk.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{stakeholderGroup.ID, delegateGroup.ID, anotherStakeholderGroup.ID, groupMember.GroupID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteRisk(t *testing.T) {
	// create objects to be deleted
	risk1 := (&RiskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	risk2 := (&RiskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  risk1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: risk1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  risk1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: risk2.ID,
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
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteRisk.DeletedID))
		})
	}
}

func TestMutationUpdateBulkRisk(t *testing.T) {
	// create risks to be updated
	risk1 := (&RiskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	risk2 := (&RiskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	risk3 := (&RiskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	stakeholderGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	delegateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	riskAnotherUser := (&RiskBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	// ensure the user does not currently have access to update the risk
	res, err := suite.client.api.UpdateBulkRisk(testUser2.UserCtx, []string{risk1.ID}, testclient.UpdateRiskInput{
		Status: lo.ToPtr(enums.RiskArchived),
	})

	assert.Assert(t, is.Nil(err))
	// make sure nothing was updated
	assert.Equal(t, len(res.UpdateBulkRisk.Risks), 0)

	testCases := []struct {
		name                 string
		ids                  []string
		input                testclient.UpdateRiskInput
		client               *testclient.TestClient
		ctx                  context.Context
		expectedErr          string
		expectedUpdatedCount int
	}{
		{
			name: "happy path, update multiple risks with same fields",
			ids:  []string{risk1.ID, risk2.ID, risk3.ID},
			input: testclient.UpdateRiskInput{
				Details: lo.ToPtr("Updated details for all risks"),
				Impact:  &enums.RiskImpactModerate,
			},
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
			expectedUpdatedCount: 3,
		},
		{
			name: "happy path, update risk type and score",
			ids:  []string{risk1.ID, risk2.ID},
			input: testclient.UpdateRiskInput{
				RiskType: lo.ToPtr("Financial"),
				Score:    lo.ToPtr(int64(8)),
			},
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
			expectedUpdatedCount: 2,
		},
		{
			name:        "empty ids array",
			ids:         []string{},
			input:       testclient.UpdateRiskInput{Details: lo.ToPtr("test")},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "ids is required",
		},
		{
			name: "mixed success and failure - some risks not authorized",
			ids:  []string{risk1.ID, riskAnotherUser.ID}, // second risk should fail authorization
			input: testclient.UpdateRiskInput{
				Status: &enums.RiskIdentified,
			},
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
			expectedUpdatedCount: 1, // only risk1 should be updated
		},
		{
			name: "update not allowed, no permissions to risks",
			ids:  []string{risk1.ID},
			input: testclient.UpdateRiskInput{
				Status: &enums.RiskArchived,
			},
			client:               suite.client.api,
			ctx:                  testUser2.UserCtx,
			expectedUpdatedCount: 0, // should not find any risks to update
		},
	}

	for _, tc := range testCases {
		t.Run("Bulk Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateBulkRisk(tc.ctx, tc.ids, tc.input)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.UpdateBulkRisk.Risks, tc.expectedUpdatedCount))
			assert.Check(t, is.Len(resp.UpdateBulkRisk.UpdatedIDs, tc.expectedUpdatedCount))

			riskMap := make(map[string]*testclient.UpdateBulkRisk_UpdateBulkRisk_Risks)
			for _, risk := range resp.UpdateBulkRisk.Risks {
				riskMap[risk.ID] = risk
			}

			for _, expectedID := range tc.ids {
				responseRisk, found := riskMap[expectedID]
				if !found {
					continue
				}

				if tc.input.Details != nil {
					assert.Check(t, is.DeepEqual(tc.input.Details, responseRisk.Details))
				}

				if tc.input.Status != nil {
					assert.Check(t, is.Equal(*tc.input.Status, *responseRisk.Status))
				}

				if tc.input.Tags != nil {
					assert.Check(t, is.DeepEqual(tc.input.Tags, responseRisk.Tags))
				}

				if tc.input.Impact != nil {
					assert.Check(t, is.Equal(*tc.input.Impact, *responseRisk.Impact))
				}

				if tc.input.Likelihood != nil {
					assert.Check(t, is.Equal(*tc.input.Likelihood, *responseRisk.Likelihood))
				}

				if tc.input.Mitigation != nil {
					assert.Check(t, is.Equal(*tc.input.Mitigation, *responseRisk.Mitigation))
				}

				if tc.input.BusinessCosts != nil {
					assert.Check(t, is.Equal(*tc.input.BusinessCosts, *responseRisk.BusinessCosts))
				}

				if tc.input.Score != nil {
					assert.Check(t, is.Equal(*tc.input.Score, *responseRisk.Score))
				}

				if tc.input.StakeholderID != nil {
					assert.Check(t, responseRisk.Stakeholder != nil)
					assert.Check(t, is.Equal(*tc.input.StakeholderID, responseRisk.Stakeholder.ID))
				}

				if tc.input.DelegateID != nil {
					assert.Check(t, responseRisk.Delegate != nil)
					assert.Check(t, is.Equal(*tc.input.DelegateID, responseRisk.Delegate.ID))
				}

				if tc.input.RiskType != nil {
					assert.Check(t, is.Equal(*tc.input.RiskType, *responseRisk.RiskType))
				}
			}

			for _, updatedID := range resp.UpdateBulkRisk.UpdatedIDs {
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

	(&Cleanup[*generated.RiskDeleteOne]{client: suite.client.db.Risk, IDs: []string{risk1.ID, risk2.ID, risk3.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.RiskDeleteOne]{client: suite.client.db.Risk, ID: riskAnotherUser.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{stakeholderGroup.ID, delegateGroup.ID}}).MustDelete(testUser1.UserCtx, t)
}
