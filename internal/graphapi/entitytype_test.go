package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/ent/generated"
)

func TestQueryEntityType(t *testing.T) {
	entityType := (&EntityTypeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
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
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetEntityTypeByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Assert(t, resp.EntityType.ID != "")
		})
	}

	// delete created entityType
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entityType.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryEntityTypes(t *testing.T) {
	e1 := (&EntityTypeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	e2 := (&EntityTypeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			name:            "another user, no new entities should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 1, // 1 is created in the setup
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllEntityTypes(tc.ctx)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.EntityTypes.Edges, tc.expectedResults), "expected %d entity types, got %d", tc.expectedResults, len(resp.EntityTypes.Edges))
		})
	}

	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, IDs: []string{e1.ID, e2.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateEntityType(t *testing.T) {
	testCases := []struct {
		name        string
		request     testclient.CreateEntityTypeInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, all input",
			request: testclient.CreateEntityTypeInput{
				Name: "cats",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input, using api token",
			request: testclient.CreateEntityTypeInput{
				Name: "horses",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, all input, using pat",
			request: testclient.CreateEntityTypeInput{
				OwnerID: &testUser1.OrganizationID,
				Name:    "bunnies",
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "do not create if not allowed",
			request: testclient.CreateEntityTypeInput{
				Name: "dogs",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "missing required field, name",
			request:     testclient.CreateEntityTypeInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateEntityType(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.request.Name, resp.CreateEntityType.EntityType.Name))

			(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: resp.CreateEntityType.EntityType.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}
}

func TestMutationUpdateEntityType(t *testing.T) {
	entityType := (&EntityTypeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.UpdateEntityTypeInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update name",
			request: testclient.UpdateEntityTypeInput{
				Name: lo.ToPtr("maine coons"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update name using api token",
			request: testclient.UpdateEntityTypeInput{
				Name: lo.ToPtr("sphynx"),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, update name using personal access token",
			request: testclient.UpdateEntityTypeInput{
				Name: lo.ToPtr("persian"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "not allowed to update",
			request: testclient.UpdateEntityTypeInput{
				Name: lo.ToPtr("dogs"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateEntityType(tc.ctx, entityType.ID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(*tc.request.Name, resp.UpdateEntityType.EntityType.Name))
		})
	}

	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entityType.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteEntityType(t *testing.T) {
	entityType1 := (&EntityTypeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	entityType2 := (&EntityTypeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	entityType3 := (&EntityTypeBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not allowed to delete",
			idToDelete:  entityType1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "not allowed to delete, no access",
			idToDelete:  entityType1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
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
			expectedErr: notFoundErrorMsg,
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
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteEntityType(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteEntityType.DeletedID))
		})
	}
}
