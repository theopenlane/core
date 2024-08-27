package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/ulids"
)

func (suite *GraphTestSuite) TestQueryEntity() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	entity := (&EntityBuilder{client: suite.client}).MustNew(reqCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenLaneClient
		ctx      context.Context
		allowed  bool
		expected *ent.Entity
		errorMsg string
	}{
		{
			name:     "happy path entity",
			allowed:  true,
			queryID:  entity.ID,
			client:   suite.client.api,
			ctx:      reqCtx,
			expected: entity,
		},
		{
			name:     "happy path entity, using api token",
			allowed:  true,
			queryID:  entity.ID,
			client:   suite.client.apiWithToken,
			ctx:      context.Background(),
			expected: entity,
		},
		{
			name:     "happy path entity, using personal access token",
			allowed:  true,
			queryID:  entity.ID,
			client:   suite.client.apiWithPAT,
			ctx:      context.Background(),
			expected: entity,
		},
		{
			name:     "no access",
			allowed:  false,
			queryID:  entity.ID,
			client:   suite.client.api,
			ctx:      reqCtx,
			errorMsg: "not authorized",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			mock_fga.CheckAny(t, suite.client.fga, tc.allowed)

			resp, err := tc.client.GetEntityByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Entity)
		})
	}

	// delete created org and entity
	(&EntityCleanup{client: suite.client, ID: entity.ID}).MustDelete(reqCtx, t)
}

func (suite *GraphTestSuite) TestQueryEntities() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	_ = (&EntityBuilder{client: suite.client}).MustNew(reqCtx, t)
	_ = (&EntityBuilder{client: suite.client}).MustNew(reqCtx, t)

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
			name:            "another user, no entities should be returned",
			client:          suite.client.api,
			ctx:             otherCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			resp, err := tc.client.GetAllEntities(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.Entities.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateEntity() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	testCases := []struct {
		name        string
		request     openlaneclient.CreateEntityInput
		client      *openlaneclient.OpenLaneClient
		ctx         context.Context
		allowed     bool
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateEntityInput{
				Name: lo.ToPtr("fraser fir"),
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateEntityInput{
				Name:        lo.ToPtr("mitb"),
				DisplayName: lo.ToPtr("fraser fir"),
				Description: lo.ToPtr("the pine trees of appalachia"),
				Domains:     []string{"https://appalachiatrees.com"},
				Status:      lo.ToPtr("Onboarding"),
				Note: &openlaneclient.CreateNoteInput{
					Text:    "matt is the best",
					OwnerID: &testOrgID,
				},
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "happy path, using api token",
			request: openlaneclient.CreateEntityInput{
				Name: lo.ToPtr("douglas fir"),
			},
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "happy path, using pat",
			request: openlaneclient.CreateEntityInput{
				Name:    lo.ToPtr("blue spruce"),
				OwnerID: &testOrgID,
			},
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "do not create if not allowed",
			request: openlaneclient.CreateEntityInput{
				Name: lo.ToPtr("test-entity"),
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     false,
			expectedErr: "you are not authorized to perform this action: create on entity",
		},
		{
			name: "missing name, but display name provided",
			request: openlaneclient.CreateEntityInput{
				DisplayName: lo.ToPtr("fraser firs"),
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "name already exists",
			request: openlaneclient.CreateEntityInput{
				Name: lo.ToPtr("blue spruce"),
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     true,
			expectedErr: "entity already exists",
		},
		{
			name: "invalid domain(s)",
			request: openlaneclient.CreateEntityInput{
				Name:    lo.ToPtr("stone pines"),
				Domains: []string{"appalachiatrees"},
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     true,
			expectedErr: "invalid or unparsable field: domains",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// check for edit permissions on the organization
			mock_fga.CheckAny(t, suite.client.fga, tc.allowed)

			resp, err := tc.client.CreateEntity(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// Name is set to the Display Name if not provided
			if tc.request.Name == nil {
				assert.Contains(t, *resp.CreateEntity.Entity.Name, *tc.request.DisplayName)
			} else {
				assert.Equal(t, *tc.request.Name, *resp.CreateEntity.Entity.Name)
			}

			// Display Name is set to the Name if not provided
			if tc.request.DisplayName == nil {
				assert.Equal(t, *tc.request.Name, *resp.CreateEntity.Entity.DisplayName)
			} else {
				assert.Equal(t, *tc.request.DisplayName, *resp.CreateEntity.Entity.DisplayName)
			}

			if tc.request.Description == nil {
				assert.Empty(t, resp.CreateEntity.Entity.Description)
			} else {
				assert.Equal(t, *tc.request.Description, *resp.CreateEntity.Entity.Description)
			}

			if tc.request.Domains != nil {
				assert.Equal(t, tc.request.Domains, resp.CreateEntity.Entity.Domains)
			}

			if tc.request.Status != nil {
				assert.Equal(t, tc.request.Status, resp.CreateEntity.Entity.Status)
			} else {
				// default status is active
				assert.Equal(t, "active", *resp.CreateEntity.Entity.Status)
			}

			if tc.request.Note != nil {
				require.Len(t, resp.CreateEntity.Entity.Notes, 1)
				assert.Equal(t, tc.request.Note.Text, resp.CreateEntity.Entity.Notes[0].Text)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateEntity() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	entity := (&EntityBuilder{client: suite.client}).MustNew(reqCtx, t)
	numNotes := 0
	numDomains := 0

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateEntityInput
		client      *openlaneclient.OpenLaneClient
		ctx         context.Context
		allowed     bool
		expectedErr string
	}{
		{
			name: "happy path, update display name",
			request: openlaneclient.UpdateEntityInput{
				DisplayName: lo.ToPtr("blue spruce"),
				Note: &openlaneclient.CreateNoteInput{
					Text: "the pine tree with blue-green colored needles",
				},
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "update description using api token",
			request: openlaneclient.UpdateEntityInput{
				Description: lo.ToPtr("the pine tree with blue-green colored needles"),
			},
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "update notes, domains using personal access token",
			request: openlaneclient.UpdateEntityInput{
				Note: &openlaneclient.CreateNoteInput{
					Text: "the pine tree with blue-green colored needles",
				},
				Domains: []string{"https://appalachiatrees.com"},
			},
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "update status and domain",
			request: openlaneclient.UpdateEntityInput{
				Status:        lo.ToPtr("Onboarding"),
				AppendDomains: []string{"example.com"},
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "not allowed to update",
			request: openlaneclient.UpdateEntityInput{
				Description: lo.ToPtr("pine trees of the west"),
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     false,
			expectedErr: "you are not authorized to perform this action: update on entity",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// check for edit permissions on the organization
			mock_fga.CheckAny(t, suite.client.fga, tc.allowed)

			resp, err := tc.client.UpdateEntity(tc.ctx, entity.ID, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			if tc.request.Description != nil {
				assert.Equal(t, *tc.request.Description, *resp.UpdateEntity.Entity.Description)
			}

			if tc.request.DisplayName != nil {
				assert.Equal(t, *tc.request.DisplayName, *resp.UpdateEntity.Entity.DisplayName)
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, *resp.UpdateEntity.Entity.Status)
			}

			if tc.request.Domains != nil {
				numDomains++
				assert.Contains(t, resp.UpdateEntity.Entity.Domains, tc.request.Domains[0])
				assert.Len(t, resp.UpdateEntity.Entity.Domains, numDomains)
			}

			if tc.request.AppendDomains != nil {
				numDomains++
				assert.Contains(t, resp.UpdateEntity.Entity.Domains, tc.request.AppendDomains[0])
				assert.Len(t, resp.UpdateEntity.Entity.Domains, numDomains)
			}

			if tc.request.Note != nil {
				numNotes++
				require.Len(t, resp.UpdateEntity.Entity.Notes, numNotes)
				assert.Equal(t, tc.request.Note.Text, resp.UpdateEntity.Entity.Notes[0].Text)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteEntity() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	entity1 := (&EntityBuilder{client: suite.client}).MustNew(reqCtx, t)
	entity2 := (&EntityBuilder{client: suite.client}).MustNew(reqCtx, t)
	entity3 := (&EntityBuilder{client: suite.client}).MustNew(reqCtx, t)

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
			idToDelete:  entity1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: true,
			allowed:     false,
			expectedErr: "you are not authorized to perform this action: delete on entity",
		},
		{
			name:        "happy path, delete entity",
			idToDelete:  entity1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "entity already deleted, not found",
			idToDelete:  entity1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: false,
			allowed:     true,
			expectedErr: "entity not found",
		},
		{
			name:        "happy path, delete entity using api token",
			idToDelete:  entity2.ID,
			client:      suite.client.apiWithToken,
			ctx:         context.Background(),
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "happy path, delete entity using personal access token",
			idToDelete:  entity3.ID,
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "unknown entity, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: false,
			allowed:     true,
			expectedErr: "entity not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// check for edit permissions on the organization if entity exists
			if tc.checkAccess {
				mock_fga.CheckAny(t, suite.client.fga, tc.allowed)
			}

			resp, err := tc.client.DeleteEntity(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteEntity.DeletedID)
		})
	}
}
