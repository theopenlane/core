package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
)

func TestQueryRisk(t *testing.T) {
	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add adminUser to the program so that they can create a Risk
	(&ProgramMemberBuilder{client: suite.client, ProgramID: program.ID,
		UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(testUser1.UserCtx, t)

	riskIDs := []string{}
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

				assert.NilError(t, err)
				assert.Assert(t, resp != nil)

				tc.queryID = resp.CreateRisk.Risk.ID
				riskIDs = append(riskIDs, tc.queryID)
			}

			resp, err := tc.client.GetRiskByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Nil(resp))

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
	groupMemberCtx := auth.NewTestContextWithOrgID(groupMember.UserID, groupMember.Edges.Orgmembership.OrganizationID)

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
		request       openlaneclient.CreateRiskInput
		addGroupToOrg bool
		client        *openlaneclient.OpenlaneClient
		ctx           context.Context
		expectedErr   string
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
			request: openlaneclient.CreateRiskInput{
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
			request: openlaneclient.CreateRiskInput{
				Name:    "Risk",
				OwnerID: &testUser1.OrganizationID,
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
			ctx:         groupMemberCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user now authorized, added group to org",
			request: openlaneclient.CreateRiskInput{
				Name: "Risk",
			},
			addGroupToOrg: true,
			client:        suite.client.api,
			ctx:           groupMemberCtx,
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
			if tc.addGroupToOrg {
				_, err := suite.client.api.UpdateOrganization(testUser1.UserCtx, testUser1.OrganizationID,
					openlaneclient.UpdateOrganizationInput{
						AddRiskCreatorIDs: []string{groupMember.GroupID},
					}, nil)
				assert.NilError(t, err)
			}

			resp, err := tc.client.CreateRisk(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

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
				assert.Check(t, is.Equal(enums.RiskOpen, *resp.CreateRisk.Risk.Status))
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
				assert.Check(t, is.Equal(*tc.request.StakeholderID, *&resp.CreateRisk.Risk.Stakeholder.ID))
			} else {
				assert.Check(t, is.Nil(resp.CreateRisk.Risk.Stakeholder))
			}

			if tc.request.DelegateID != nil {
				assert.Check(t, is.Equal(*tc.request.DelegateID, *&resp.CreateRisk.Risk.Delegate.ID))
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
	res, err := suite.client.api.GetRiskByID(anotherAdminUser.UserCtx, risk.ID)
	assert.ErrorContains(t, err, notFoundErrorMsg)
	assert.Assert(t, is.Nil(res))

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
			request: openlaneclient.UpdateRiskInput{
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
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

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
		client      *openlaneclient.OpenlaneClient
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
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteRisk.DeletedID))
		})
	}
}
