package graphapi_test

import (
	"context"
	"testing"

	mock_fga "github.com/datumforge/fgax/mockery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/ulids"
)

func (suite *GraphTestSuite) TestQueryEntitlementPlanFeature() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	planFeature := (&EntitlementPlanFeatureBuilder{client: suite.client}).MustNew(reqCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.DatumClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: planFeature.ID,
			client:  suite.client.api,
			ctx:     reqCtx,
		},
		{
			name:    "happy path, with api token",
			queryID: planFeature.ID,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
		},
		{
			name:    "happy path, with pat",
			queryID: planFeature.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "not found",
			queryID:  "notfound",
			client:   suite.client.api,
			ctx:      reqCtx,
			errorMsg: "not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			if tc.errorMsg == "" {
				mock_fga.CheckAny(t, suite.client.fga, true)
			}

			resp, err := tc.client.GetEntitlementPlanFeatureByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Equal(t, tc.queryID, resp.EntitlementPlanFeature.ID)
			require.NotEmpty(t, resp.EntitlementPlanFeature.GetFeature())
			assert.Equal(t, planFeature.FeatureID, resp.EntitlementPlanFeature.Feature.ID)
			require.NotEmpty(t, resp.EntitlementPlanFeature.GetPlan())
			assert.Equal(t, planFeature.PlanID, resp.EntitlementPlanFeature.Plan.ID)
			require.NotEmpty(t, resp.EntitlementPlanFeature.GetMetadata())
		})
	}
}

func (suite *GraphTestSuite) TestQueryEntitlementPlanFeatures() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	_ = (&EntitlementPlanFeatureBuilder{client: suite.client}).MustNew(reqCtx, t)
	_ = (&EntitlementPlanFeatureBuilder{client: suite.client}).MustNew(reqCtx, t)

	otherUser := (&UserBuilder{client: suite.client}).MustNew(reqCtx, t)
	otherCtx, err := userContextWithID(otherUser.ID)
	require.NoError(t, err)

	testCases := []struct {
		name            string
		client          *openlaneclient.DatumClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             reqCtx,
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
			name:            "another user, no planFeatures should be returned",
			client:          suite.client.api,
			ctx:             otherCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			resp, err := tc.client.GetAllEntitlementPlanFeatures(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.EntitlementPlanFeatures.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateEntitlementPlanFeature() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	// setup for creation of planFeature
	plan := (&EntitlementPlanBuilder{client: suite.client}).MustNew(reqCtx, t)
	feature1 := (&FeatureBuilder{client: suite.client}).MustNew(reqCtx, t)
	feature2 := (&FeatureBuilder{client: suite.client}).MustNew(reqCtx, t)
	feature3 := (&FeatureBuilder{client: suite.client}).MustNew(reqCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.CreateEntitlementPlanFeatureInput
		client      *openlaneclient.DatumClient
		ctx         context.Context
		allowed     bool
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateEntitlementPlanFeatureInput{
				PlanID:    plan.ID,
				FeatureID: feature1.ID,
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "happy path, all input using api token",
			request: openlaneclient.CreateEntitlementPlanFeatureInput{
				PlanID:    plan.ID,
				FeatureID: feature2.ID,
				Metadata: map[string]interface{}{
					"limit_type": "days",
					"limit":      "30",
				},
			},
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "happy path, all input using personal access token",
			request: openlaneclient.CreateEntitlementPlanFeatureInput{
				OwnerID:   &testPersonalOrgID,
				PlanID:    plan.ID,
				FeatureID: feature3.ID,
				Metadata: map[string]interface{}{
					"limit_type": "days",
					"limit":      "30",
				},
			},
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "already exists",
			request: openlaneclient.CreateEntitlementPlanFeatureInput{
				PlanID:    plan.ID,
				FeatureID: feature2.ID,
				Metadata: map[string]interface{}{
					"limit_type": "days",
					"limit":      "30",
				},
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     true,
			expectedErr: "entitlementplanfeature already exists",
		},
		{
			name: "do not create if not allowed",
			request: openlaneclient.CreateEntitlementPlanFeatureInput{
				PlanID:    plan.ID,
				FeatureID: feature3.ID,
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     false,
			expectedErr: "you are not authorized to perform this action: create on entitlementplanfeature",
		},
		{
			name: "missing required field, feature",
			request: openlaneclient.CreateEntitlementPlanFeatureInput{
				PlanID: plan.ID,
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     true,
			expectedErr: "value is less than the required length",
		},
		{
			name: "missing required field, plan",
			request: openlaneclient.CreateEntitlementPlanFeatureInput{
				FeatureID: feature1.ID,
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     true,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// check for edit permissions on the organization
			mock_fga.CheckAny(t, suite.client.fga, tc.allowed)

			resp, err := tc.client.CreateEntitlementPlanFeature(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Equal(t, tc.request.PlanID, resp.CreateEntitlementPlanFeature.EntitlementPlanFeature.Plan.GetID())
			assert.Equal(t, tc.request.FeatureID, resp.CreateEntitlementPlanFeature.EntitlementPlanFeature.Feature.GetID())

			if tc.request.Metadata != nil {
				assert.Equal(t, tc.request.Metadata, resp.CreateEntitlementPlanFeature.EntitlementPlanFeature.Metadata)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateEntitlementPlanFeature() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	planFeature := (&EntitlementPlanFeatureBuilder{client: suite.client}).MustNew(reqCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateEntitlementPlanFeatureInput
		client      *openlaneclient.DatumClient
		ctx         context.Context
		allowed     bool
		expectedErr string
	}{
		{
			name: "happy path, update metadata",
			request: openlaneclient.UpdateEntitlementPlanFeatureInput{
				Metadata: map[string]interface{}{
					"limit_type": "days",
					"limit":      "15",
				},
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "happy path, update metadata using api token",
			request: openlaneclient.UpdateEntitlementPlanFeatureInput{
				Metadata: map[string]interface{}{
					"limit_type": "days",
					"limit":      "16",
				},
			},
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "happy path, update metadata using personal access token",
			request: openlaneclient.UpdateEntitlementPlanFeatureInput{
				Metadata: map[string]interface{}{
					"limit_type": "days",
					"limit":      "77",
				},
			},
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "not allowed to update",
			request: openlaneclient.UpdateEntitlementPlanFeatureInput{
				Metadata: map[string]interface{}{
					"limit_type": "days",
					"limit":      "65",
				}},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     false,
			expectedErr: "you are not authorized to perform this action: update on entitlementplanfeature",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// check for edit permissions on the organization
			mock_fga.CheckAny(t, suite.client.fga, tc.allowed)

			resp, err := tc.client.UpdateEntitlementPlanFeature(tc.ctx, planFeature.ID, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.request.Metadata, resp.UpdateEntitlementPlanFeature.EntitlementPlanFeature.GetMetadata())
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteEntitlementPlanFeature() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	planFeature1 := (&EntitlementPlanFeatureBuilder{client: suite.client}).MustNew(reqCtx, t)
	planFeature2 := (&EntitlementPlanFeatureBuilder{client: suite.client}).MustNew(reqCtx, t)
	planFeature3 := (&EntitlementPlanFeatureBuilder{client: suite.client}).MustNew(reqCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.DatumClient
		ctx         context.Context
		allowed     bool
		checkAccess bool
		expectedErr string
	}{
		{
			name:        "not allowed to delete",
			idToDelete:  planFeature1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: true,
			allowed:     false,
			expectedErr: "you are not authorized to perform this action: delete on entitlementplanfeature",
		},
		{
			name:        "happy path, delete plan feature",
			idToDelete:  planFeature1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "plan feature already deleted, not found",
			idToDelete:  planFeature1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: false,
			allowed:     true,
			expectedErr: "entitlement_plan_feature not found",
		},
		{
			name:        "happy path, delete plan feature using api token",
			idToDelete:  planFeature2.ID,
			client:      suite.client.apiWithToken,
			ctx:         context.Background(),
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "happy path, delete plan feature using personal access token",
			idToDelete:  planFeature3.ID,
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "unknown plan feature, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: false,
			allowed:     true,
			expectedErr: "entitlement_plan_feature not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// check for edit permissions on the organization if planFeature exists
			if tc.checkAccess {
				mock_fga.CheckAny(t, suite.client.fga, tc.allowed)
			}

			resp, err := tc.client.DeleteEntitlementPlanFeature(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteEntitlementPlanFeature.DeletedID)
		})
	}
}
