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

func (suite *GraphTestSuite) TestQueryEntityType() {
	t := suite.T()

	entityType := (&EntityTypeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path entity type",
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			queryID: entityType.ID,
		},
		{
			name:    "happy path entity type, using api token",
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			queryID: entityType.ID,
		},
		{
			name:    "happy path entity type, using pat",
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			queryID: entityType.ID,
		},
		{
			name:     "no access",
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			queryID:  entityType.ID,
			errorMsg: "entity_type not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetEntityTypeByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.EntityType)
		})
	}

	// delete created org and entityType
	(&EntityTypeCleanup{client: suite.client, ID: entityType.ID}).MustDelete(testUser1.UserCtx, t)
}

func (suite *GraphTestSuite) TestQueryEntityTypes() {
	t := suite.T()

	_ = (&EntityTypeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	_ = (&EntityTypeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			expectedResults: 3, // 1 is created in the setup
		},
		{
			name:            "happy path, using api token",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: 3, // 1 is created in the setup
		},
		{
			name:            "happy path, using pat",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 3, // 1 is created in the setup
		},
		{
			name:            "another user, no entities should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 1, // 1 is created in the setup
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllEntityTypes(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.EntityTypes.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateEntityType() {
	t := suite.T()

	testCases := []struct {
		name        string
		request     openlaneclient.CreateEntityTypeInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, all input",
			request: openlaneclient.CreateEntityTypeInput{
				Name: "cats",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input, using api token",
			request: openlaneclient.CreateEntityTypeInput{
				Name: "horses",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, all input, using pat",
			request: openlaneclient.CreateEntityTypeInput{
				OwnerID: &testUser1.OrganizationID,
				Name:    "bunnies",
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "do not create if not allowed",
			request: openlaneclient.CreateEntityTypeInput{
				Name: "dogs",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: "you are not authorized to perform this action: create on entitytype",
		},
		{
			name:        "missing required field, name",
			request:     openlaneclient.CreateEntityTypeInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateEntityType(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Equal(t, tc.request.Name, resp.CreateEntityType.EntityType.Name)
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateEntityType() {
	t := suite.T()

	entityType := (&EntityTypeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateEntityTypeInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update name",
			request: openlaneclient.UpdateEntityTypeInput{
				Name: lo.ToPtr("maine coons"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update name using api token",
			request: openlaneclient.UpdateEntityTypeInput{
				Name: lo.ToPtr("sphynx"),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, update name using personal access token",
			request: openlaneclient.UpdateEntityTypeInput{
				Name: lo.ToPtr("persian"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "not allowed to update",
			request: openlaneclient.UpdateEntityTypeInput{
				Name: lo.ToPtr("dogs"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: "you are not authorized to perform this action: update on entitytype",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateEntityType(tc.ctx, entityType.ID, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, *tc.request.Name, resp.UpdateEntityType.EntityType.Name)
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteEntityType() {
	t := suite.T()

	entityType1 := (&EntityTypeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	entityType2 := (&EntityTypeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	entityType3 := (&EntityTypeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not allowed to delete",
			idToDelete:  entityType1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: "you are not authorized to perform this action: delete on entitytype",
		},
		{
			name:       "happy path, delete entity type",
			idToDelete: entityType1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "entityType already deleted, not found",
			idToDelete:  entityType1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "entity_type not found",
		},
		{
			name:       "happy path, delete entity type using api token",
			idToDelete: entityType2.ID,
			client:     suite.client.apiWithToken,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete entity type using pat",
			idToDelete: entityType3.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown entitytype, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "entity_type not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteEntityType(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteEntityType.DeletedID)
		})
	}
}
