package graphapi_test

import (
	"context"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/iam/auth"
)

func TestQueryTrustCenterEntity(t *testing.T) {
	cleanupTrustCenterData(t)

	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenterEntity := (&TrustCenterEntityBuilder{
		client:        suite.client,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: trustCenterEntity.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, using api token",
			queryID: trustCenterEntity.ID,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
		},
		{
			name:    "happy path, using personal access token",
			queryID: trustCenterEntity.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "not found",
			queryID:  "non-existent-id",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "no access, different org user",
			queryID:  trustCenterEntity.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenterEntityByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.TrustCenterEntity.ID != "")
			assert.Check(t, resp.TrustCenterEntity.Name != "")
			assert.Check(t, resp.TrustCenterEntity.TrustCenterID != nil)
			assert.Check(t, is.Equal(trustCenter.ID, *resp.TrustCenterEntity.TrustCenterID))
			assert.Check(t, resp.TrustCenterEntity.EntityTypeID != nil)
			entityType, err := suite.client.db.EntityType.Get(testUser1.UserCtx, *resp.TrustCenterEntity.EntityTypeID)
			assert.NilError(t, err)
			assert.Check(t, is.Equal("customer", entityType.Name))
		})
	}

	(&Cleanup[*generated.TrustCenterEntityDeleteOne]{client: suite.client.db.TrustCenterEntity, ID: trustCenterEntity.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryTrustCenterEntities(t *testing.T) {
	cleanupTrustCenterData(t)

	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenterEntity1 := (&TrustCenterEntityBuilder{
		client:        suite.client,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser1.UserCtx, t)
	trustCenterEntity2 := (&TrustCenterEntityBuilder{
		client:        suite.client,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser1.UserCtx, t)

	createLogoUpload := logoFileFunc(t)
	logoFile := createLogoUpload()

	expectUpload(t, suite.client.mockProvider, []graphql.Upload{*logoFile})

	entityWithFile, err := suite.client.api.CreateTrustCenterEntity(testUser1.UserCtx, testclient.CreateTrustCenterEntityInput{
		Name: "Entity With File",
	}, logoFile)
	assert.NilError(t, err)
	assert.Assert(t, entityWithFile != nil)
	assert.Assert(t, entityWithFile.CreateTrustCenterEntity.TrustCenterEntity.ID != "")
	assert.Assert(t, entityWithFile.CreateTrustCenterEntity.TrustCenterEntity.LogoFile.ID != "")

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
			expectedResults: 3,
		},
		{
			name:            "happy path, using api token",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: 3,
		},
		{
			name:            "happy path, using pat",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 3,
		},
		{
			name:            "anonymous user can see trust center entities for trust center they have access to",
			client:          suite.client.api,
			ctx:             createAnonymousTrustCenterContext(trustCenter.ID, testUser1.OrganizationID),
			expectedResults: 3,
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
			resp, err := tc.client.GetTrustCenterEntities(tc.ctx, nil, nil, nil)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.TrustCenterEntities.Edges, tc.expectedResults))

			for _, edge := range resp.TrustCenterEntities.Edges {
				if edge.Node.ID == entityWithFile.CreateTrustCenterEntity.TrustCenterEntity.ID {
					assert.Check(t, edge.Node.LogoFile != nil)
					assert.Check(t, edge.Node.LogoFile.ID != "")
				}
			}
		})
	}

	(&Cleanup[*generated.TrustCenterEntityDeleteOne]{client: suite.client.db.TrustCenterEntity, IDs: []string{trustCenterEntity1.ID, trustCenterEntity2.ID, entityWithFile.CreateTrustCenterEntity.TrustCenterEntity.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateTrustCenterEntity(t *testing.T) {
	testUser := suite.userBuilder(t.Context(), t)
	apiClient := suite.setupAPITokenClient(testUser.UserCtx, t)
	patClient := suite.setupPatClient(testUser, t)

	om := (&OrgMemberBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	viewOnlyUserCtx := auth.NewTestContextWithOrgID(om.UserID, testUser.OrganizationID)

	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	createLogoUpload := logoFileFunc(t)

	testCases := []struct {
		name        string
		request     testclient.CreateTrustCenterEntityInput
		logoFile    *graphql.Upload
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateTrustCenterEntityInput{
				Name: "Test Entity",
			},
			client: suite.client.api,
			ctx:    testUser.UserCtx,
		},
		{
			name: "happy path, full input",
			request: testclient.CreateTrustCenterEntityInput{
				Name: "Full Test Entity",
				URL:  lo.ToPtr("https://example.com"),
			},
			client: suite.client.api,
			ctx:    testUser.UserCtx,
		},
		{
			name: "happy path, with logo file",
			request: testclient.CreateTrustCenterEntityInput{
				Name: "Entity With Logo",
				URL:  lo.ToPtr("https://example.com"),
			},
			logoFile: createLogoUpload(),
			client:   suite.client.api,
			ctx:      testUser.UserCtx,
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateTrustCenterEntityInput{
				Name: "API Token Entity",
				URL:  lo.ToPtr("https://example.com"),
			},
			client: apiClient,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateTrustCenterEntityInput{
				Name: "PAT Entity",
				URL:  lo.ToPtr("https://example.com"),
			},
			client: patClient,
			ctx:    context.Background(),
		},
		{
			name: "not authorized, view only user",
			request: testclient.CreateTrustCenterEntityInput{
				Name: "Unauthorized Entity",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "invalid URL",
			request: testclient.CreateTrustCenterEntityInput{
				Name: "Invalid URL Entity",
				URL:  lo.ToPtr("not-a-valid-url"),
			},
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectedErr: invalidInputErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.logoFile != nil {
				expectUpload(t, suite.client.mockProvider, []graphql.Upload{*tc.logoFile})
			}

			resp, err := tc.client.CreateTrustCenterEntity(tc.ctx, tc.request, tc.logoFile)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.CreateTrustCenterEntity.TrustCenterEntity.ID != "")
			assert.Check(t, is.Equal(tc.request.Name, resp.CreateTrustCenterEntity.TrustCenterEntity.Name))

			if tc.request.URL != nil {
				assert.Check(t, resp.CreateTrustCenterEntity.TrustCenterEntity.URL != nil)
				assert.Check(t, is.Equal(*tc.request.URL, *resp.CreateTrustCenterEntity.TrustCenterEntity.URL))
			}

			if tc.logoFile != nil {
				assert.Check(t, resp.CreateTrustCenterEntity.TrustCenterEntity.LogoFile != nil)
				assert.Check(t, resp.CreateTrustCenterEntity.TrustCenterEntity.LogoFile.ID != "")
			}

			assert.Check(t, resp.CreateTrustCenterEntity.TrustCenterEntity.EntityTypeID != nil)

			(&Cleanup[*generated.TrustCenterEntityDeleteOne]{client: suite.client.db.TrustCenterEntity, ID: resp.CreateTrustCenterEntity.TrustCenterEntity.ID}).MustDelete(testUser.UserCtx, t)
		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser.UserCtx, t)
}

func TestMutationUpdateTrustCenterEntity(t *testing.T) {
	cleanupTrustCenterData(t)

	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenterEntity := (&TrustCenterEntityBuilder{
		client:        suite.client,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser1.UserCtx, t)

	createLogoUpload := logoFileFunc(t)

	testCases := []struct {
		name        string
		request     testclient.UpdateTrustCenterEntityInput
		logoFile    *graphql.Upload
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:    "happy path, minimal input",
			request: testclient.UpdateTrustCenterEntityInput{},
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name: "happy path, full input",
			request: testclient.UpdateTrustCenterEntityInput{
				URL: lo.ToPtr("https://example.com"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, with logo file",
			request: testclient.UpdateTrustCenterEntityInput{
				URL: lo.ToPtr("https://example.com"),
			},
			logoFile: createLogoUpload(),
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
		},
		{
			name: "happy path, using api token",
			request: testclient.UpdateTrustCenterEntityInput{
				URL: lo.ToPtr("https://example.com"),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using pat",
			request: testclient.UpdateTrustCenterEntityInput{
				URL: lo.ToPtr("https://example.com"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:        "not authorized, view only user",
			request:     testclient.UpdateTrustCenterEntityInput{},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "invalid URL",
			request: testclient.UpdateTrustCenterEntityInput{
				URL: lo.ToPtr("not-a-valid-url"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: invalidInputErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			if tc.logoFile != nil {
				expectUpload(t, suite.client.mockProvider, []graphql.Upload{*tc.logoFile})
			}

			resp, err := tc.client.UpdateTrustCenterEntity(tc.ctx, trustCenterEntity.ID, tc.request, tc.logoFile)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.UpdateTrustCenterEntity.TrustCenterEntity.ID != "")
			assert.Check(t, is.Equal(trustCenterEntity.ID, resp.UpdateTrustCenterEntity.TrustCenterEntity.ID))

			if tc.request.URL != nil {
				assert.Check(t, resp.UpdateTrustCenterEntity.TrustCenterEntity.URL != nil)
				assert.Check(t, is.Equal(*tc.request.URL, *resp.UpdateTrustCenterEntity.TrustCenterEntity.URL))
			}

			if tc.logoFile != nil {
				assert.Check(t, resp.UpdateTrustCenterEntity.TrustCenterEntity.LogoFile != nil)
				assert.Check(t, resp.UpdateTrustCenterEntity.TrustCenterEntity.LogoFile.ID != "")
			}
		})
	}

	(&Cleanup[*generated.TrustCenterEntityDeleteOne]{client: suite.client.db.TrustCenterEntity, ID: trustCenterEntity.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteTrustCenterEntity(t *testing.T) {
	cleanupTrustCenterData(t)

	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenterEntity1 := (&TrustCenterEntityBuilder{
		client:        suite.client,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser1.UserCtx, t)
	trustCenterEntity2 := (&TrustCenterEntityBuilder{
		client:        suite.client,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:       "happy path, delete trustcenter entity",
			idToDelete: trustCenterEntity1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:       "happy path, using api token",
			idToDelete: trustCenterEntity2.ID,
			client:     suite.client.apiWithToken,
			ctx:        context.Background(),
		},
		{
			name:        "not found",
			idToDelete:  "non-existent-id",
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "no access, different org user",
			idToDelete:  trustCenterEntity1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteTrustCenterEntity(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteTrustCenterEntity.DeletedID))
		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestTrustCenterEntityHookCustomerEntityType(t *testing.T) {
	cleanupTrustCenterData(t)

	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.CreateTrustCenterEntityInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "creates customer entity type if it doesn't exist",
			request: testclient.CreateTrustCenterEntityInput{
				Name: "Test Entity",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "uses existing customer entity type if it exists",
			request: testclient.CreateTrustCenterEntityInput{
				Name: "Test Entity 2",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := setContext(tc.ctx, suite.client.db)

			resp, err := tc.client.CreateTrustCenterEntity(tc.ctx, tc.request, nil)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, resp.CreateTrustCenterEntity.TrustCenterEntity.EntityTypeID != nil)

			entityType, err := suite.client.db.EntityType.Get(ctx, *resp.CreateTrustCenterEntity.TrustCenterEntity.EntityTypeID)
			assert.NilError(t, err)
			assert.Check(t, is.Equal("customer", entityType.Name))

			(&Cleanup[*generated.TrustCenterEntityDeleteOne]{client: suite.client.db.TrustCenterEntity, ID: resp.CreateTrustCenterEntity.TrustCenterEntity.ID}).MustDelete(tc.ctx, t)
		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}
