package graphapi_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQueryEntitlementPlanFeature() {
	t := suite.T()

	planFeature := (&EntitlementPlanFeatureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: planFeature.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, view only user",
			queryID: planFeature.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
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
			ctx:      testUser1.UserCtx,
			errorMsg: "not found",
		},
		{
			name:     "not found",
			queryID:  planFeature.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: "not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
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

	_ = (&EntitlementPlanFeatureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	_ = (&EntitlementPlanFeatureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			name:            "happy path, view only user",
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
			name:            "another user, no planFeatures should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllEntitlementPlanFeatures(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.EntitlementPlanFeatures.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateEntitlementPlanFeature() {
	t := suite.T()

	// setup for creation of planFeature
	plan := (&EntitlementPlanBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	feature1 := (&FeatureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	feature2 := (&FeatureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	feature3 := (&FeatureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.CreateEntitlementPlanFeatureInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateEntitlementPlanFeatureInput{
				PlanID:    plan.ID,
				FeatureID: feature1.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
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
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, all input using personal access token",
			request: openlaneclient.CreateEntitlementPlanFeatureInput{
				OwnerID:   &testUser1.OrganizationID,
				PlanID:    plan.ID,
				FeatureID: feature3.ID,
				Metadata: map[string]interface{}{
					"limit_type": "days",
					"limit":      "30",
				},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
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
			ctx:         testUser1.UserCtx,
			expectedErr: "entitlementplanfeature already exists",
		},
		{
			name: "missing required field, feature",
			request: openlaneclient.CreateEntitlementPlanFeatureInput{
				PlanID: plan.ID,
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "missing required field, plan",
			request: openlaneclient.CreateEntitlementPlanFeatureInput{
				FeatureID: feature1.ID,
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
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

	planFeature := (&EntitlementPlanFeatureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateEntitlementPlanFeatureInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
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
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update metadata using api token",
			request: openlaneclient.UpdateEntitlementPlanFeatureInput{
				Metadata: map[string]interface{}{
					"limit_type": "days",
					"limit":      "16",
				},
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, update metadata using personal access token",
			request: openlaneclient.UpdateEntitlementPlanFeatureInput{
				Metadata: map[string]interface{}{
					"limit_type": "days",
					"limit":      "77",
				},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "not allowed to update",
			request: openlaneclient.UpdateEntitlementPlanFeatureInput{
				Metadata: map[string]interface{}{
					"limit_type": "days",
					"limit":      "65",
				}},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: "you are not authorized to perform this action",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
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

	planFeature1 := (&EntitlementPlanFeatureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	planFeature2 := (&EntitlementPlanFeatureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	planFeature3 := (&EntitlementPlanFeatureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not allowed to delete",
			idToDelete:  planFeature1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: "you are not authorized to perform this action",
		},
		{
			name:       "happy path, delete plan feature",
			idToDelete: planFeature1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "plan feature already deleted, not found",
			idToDelete:  planFeature1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete plan feature using api token",
			idToDelete: planFeature2.ID,
			client:     suite.client.apiWithToken,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete plan feature using personal access token",
			idToDelete: planFeature3.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown plan feature, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
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
