package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQueryEntitlementPlan() {
	t := suite.T()

	plan := (&EntitlementPlanBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: plan.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, using api token",
			queryID: plan.ID,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
		},
		{
			name:    "happy path, using personal access token",
			queryID: plan.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "not found",
			queryID:  "notfound",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: "not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetEntitlementPlanByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Equal(t, tc.queryID, resp.EntitlementPlan.ID)
			assert.NotEmpty(t, resp.EntitlementPlan.Name)
			assert.NotEmpty(t, resp.EntitlementPlan.Version)
			assert.NotEmpty(t, resp.EntitlementPlan.Description)
			assert.NotEmpty(t, resp.EntitlementPlan.DisplayName)
		})
	}
}

func (suite *GraphTestSuite) TestQueryEntitlementPlans() {
	t := suite.T()

	_ = (&EntitlementPlanBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	_ = (&EntitlementPlanBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			name:            "happy path, view only user should see the same plans",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using api token",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "happy path, using pat",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no plans should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllEntitlementPlans(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.EntitlementPlans.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateEntitlementPlan() {
	t := suite.T()

	testCases := []struct {
		name        string
		request     openlaneclient.CreateEntitlementPlanInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateEntitlementPlanInput{
				Name:    "test-plan",
				Version: "v1",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using api token",
			request: openlaneclient.CreateEntitlementPlanInput{
				Name:    "test-plan-with-api-token",
				Version: "v1",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using personal access token",
			request: openlaneclient.CreateEntitlementPlanInput{
				OwnerID: &testUser1.OrganizationID,
				Name:    "test-plan-with-pat",
				Version: "v1",
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateEntitlementPlanInput{
				Name:        "mitb",
				Version:     "v1",
				DisplayName: lo.ToPtr("Matt is the Best"),
				Description: lo.ToPtr("Matt is the best plan, hands down!"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "missing required field, version",
			request: openlaneclient.CreateEntitlementPlanInput{
				Name: "Matt is the Best",
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateEntitlementPlan(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Equal(t, tc.request.Name, resp.CreateEntitlementPlan.EntitlementPlan.Name)
			assert.Equal(t, tc.request.Version, resp.CreateEntitlementPlan.EntitlementPlan.Version)

			if tc.request.Description == nil {
				assert.Empty(t, resp.CreateEntitlementPlan.EntitlementPlan.Description)
			} else {
				assert.Equal(t, *tc.request.Description, *resp.CreateEntitlementPlan.EntitlementPlan.Description)
			}

			// Display Name is set to the Name if not provided
			if tc.request.DisplayName == nil {
				assert.Equal(t, tc.request.Name, *resp.CreateEntitlementPlan.EntitlementPlan.DisplayName)
			} else {
				assert.Equal(t, *tc.request.DisplayName, *resp.CreateEntitlementPlan.EntitlementPlan.DisplayName)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateEntitlementPlan() {
	t := suite.T()

	plan := (&EntitlementPlanBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateEntitlementPlanInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update display name",
			request: openlaneclient.UpdateEntitlementPlanInput{
				DisplayName: lo.ToPtr("test-plan"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "update description, using api token",
			request: openlaneclient.UpdateEntitlementPlanInput{
				Description: lo.ToPtr("To infinity and beyond!"),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "update description again, using personal access token",
			request: openlaneclient.UpdateEntitlementPlanInput{
				Description: lo.ToPtr("To infinity and beyond, and beyond!"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "not allowed to update",
			request: openlaneclient.UpdateEntitlementPlanInput{
				Description: lo.ToPtr("Howdy, partner!"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: "you are not authorized to perform this action: update on entitlementplan",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateEntitlementPlan(tc.ctx, plan.ID, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.UpdateEntitlementPlan.EntitlementPlan.Description)
			}

			if tc.request.DisplayName != nil {
				assert.Equal(t, *tc.request.DisplayName, *resp.UpdateEntitlementPlan.EntitlementPlan.DisplayName)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteEntitlementPlan() {
	t := suite.T()

	plan1 := (&EntitlementPlanBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	plan2 := (&EntitlementPlanBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	plan3 := (&EntitlementPlanBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not allowed to delete",
			idToDelete:  plan1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: "you are not authorized to perform this action: delete on entitlementplan",
		},
		{
			name:       "happy path, delete plan",
			idToDelete: plan1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "plan already deleted, not found",
			idToDelete:  plan1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "plan not found",
		},
		{
			name:       "happy path, delete plan using api token",
			idToDelete: plan2.ID,
			client:     suite.client.apiWithToken,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete plan using personal access token",
			idToDelete: plan3.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown plan, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "plan not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteEntitlementPlan(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteEntitlementPlan.DeletedID)
		})
	}
}
