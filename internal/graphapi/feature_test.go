package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/pkg/openlaneclient"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"
	"github.com/theopenlane/utils/ulids"
)

func (suite *GraphTestSuite) TestQueryFeature() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	feature := (&FeatureBuilder{client: suite.client}).MustNew(reqCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenLaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: feature.ID,
			client:  suite.client.api,
			ctx:     reqCtx,
		},
		{
			name:    "happy path using api token",
			queryID: feature.ID,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
		},
		{
			name:    "happy path using personal access token",
			queryID: feature.ID,
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

			resp, err := tc.client.GetFeatureByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Equal(t, tc.queryID, resp.Feature.ID)
			assert.NotEmpty(t, resp.Feature.Name)
			assert.NotEmpty(t, resp.Feature.Description)
			assert.NotEmpty(t, resp.Feature.DisplayName)
		})
	}
}

func (suite *GraphTestSuite) TestQueryFeatures() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	_ = (&FeatureBuilder{client: suite.client}).MustNew(reqCtx, t)
	_ = (&FeatureBuilder{client: suite.client}).MustNew(reqCtx, t)

	otherUser := (&UserBuilder{client: suite.client}).MustNew(reqCtx, t)
	otherCtx, err := userContextWithID(otherUser.ID)
	require.NoError(t, err)

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenLaneClient
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
			name:            "another user, no features should be returned",
			client:          suite.client.api,
			ctx:             otherCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			resp, err := tc.client.GetAllFeatures(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.Features.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateFeature() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	testCases := []struct {
		name        string
		request     openlaneclient.CreateFeatureInput
		client      *openlaneclient.OpenLaneClient
		ctx         context.Context
		allowed     bool
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateFeatureInput{
				Name: "test-feature",
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "happy path, using api token",
			request: openlaneclient.CreateFeatureInput{
				OwnerID: &testOrgID,
				Name:    "meows",
			},
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "happy path, using pat",
			request: openlaneclient.CreateFeatureInput{
				OwnerID: &testOrgID,
				Name:    "woofs",
			},
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateFeatureInput{
				Name:        "mitb",
				DisplayName: lo.ToPtr("Matt is the Best"),
				Enabled:     lo.ToPtr(true),
				Description: lo.ToPtr("Matt is the best feature, hands down!"),
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "do not create if not allowed",
			request: openlaneclient.CreateFeatureInput{
				Name: "test-feature",
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     false,
			expectedErr: "you are not authorized to perform this action: create on feature",
		},
		{
			name: "missing required field",
			request: openlaneclient.CreateFeatureInput{
				DisplayName: lo.ToPtr("Matt is the Best"),
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

			resp, err := tc.client.CreateFeature(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Equal(t, tc.request.Name, resp.CreateFeature.Feature.Name)

			if tc.request.Enabled == nil {
				assert.False(t, resp.CreateFeature.Feature.Enabled)
			} else {
				assert.Equal(t, *tc.request.Enabled, resp.CreateFeature.Feature.Enabled)
			}

			if tc.request.Description == nil {
				assert.Nil(t, resp.CreateFeature.Feature.Description)
			} else {
				assert.Equal(t, *tc.request.Description, *resp.CreateFeature.Feature.Description)
			}

			// Display Name is set to the Name if not provided
			if tc.request.DisplayName == nil {
				assert.Equal(t, tc.request.Name, *resp.CreateFeature.Feature.DisplayName)
			} else {
				assert.Equal(t, *tc.request.DisplayName, *resp.CreateFeature.Feature.DisplayName)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateFeature() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	feature := (&FeatureBuilder{client: suite.client}).MustNew(reqCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateFeatureInput
		client      *openlaneclient.OpenLaneClient
		ctx         context.Context
		allowed     bool
		expectedErr string
	}{
		{
			name: "happy path, update display name",
			request: openlaneclient.UpdateFeatureInput{
				DisplayName: lo.ToPtr("test-feature"),
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "enable feature using api token",
			request: openlaneclient.UpdateFeatureInput{
				Enabled: lo.ToPtr(true),
			},
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "update description using pat",
			request: openlaneclient.UpdateFeatureInput{
				Description: lo.ToPtr("To infinity and beyond!"),
			},
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "not allowed to update",
			request: openlaneclient.UpdateFeatureInput{
				Enabled: lo.ToPtr(false),
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     false,
			expectedErr: "you are not authorized to perform this action: update on feature",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// check for edit permissions on the organization
			mock_fga.CheckAny(t, suite.client.fga, tc.allowed)

			resp, err := tc.client.UpdateFeature(tc.ctx, feature.ID, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.UpdateFeature.Feature.Description)
			}

			if tc.request.DisplayName != nil {
				assert.Equal(t, *tc.request.DisplayName, *resp.UpdateFeature.Feature.DisplayName)
			}

			if tc.request.Enabled != nil {
				assert.Equal(t, *tc.request.Enabled, resp.UpdateFeature.Feature.Enabled)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteFeature() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	feature1 := (&FeatureBuilder{client: suite.client}).MustNew(reqCtx, t)
	feature2 := (&FeatureBuilder{client: suite.client}).MustNew(reqCtx, t)
	feature3 := (&FeatureBuilder{client: suite.client}).MustNew(reqCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenLaneClient
		ctx         context.Context
		allowed     bool
		checkAccess bool
		expectedErr string
	}{
		{
			name:        "not allowed to delete",
			idToDelete:  feature1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: true,
			allowed:     false,
			expectedErr: "you are not authorized to perform this action: delete on feature",
		},
		{
			name:        "happy path, delete feature",
			idToDelete:  feature1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "feature already deleted, not found",
			idToDelete:  feature1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: false,
			allowed:     true,
			expectedErr: "feature not found",
		},
		{
			name:        "happy path, delete feature using api token",
			idToDelete:  feature2.ID,
			client:      suite.client.apiWithToken,
			ctx:         context.Background(),
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "happy path, delete feature using pat",
			idToDelete:  feature3.ID,
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "unknown feature, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: false,
			allowed:     true,
			expectedErr: "feature not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// check for edit permissions on the organization if feature exists
			if tc.checkAccess {
				mock_fga.CheckAny(t, suite.client.fga, tc.allowed)
			}

			resp, err := tc.client.DeleteFeature(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteFeature.DeletedID)
		})
	}
}
