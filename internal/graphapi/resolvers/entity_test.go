package resolvers_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestQueryEntity(t *testing.T) {
	entity := (&EntityBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	anonymousContext := createAnonymousTrustCenterContext(ulids.New().String(), testUser1.OrganizationID)

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path entity",
			queryID: entity.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path entity, using api token",
			queryID: entity.ID,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
		},
		{
			name:    "happy path entity, using personal access token",
			queryID: entity.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "no access",
			queryID:  entity.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: "entity not found",
		},
		{
			name:     "no access, anonymous user",
			client:   suite.client.api,
			ctx:      anonymousContext,
			queryID:  entity.ID,
			errorMsg: "entity not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetEntityByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.Entity.ID != "")
		})
	}

	// delete created entity
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: entity.ID}).MustDelete(testUser1.UserCtx, t)
	// delete the entityType
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entity.EntityTypeID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryEntities(t *testing.T) {
	entity1 := (&EntityBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	entity2 := (&EntityBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllEntities(tc.ctx)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Entities.Edges, tc.expectedResults))
		})
	}

	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, IDs: []string{entity1.ID, entity2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, IDs: []string{entity1.EntityTypeID, entity2.EntityTypeID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateEntity(t *testing.T) {
	entitiesToDelete := []string{}
	entityTypesToDelete := []string{}

	testCases := []struct {
		name        string
		request     testclient.CreateEntityInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateEntityInput{
				Name: lo.ToPtr("fraser fir"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateEntityInput{
				Name:        lo.ToPtr("mitb"),
				DisplayName: lo.ToPtr("fraser fir"),
				Description: lo.ToPtr("the pine trees of appalachia"),
				Domains:     []string{"https://appalachiatrees.com"},
				Status:      lo.ToPtr("Onboarding"),
				Note: &testclient.CreateNoteInput{
					Text:    "matt is the best",
					OwnerID: &testUser1.OrganizationID,
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateEntityInput{
				Name: lo.ToPtr("douglas fir"),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateEntityInput{
				Name:    lo.ToPtr("blue spruce"),
				OwnerID: &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "do not create if not allowed",
			request: testclient.CreateEntityInput{
				Name: lo.ToPtr("test-entity"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "missing name, but display name provided",
			request: testclient.CreateEntityInput{
				DisplayName: lo.ToPtr("fraser firs"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "name already exists, different casing",
			request: testclient.CreateEntityInput{
				Name: lo.ToPtr("Blue spruce"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "entity already exists",
		},
		{
			name: "invalid domain(s)",
			request: testclient.CreateEntityInput{
				Name:    lo.ToPtr("stone pines"),
				Domains: []string{"appalachiatrees"},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "invalid or unparsable field: domains",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateEntity(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Name is set to the Display Name if not provided
			if tc.request.Name == nil {
				assert.Check(t, is.Contains(*resp.CreateEntity.Entity.Name, *tc.request.DisplayName))
			} else {
				assert.Check(t, is.Equal(*tc.request.Name, *resp.CreateEntity.Entity.Name))
			}

			// Display Name is set to the Name if not provided
			if tc.request.DisplayName == nil {
				assert.Check(t, is.Equal(*tc.request.Name, *resp.CreateEntity.Entity.DisplayName))
			} else {
				assert.Check(t, is.Equal(*tc.request.DisplayName, *resp.CreateEntity.Entity.DisplayName))
			}

			if tc.request.Description == nil {
				assert.Check(t, is.Equal(*resp.CreateEntity.Entity.Description, ""))
			} else {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.CreateEntity.Entity.Description))
			}

			if tc.request.Domains != nil {
				assert.Check(t, is.DeepEqual(tc.request.Domains, resp.CreateEntity.Entity.Domains))
			}

			if tc.request.Status != nil {
				assert.Check(t, is.DeepEqual(tc.request.Status, resp.CreateEntity.Entity.Status))
			} else {
				// default status is active
				assert.Check(t, is.Equal("active", *resp.CreateEntity.Entity.Status))
			}

			if tc.request.Note != nil {
				assert.Check(t, is.Len(resp.CreateEntity.Entity.Notes.Edges, 1))
				assert.Check(t, is.Equal(tc.request.Note.Text, resp.CreateEntity.Entity.Notes.Edges[0].Node.Text))
			}

			entitiesToDelete = append(entitiesToDelete, resp.CreateEntity.Entity.ID)

			if resp.CreateEntity.Entity.EntityType != nil {
				entityTypesToDelete = append(entityTypesToDelete, resp.CreateEntity.Entity.EntityType.ID)
			}
		})
	}

	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, IDs: entitiesToDelete}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, IDs: entityTypesToDelete}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateEntity(t *testing.T) {
	entity := (&EntityBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	numNotes := 0
	numDomains := 0

	testCases := []struct {
		name        string
		request     testclient.UpdateEntityInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update display name",
			request: testclient.UpdateEntityInput{
				DisplayName: lo.ToPtr("blue spruce"),
				Note: &testclient.CreateNoteInput{
					Text: "the pine tree with blue-green colored needles",
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "update description using api token",
			request: testclient.UpdateEntityInput{
				Description: lo.ToPtr("the pine tree with blue-green colored needles"),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "update notes, domains using personal access token",
			request: testclient.UpdateEntityInput{
				Note: &testclient.CreateNoteInput{
					Text: "the pine tree with blue-green colored needles",
				},
				Domains: []string{"https://appalachiatrees.com"},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update status and domain",
			request: testclient.UpdateEntityInput{
				Status:        lo.ToPtr("Onboarding"),
				AppendDomains: []string{"example.com"},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "not allowed to update",
			request: testclient.UpdateEntityInput{
				Description: lo.ToPtr("pine trees of the west"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "not allowed to update, no access to entity",
			request: testclient.UpdateEntityInput{
				Description: lo.ToPtr("pine trees of the west"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: "entity not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateEntity(tc.ctx, entity.ID, tc.request)
			if tc.expectedErr != "" {

				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.UpdateEntity.Entity.Description))
			}

			if tc.request.DisplayName != nil {
				assert.Check(t, is.Equal(*tc.request.DisplayName, *resp.UpdateEntity.Entity.DisplayName))
			}

			if tc.request.Status != nil {
				assert.Check(t, is.Equal(*tc.request.Status, *resp.UpdateEntity.Entity.Status))
			}

			if tc.request.Domains != nil {
				numDomains++

				assert.Check(t, is.Contains(resp.UpdateEntity.Entity.Domains, tc.request.Domains[0]))
				assert.Check(t, is.Len(resp.UpdateEntity.Entity.Domains, numDomains))
			}

			if tc.request.AppendDomains != nil {
				numDomains++

				assert.Check(t, is.Contains(resp.UpdateEntity.Entity.Domains, tc.request.AppendDomains[0]))
				assert.Check(t, is.Len(resp.UpdateEntity.Entity.Domains, numDomains))
			}

			if tc.request.Note != nil {
				numNotes++

				assert.Check(t, is.Len(resp.UpdateEntity.Entity.Notes.Edges, numNotes))
				assert.Check(t, is.Equal(tc.request.Note.Text, resp.UpdateEntity.Entity.Notes.Edges[0].Node.Text))
			}
		})
	}

	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: entity.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entity.EntityTypeID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteEntity(t *testing.T) {
	entity1 := (&EntityBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	entity2 := (&EntityBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	entity3 := (&EntityBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not allowed to delete",
			idToDelete:  entity1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:       "happy path, delete entity",
			idToDelete: entity1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "entity already deleted, not found",
			idToDelete:  entity1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "entity not found",
		},
		{
			name:       "happy path, delete entity using api token",
			idToDelete: entity2.ID,
			client:     suite.client.apiWithToken,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete entity using personal access token",
			idToDelete: entity3.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown entity, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "entity not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteEntity(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteEntity.DeletedID))
		})
	}

	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, IDs: []string{entity1.EntityTypeID, entity2.EntityTypeID, entity3.EntityTypeID}}).MustDelete(testUser1.UserCtx, t)
}
