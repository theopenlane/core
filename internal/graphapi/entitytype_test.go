package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestQueryEntityType(t *testing.T) {
	entityType := (&th.EntityTypeBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path entity type",
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
			queryID: entityType.ID,
		},
		{
			name:    "happy path entity type, using api token",
			client:  suite.Client.APIWithToken,
			ctx:     context.Background(),
			queryID: entityType.ID,
		},
		{
			name:    "happy path entity type, using pat",
			client:  suite.Client.APIWithPAT,
			ctx:     context.Background(),
			queryID: entityType.ID,
		},
		{
			name:     "no access",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			queryID:  entityType.ID,
			errorMsg: th.NotFoundErrorMsg,
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
	(&th.Cleanup[*generated.EntityTypeDeleteOne]{Client: suite.Client.DB.EntityType, ID: entityType.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryEntityTypes(t *testing.T) {
	e1 := (&th.EntityTypeBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	e2 := (&th.EntityTypeBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name         string
		client       *testclient.TestClient
		ctx          context.Context
		shouldSeeNew bool
	}{
		{
			name:         "happy path",
			client:       suite.Client.API,
			ctx:          th.SharedTestUser1.UserCtx,
			shouldSeeNew: true,
		},
		{
			name:         "happy path, using api token",
			client:       suite.Client.APIWithToken,
			ctx:          context.Background(),
			shouldSeeNew: true,
		},
		{
			name:         "happy path, using pat",
			client:       suite.Client.APIWithPAT,
			ctx:          context.Background(),
			shouldSeeNew: true,
		},
		{
			name:         "another user, no new entities should be returned",
			client:       suite.Client.API,
			ctx:          th.SharedTestUser2.UserCtx,
			shouldSeeNew: false,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllEntityTypes(tc.ctx)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			ids := map[string]struct{}{}
			for _, edge := range resp.EntityTypes.Edges {
				if edge == nil || edge.Node == nil {
					continue
				}

				ids[edge.Node.ID] = struct{}{}
			}

			_, e1Visible := ids[e1.ID]
			_, e2Visible := ids[e2.ID]

			if tc.shouldSeeNew {
				assert.Check(t, e1Visible, "expected entity type %s to be visible", e1.ID)
				assert.Check(t, e2Visible, "expected entity type %s to be visible", e2.ID)

				return
			}

			assert.Check(t, !e1Visible, "did not expect entity type %s to be visible", e1.ID)
			assert.Check(t, !e2Visible, "did not expect entity type %s to be visible", e2.ID)
		})
	}

	(&th.Cleanup[*generated.EntityTypeDeleteOne]{Client: suite.Client.DB.EntityType, IDs: []string{e1.ID, e2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
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
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, all input, using api token",
			request: testclient.CreateEntityTypeInput{
				Name: "horses",
			},
			client: suite.Client.APIWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, all input, using pat",
			request: testclient.CreateEntityTypeInput{
				OwnerID: &th.SharedTestUser1.OrganizationID,
				Name:    "bunnies",
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "do not create if not allowed",
			request: testclient.CreateEntityTypeInput{
				Name: "dogs",
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:        "missing required field, name",
			request:     testclient.CreateEntityTypeInput{},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
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

			(&th.Cleanup[*generated.EntityTypeDeleteOne]{Client: suite.Client.DB.EntityType, ID: resp.CreateEntityType.EntityType.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		})
	}
}

func TestMutationUpdateEntityType(t *testing.T) {
	entityType := (&th.EntityTypeBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, update name using api token",
			request: testclient.UpdateEntityTypeInput{
				Name: lo.ToPtr("sphynx"),
			},
			client: suite.Client.APIWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, update name using personal access token",
			request: testclient.UpdateEntityTypeInput{
				Name: lo.ToPtr("persian"),
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "not allowed to update",
			request: testclient.UpdateEntityTypeInput{
				Name: lo.ToPtr("dogs"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
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

	(&th.Cleanup[*generated.EntityTypeDeleteOne]{Client: suite.Client.DB.EntityType, ID: entityType.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteEntityType(t *testing.T) {
	entityType1 := (&th.EntityTypeBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	entityType2 := (&th.EntityTypeBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	entityType3 := (&th.EntityTypeBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:        "not allowed to delete, no access",
			idToDelete:  entityType1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:       "happy path, delete entity type",
			idToDelete: entityType1.ID,
			client:     suite.Client.API,
			ctx:        th.SharedTestUser1.UserCtx,
		},
		{
			name:        "entityType already deleted, not found",
			idToDelete:  entityType1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:       "happy path, delete entity type using api token",
			idToDelete: entityType2.ID,
			client:     suite.Client.APIWithToken,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete entity type using pat",
			idToDelete: entityType3.ID,
			client:     suite.Client.APIWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown entitytype, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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
