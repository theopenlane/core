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

func (suite *GraphTestSuite) TestQueryFeature() {
	t := suite.T()

	feature := (&FeatureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: feature.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
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
			name:     notFoundErrorMsg,
			queryID:  "notfound",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
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

	_ = (&FeatureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	_ = (&FeatureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllFeatures(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.Features.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateFeature() {
	t := suite.T()

	testCases := []struct {
		name        string
		request     openlaneclient.CreateFeatureInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateFeatureInput{
				Name: "test-feature",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using api token",
			request: openlaneclient.CreateFeatureInput{
				OwnerID: &testUser1.OrganizationID,
				Name:    "meows",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using pat",
			request: openlaneclient.CreateFeatureInput{
				OwnerID: &testUser1.OrganizationID,
				Name:    "woofs",
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateFeatureInput{
				Name:        "mitb",
				DisplayName: lo.ToPtr("Matt is the Best"),
				Enabled:     lo.ToPtr(true),
				Description: lo.ToPtr("Matt is the best feature, hands down!"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "do not create if not allowed",
			request: openlaneclient.CreateFeatureInput{
				Name: "test-feature",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "missing required field",
			request: openlaneclient.CreateFeatureInput{
				DisplayName: lo.ToPtr("Matt is the Best"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
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

	feature := (&FeatureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateFeatureInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update display name",
			request: openlaneclient.UpdateFeatureInput{
				DisplayName: lo.ToPtr("test-feature"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "enable feature using api token",
			request: openlaneclient.UpdateFeatureInput{
				Enabled: lo.ToPtr(true),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "update description using pat",
			request: openlaneclient.UpdateFeatureInput{
				Description: lo.ToPtr("To infinity and beyond!"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "not allowed to update",
			request: openlaneclient.UpdateFeatureInput{
				Enabled: lo.ToPtr(false),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
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

	feature1 := (&FeatureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	feature2 := (&FeatureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	feature3 := (&FeatureBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not allowed to delete",
			idToDelete:  feature1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:       "happy path, delete feature",
			idToDelete: feature1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "feature already deleted, not found",
			idToDelete:  feature1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete feature using api token",
			idToDelete: feature2.ID,
			client:     suite.client.apiWithToken,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete feature using pat",
			idToDelete: feature3.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown feature, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
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
