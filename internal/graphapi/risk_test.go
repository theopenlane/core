package graphapi_test

import (
	"context"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestQueryRisk(t *testing.T) {
	viewUser := suite.UserBuilder(context.Background(), t)
	suite.AddUserToOrganization(th.SharedTestUser1.UserCtx, t, &viewUser, enums.RoleMember, th.SharedTestUser1.OrganizationID)

	program := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// add viewer user to the program so that they can see/edit risk
	(&th.ProgramMemberBuilder{Client: suite.Client, ProgramID: program.ID,
		UserID: viewUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(th.SharedTestUser1.UserCtx, t)
	anonymousContext := th.CreateAnonymousTrustCenterContext(ulids.New().String(), th.SharedTestUser1.OrganizationID)

	riskIDs := []string{}
	// add test cases for querying the Risk
	testCases := []struct {
		name             string
		queryID          string
		client           *testclient.TestClient
		ctx              context.Context
		errorMsg         string
		hasProgramAccess bool
	}{
		{
			name:             "happy path",
			client:           suite.Client.API,
			ctx:              th.SharedTestUser1.UserCtx,
			hasProgramAccess: true,
		},
		{
			name:     "read only user, same org, no access to the program",
			client:   suite.Client.API,
			ctx:      th.SharedViewOnlyUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:             "admin user, has full access to risks",
			client:           suite.Client.API,
			ctx:              th.SharedAdminUser.UserCtx,
			hasProgramAccess: false, // admins do not automatically have program access, only super admins + owners
		},
		{
			name:             "member user, but has access to the program",
			client:           suite.Client.API,
			ctx:              viewUser.UserCtx,
			hasProgramAccess: true, // member was given program access
		},
		{
			name:             "happy path using personal access token",
			client:           suite.Client.APIWithPAT,
			ctx:              context.Background(),
			hasProgramAccess: true, // this is the owner's PAT
		},
		{
			name:     "risk not found, invalid ID",
			queryID:  "invalid",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "risk not found, using not authorized user",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "no access, anonymous user",
			client:   suite.Client.API,
			ctx:      anonymousContext,
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			// setup the risk if it is not already created
			if tc.queryID == "" {
				resp, err := suite.Client.API.CreateRisk(th.SharedTestUser1.UserCtx,
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

			if tc.hasProgramAccess {
				assert.Assert(t, is.Len(resp.Risk.Programs.Edges, 1))
				assert.Check(t, len(resp.Risk.Programs.Edges[0].Node.ID) != 0)
			} else {
				assert.Assert(t, is.Len(resp.Risk.Programs.Edges, 0))
			}
		})
	}

	// cleanup
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, ID: program.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.RiskDeleteOne]{Client: suite.Client.DB.Risk, IDs: riskIDs}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryRisks(t *testing.T) {
	// create multiple objects to be queried using th.SharedTestUser1
	risk1 := (&th.RiskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	risk2 := (&th.RiskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			name:            "happy path, using read only user of the same org, no programs or groups associated",
			client:          suite.Client.API,
			ctx:             th.SharedViewOnlyUser.UserCtx,
			expectedResults: 0,
		},
		{
			name:            "happy path, has scope using api token",
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
			name:            "another user, no risks should be returned",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser2.UserCtx,
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
	(&th.Cleanup[*generated.RiskDeleteOne]{Client: suite.Client.DB.Risk, IDs: []string{risk1.ID, risk2.ID}}).
		MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationCreateRisk(t *testing.T) {
	program1 := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	program2 := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	programAnotherUser := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser2.UserCtx, t)

	// group to be used for checking access, defaulting to a read only user
	groupMember := (&th.GroupMemberBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	groupMemberCtx := auth.NewTestContextWithOrgID(groupMember.UserID, groupMember.Edges.OrgMembership.OrganizationID)

	// add adminUser to the program so that they can create a risk associated with the program1
	(&th.ProgramMemberBuilder{Client: suite.Client, ProgramID: program1.ID,
		UserID: th.SharedAdminUser.ID, Role: enums.RoleAdmin.String()}).
		MustNew(th.SharedTestUser1.UserCtx, t)

	// create groups to be associated with the risk
	blockedGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	viewerGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	stakeholderGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	delegateGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name           string
		request        testclient.CreateRiskInput
		addGroupToOrg  bool
		client         *testclient.TestClient
		ctx            context.Context
		expectedErr    string
		expectedStatus enums.RiskStatus
		expectedImpact enums.RiskImpact
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateRiskInput{
				Name: "Risk",
			},
			client:         suite.Client.API,
			ctx:            th.SharedTestUser1.UserCtx,
			expectedStatus: enums.RiskIdentified,
			expectedImpact: enums.RiskImpactLow,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateRiskInput{
				Name:          "Another Risk",
				Details:       lo.ToPtr("details of the Risk"),
				Status:        &enums.RiskMitigated,
				BusinessCosts: lo.ToPtr("much money"),
				Impact:        &enums.RiskImpactLow,
				Likelihood:    &enums.RiskLikelihoodHigh,
				Mitigation:    lo.ToPtr("did the thing"),
				Score:         lo.ToPtr(int64(5)),
				ProgramIDs:    []string{program1.ID, program2.ID}, // multiple programs
				StakeholderID: &stakeholderGroup.ID,
				DelegateID:    &delegateGroup.ID,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, hook defaults are applied",
			request: testclient.CreateRiskInput{
				Name:  "Another Risk",
				Score: lo.ToPtr(int64(19)),
			},
			client:         suite.Client.API,
			ctx:            th.SharedTestUser1.UserCtx,
			expectedStatus: enums.RiskIdentified,
			expectedImpact: enums.RiskImpactCritical,
		},
		{
			name: "add groups",
			request: testclient.CreateRiskInput{
				Name:            "Test Risk",
				EditorIDs:       []string{th.SharedTestUser1.GroupID},
				BlockedGroupIDs: []string{blockedGroup.ID},
				ViewerIDs:       []string{viewerGroup.ID},
			},
			client:         suite.Client.API,
			ctx:            th.SharedTestUser1.UserCtx,
			expectedStatus: enums.RiskIdentified,
			expectedImpact: enums.RiskImpactLow,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateRiskInput{
				Name:    "Risk",
				OwnerID: &th.SharedTestUser1.OrganizationID,
			},
			client:         suite.Client.APIWithPAT,
			ctx:            context.Background(),
			expectedStatus: enums.RiskIdentified,
			expectedImpact: enums.RiskImpactLow,
		},
		{
			name: "using api token",
			request: testclient.CreateRiskInput{
				Name: "Risk",
			},
			client:         suite.Client.APIWithToken,
			ctx:            context.Background(),
			expectedStatus: enums.RiskIdentified,
			expectedImpact: enums.RiskImpactLow,
		},
		{
			name: "using api token with only risk read/write scope, missing sla_definition scope",
			request: testclient.CreateRiskInput{
				Name: "Risk",
			},
			// the group and program are specific to the risk query in CreateRisk
			client:         th.SetupAPIToken(th.SharedTestUser1.UserCtx, t, []string{"risk:write", "group:read", "program:read"}),
			ctx:            context.Background(),
			expectedStatus: enums.RiskIdentified,
			expectedImpact: enums.RiskImpactLow,
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateRiskInput{
				Name: "Risk",
			},
			client:      suite.Client.API,
			ctx:         groupMemberCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "user now authorized, added group to org",
			request: testclient.CreateRiskInput{
				Name: "Risk",
			},
			addGroupToOrg:  true,
			client:         suite.Client.API,
			ctx:            groupMemberCtx,
			expectedStatus: enums.RiskIdentified,
			expectedImpact: enums.RiskImpactLow,
		},
		{
			name: "user authorized, they were added to the program",
			request: testclient.CreateRiskInput{
				Name:       "Risk",
				ProgramIDs: []string{program1.ID},
			},
			client:         suite.Client.API,
			ctx:            th.SharedAdminUser.UserCtx,
			expectedStatus: enums.RiskIdentified,
			expectedImpact: enums.RiskImpactLow,
		},
		{
			name: "user authorized, user not authorized to one of the programs",
			request: testclient.CreateRiskInput{
				Name:       "Risk",
				ProgramIDs: []string{program1.ID, program2.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedAdminUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:        "missing required name",
			request:     testclient.CreateRiskInput{},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "user not authorized, no permissions to one of the programs",
			request: testclient.CreateRiskInput{
				Name:       "Risk",
				ProgramIDs: []string{programAnotherUser.ID, program1.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.addGroupToOrg {
				_, err := suite.Client.API.UpdateOrganization(th.SharedTestUser1.UserCtx, th.SharedTestUser1.OrganizationID,
					testclient.UpdateOrganizationInput{
						AddRiskCreatorIDs: []string{groupMember.GroupID},
					}, nil, nil)
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
				assert.Check(t, is.Equal(tc.expectedStatus, *resp.CreateRisk.Risk.Status))
			}

			if tc.request.BusinessCosts != nil {
				assert.Check(t, is.Equal(*tc.request.BusinessCosts, *resp.CreateRisk.Risk.BusinessCosts))
			} else {
				assert.Check(t, is.Len(*resp.CreateRisk.Risk.BusinessCosts, 0))
			}

			if tc.request.Impact != nil {
				assert.Check(t, is.Equal(*tc.request.Impact, *resp.CreateRisk.Risk.Impact))
			} else {
				assert.Check(t, is.Equal(tc.expectedImpact, *resp.CreateRisk.Risk.Impact))
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
					assert.Check(t, is.Equal(th.SharedTestUser1.GroupID, edge.Node.ID))
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

			// check due date based on the sla config, which should be 60 days from now for a risk with low impact
			if tc.request.Impact != nil && *tc.request.Impact == enums.RiskImpactLow {
				assert.Check(t, resp.CreateRisk.Risk.DueDate != nil)
				due := time.Time(*resp.CreateRisk.Risk.DueDate)
				assert.Check(t, due.After(time.Now().Add(59*24*time.Hour)), "due date is not after 59 days from now %s", due.String())
				assert.Check(t, due.Before(time.Now().Add(61*24*time.Hour)), "due date is not before 61 days from now %s", due.String())
			}

			// ensure the org owner has access to the risk that was created by an api token
			if tc.client == suite.Client.APIWithToken {
				res, err := suite.Client.API.GetRiskByID(th.SharedTestUser1.UserCtx, resp.CreateRisk.Risk.ID)
				assert.NilError(t, err)
				assert.Assert(t, res != nil)
				assert.Check(t, is.Equal(resp.CreateRisk.Risk.ID, res.Risk.ID))
			}
		})
	}

	// cleanup
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, IDs: []string{program1.ID, program2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, ID: programAnotherUser.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, IDs: []string{blockedGroup.ID, viewerGroup.ID, groupMember.GroupID, stakeholderGroup.ID, delegateGroup.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)

}

func TestMutationUpdateRisk(t *testing.T) {
	program := (&th.ProgramBuilder{Client: suite.Client, EditorIDs: th.SharedTestUser1.GroupID}).MustNew(th.SharedTestUser1.UserCtx, t)
	risk := (&th.RiskBuilder{Client: suite.Client, ProgramID: program.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// create another viewer user and add them to the same organization and group as th.SharedTestUser1
	// this will allow us to test the group editor/viewer permissions
	anotherViewUser := suite.UserBuilder(context.Background(), t)
	suite.AddUserToOrganization(th.SharedTestUser1.UserCtx, t, &anotherViewUser, enums.RoleMember, th.SharedTestUser1.OrganizationID)

	groupMember := (&th.GroupMemberBuilder{Client: suite.Client, UserID: anotherViewUser.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	stakeholderGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	delegateGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	anotherStakeholderGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// ensure the user does not currently have access to the risk
	_, err := suite.Client.API.GetRiskByID(anotherViewUser.UserCtx, risk.ID)
	assert.ErrorContains(t, err, th.NotFoundErrorMsg)

	createRemediation, err := suite.Client.API.CreateRemediation(th.SharedTestUser1.UserCtx, testclient.CreateRemediationInput{
		Title:   lo.ToPtr("Test Remediation"),
		Summary: lo.ToPtr("Test summary for query"),
		Status:  &enums.RemediationStatusCompleted,
	})
	assert.NilError(t, err)

	createActionPlan, err := suite.Client.API.CreateActionPlan(th.SharedTestUser1.UserCtx, testclient.CreateActionPlanInput{
		Name:  "Test Action Plan",
		Title: "Test Action Plan",
	})
	assert.NilError(t, err)

	createReview, err := suite.Client.API.CreateReview(th.SharedAdminUser.UserCtx, testclient.CreateReviewInput{
		Title:  "Test Review",
		Status: &enums.ReviewStatusCompleted,
	})
	assert.NilError(t, err)

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
				Status:        &enums.RiskOpen,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, add action plan, status should be set to in progress",
			request: testclient.UpdateRiskInput{
				AddActionPlanIDs: []string{createActionPlan.CreateActionPlan.ActionPlan.ID},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: testclient.UpdateRiskInput{
				Tags:              []string{"tag1", "tag2"},
				Mitigation:        lo.ToPtr("Updated mitigation"),
				Impact:            &enums.RiskImpactModerate,
				Likelihood:        &enums.RiskLikelihoodLow,
				StakeholderID:     &anotherStakeholderGroup.ID,
				RiskDecision:      &enums.RiskDecisionTransfer,
				AddRemediationIDs: []string{createRemediation.CreateRemediation.Remediation.ID},
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, set status to mitigated, timestamp should be updated",
			request: testclient.UpdateRiskInput{
				Status:          &enums.RiskMitigated,
				ReviewFrequency: &enums.FrequencyYearly,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, add completed review, last reviewed timestamp should be updated",
			request: testclient.UpdateRiskInput{
				AddReviewIDs: []string{createReview.CreateReview.Review.ID},
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, update review frequency, next review timestamp should be updated",
			request: testclient.UpdateRiskInput{
				ReviewFrequency: &enums.FrequencyMonthly,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, not permissions in same org",
			request: testclient.UpdateRiskInput{
				Likelihood: &enums.RiskLikelihoodLow,
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateRiskInput{
				Likelihood: &enums.RiskLikelihoodLow,
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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

			if tc.request.RiskDecision != nil {
				assert.Check(t, is.Equal(*tc.request.RiskDecision, *resp.UpdateRisk.Risk.RiskDecision))
			}

			if len(tc.request.AddActionPlanIDs) > 0 {
				// risk should be set to in progress when an action plan is added
				assert.Check(t, is.Equal(*resp.UpdateRisk.Risk.Status, enums.RiskInProgress))
			}

			if len(tc.request.AddRemediationIDs) > 0 {
				// risk should be set to mitigated when a completed remediation is added
				assert.Check(t, is.Equal(*resp.UpdateRisk.Risk.Status, enums.RiskMitigated))
			}

			if len(tc.request.AddReviewIDs) > 0 {
				// last reviewed at should be updated when a completed review is added
				assert.Check(t, resp.UpdateRisk.Risk.LastReviewedAt != nil)
				assert.Check(t, resp.UpdateRisk.Risk.NextReviewDueAt != nil)
				due := time.Time(*resp.UpdateRisk.Risk.NextReviewDueAt)
				// ensure the next review due at is approximately one year from now based on the default review frequency
				assert.Check(t, due.After(time.Now().Add(365*24*time.Hour-time.Hour)), "next review due at is not after one year from now %s", due.String())
				assert.Check(t, due.Before(time.Now().Add(365*24*time.Hour+time.Hour)), "next review due at is not before one year and one hour from now %s", due.String())
			}

			if tc.request.Status != nil && *tc.request.Status == enums.RiskMitigated {
				assert.Check(t, resp.UpdateRisk.Risk.MitigatedAt != nil)
			}

			if tc.request.ReviewFrequency != nil {
				// next review due at should be updated when the review frequency is updated
				// this is based on the previous test case that added a completed review, so the next review due at should be approximately one month from now
				if *tc.request.ReviewFrequency == enums.FrequencyMonthly {
					assert.Check(t, resp.UpdateRisk.Risk.NextReviewDueAt != nil)
					// ensure the next review due at is approximately one month from now
					due := time.Time(*resp.UpdateRisk.Risk.NextReviewDueAt)

					assert.Check(t, due.After(time.Now().Add(28*24*time.Hour)))
					assert.Check(t, due.Before(time.Now().Add(31*24*time.Hour)))
				}

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
				res, err := suite.Client.API.GetRiskByID(anotherViewUser.UserCtx, risk.ID)
				assert.NilError(t, err)
				assert.Assert(t, res != nil)
				assert.Check(t, is.Equal(risk.ID, res.Risk.ID))
			}
		})
	}

	// cleanup
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, ID: program.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.RiskDeleteOne]{Client: suite.Client.DB.Risk, ID: risk.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, IDs: []string{stakeholderGroup.ID, delegateGroup.ID, anotherStakeholderGroup.ID, groupMember.GroupID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteRisk(t *testing.T) {
	// create objects to be deleted
	risk1 := (&th.RiskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	risk2 := (&th.RiskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: risk1.ID,
			client:     suite.Client.API,
			ctx:        th.SharedTestUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  risk1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: risk2.ID,
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
	risk1 := (&th.RiskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	risk2 := (&th.RiskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	risk3 := (&th.RiskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	stakeholderGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	delegateGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	riskAnotherUser := (&th.RiskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser2.UserCtx, t)

	// ensure the user does not currently have access to update the risk
	res, err := suite.Client.API.UpdateBulkRisk(th.SharedTestUser2.UserCtx, []string{risk1.ID}, testclient.UpdateRiskInput{
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
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser1.UserCtx,
			expectedUpdatedCount: 3,
		},
		{
			name: "happy path, update risk type and score",
			ids:  []string{risk1.ID, risk2.ID},
			input: testclient.UpdateRiskInput{
				Score: lo.ToPtr(int64(8)),
			},
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser1.UserCtx,
			expectedUpdatedCount: 2,
		},
		{
			name:        "empty ids array",
			ids:         []string{},
			input:       testclient.UpdateRiskInput{Details: lo.ToPtr("test")},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "ids is required",
		},
		{
			name: "mixed success and failure - some risks not authorized",
			ids:  []string{risk1.ID, riskAnotherUser.ID}, // second risk should fail authorization
			input: testclient.UpdateRiskInput{
				Status: &enums.RiskIdentified,
			},
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser1.UserCtx,
			expectedUpdatedCount: 1, // only risk1 should be updated
		},
		{
			name: "update not allowed, no permissions to risks",
			ids:  []string{risk1.ID},
			input: testclient.UpdateRiskInput{
				Status: &enums.RiskArchived,
			},
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser2.UserCtx,
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

	(&th.Cleanup[*generated.RiskDeleteOne]{Client: suite.Client.DB.Risk, IDs: []string{risk1.ID, risk2.ID, risk3.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.RiskDeleteOne]{Client: suite.Client.DB.Risk, ID: riskAnotherUser.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, IDs: []string{stakeholderGroup.ID, delegateGroup.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}
