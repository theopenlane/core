package graphapi_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQueryMappableDomainByID() {
	t := suite.T()

	mappableDomain := (&MappableDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name         string
		expectedName string
		queryID      string
		client       *openlaneclient.OpenlaneClient
		ctx          context.Context
		errorMsg     string
	}{
		{
			name:         "happy path",
			expectedName: mappableDomain.Name,
			queryID:      mappableDomain.ID,
			client:       suite.client.api,
			ctx:          testUser1.UserCtx,
		},
		{
			name:     "done",
			queryID:  "iddne",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetMappableDomainByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.MappableDomain)

			assert.Equal(t, tc.queryID, resp.MappableDomain.ID)
			assert.Equal(t, tc.expectedName, resp.MappableDomain.Name)
		})
	}
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: mappableDomain.ID}).MustDelete(t.Context(), suite)

}

func (suite *GraphTestSuite) TestQueryMappableDomains() {
	t := suite.T()

	mappableDomain1 := (&MappableDomainBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	mappableDomain2 := (&MappableDomainBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	bologneName := "bologne.io"

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		expectedResults int
		where           *openlaneclient.MappableDomainWhereInput
	}{
		{
			name:            "return all",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 2,
		},
		{
			name:   "query by name",
			client: suite.client.api,
			ctx:    testUser2.UserCtx,
			where: &openlaneclient.MappableDomainWhereInput{
				Name: &mappableDomain1.Name,
			},
			expectedResults: 1,
		},
		{
			name:   "query by name, does not exist",
			client: suite.client.api,
			ctx:    testUser2.UserCtx,
			where: &openlaneclient.MappableDomainWhereInput{
				Name: &bologneName,
			},
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetMappableDomains(tc.ctx, nil, nil, tc.where)

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.MappableDomains)
			assert.Equal(t, int64(tc.expectedResults), resp.MappableDomains.TotalCount)
		})
	}

	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: mappableDomain1.ID}).MustDelete(t.Context(), suite)
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: mappableDomain2.ID}).MustDelete(t.Context(), suite)
}

func (suite *GraphTestSuite) TestMutationCreateMappableDomain() {
	t := suite.T()
	testCases := []struct {
		name        string
		request     openlaneclient.CreateMappableDomainInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path",
			request: openlaneclient.CreateMappableDomainInput{
				Name: "trust.theopenlane.io",
			},
			client: suite.client.api,
			ctx:    systemAdminUser.UserCtx,
		},
		{
			name: "invalid domain",
			request: openlaneclient.CreateMappableDomainInput{
				Name: "!not-a-domain",
			},
			client:      suite.client.api,
			ctx:         systemAdminUser.UserCtx,
			expectedErr: "invalid or unparsable field: url",
		},
		{
			name: "not system admin, unauthorized",
			request: openlaneclient.CreateMappableDomainInput{
				Name: "trust.theopenlane.io",
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
	}
	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateMappableDomain(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Equal(t, tc.request.Name, resp.CreateMappableDomain.MappableDomain.Name)
			(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: resp.CreateMappableDomain.MappableDomain.ID}).MustDelete(tc.ctx, suite)
		})
	}
}

func (suite *GraphTestSuite) TestUpdateMappableDomain() {
	t := suite.T()

	mappableDomain := (&MappableDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
		input    openlaneclient.UpdateMappableDomainInput
	}{
		{
			name:    "happy path",
			queryID: mappableDomain.ID,
			client:  suite.client.api,
			ctx:     systemAdminUser.UserCtx,
			input: openlaneclient.UpdateMappableDomainInput{
				Tags: []string{"hello"},
			},
		},
		{
			name:     "does not exist",
			queryID:  "iddne",
			client:   suite.client.api,
			ctx:      systemAdminUser.UserCtx,
			errorMsg: notFoundErrorMsg,
			input: openlaneclient.UpdateMappableDomainInput{
				Tags: []string{"hello"},
			},
		},
		{
			name:     "not allowed",
			queryID:  mappableDomain.ID,
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
			input: openlaneclient.UpdateMappableDomainInput{
				Tags: []string{"hello"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateMappableDomain(tc.ctx, tc.queryID, tc.input)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: mappableDomain.ID}).MustDelete(t.Context(), suite)

}
