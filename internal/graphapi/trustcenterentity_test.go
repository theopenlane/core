package graphapi_test

import (
	"context"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/shared/objects/storage"
)

func TestQueryTrustcenterEntity(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustcenterEntity := (&TrustcenterEntityBuilder{
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
			queryID: trustcenterEntity.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, using api token",
			queryID: trustcenterEntity.ID,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
		},
		{
			name:    "happy path, using personal access token",
			queryID: trustcenterEntity.ID,
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
			queryID:  trustcenterEntity.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustcenterEntityByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.TrustcenterEntity.ID != "")
			assert.Check(t, resp.TrustcenterEntity.Name != "")
			assert.Check(t, resp.TrustcenterEntity.TrustCenterID != nil)
			assert.Check(t, is.Equal(trustCenter.ID, *resp.TrustcenterEntity.TrustCenterID))
			assert.Check(t, resp.TrustcenterEntity.EntityTypeID != nil)
			entityType, err := suite.client.db.EntityType.Get(testUser1.UserCtx, *resp.TrustcenterEntity.EntityTypeID)
			assert.NilError(t, err)
			assert.Check(t, is.Equal("customer", entityType.Name))
		})
	}

	(&Cleanup[*generated.TrustcenterEntityDeleteOne]{client: suite.client.db.TrustcenterEntity, ID: trustcenterEntity.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryTrustcenterEntities(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustcenterEntity1 := (&TrustcenterEntityBuilder{
		client:        suite.client,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser1.UserCtx, t)
	trustcenterEntity2 := (&TrustcenterEntityBuilder{
		client:        suite.client,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser1.UserCtx, t)

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
			resp, err := tc.client.GetTrustcenterEntities(tc.ctx, nil, nil, nil)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.TrustcenterEntities.Edges, tc.expectedResults))
		})
	}

	(&Cleanup[*generated.TrustcenterEntityDeleteOne]{client: suite.client.db.TrustcenterEntity, IDs: []string{trustcenterEntity1.ID, trustcenterEntity2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateTrustcenterEntity(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	createLogoUpload := func() *graphql.Upload {
		logoFile, err := storage.NewUploadFile("testdata/uploads/logo.png")
		assert.NilError(t, err)
		return &graphql.Upload{
			File:        logoFile.RawFile,
			Filename:    logoFile.OriginalName,
			Size:        logoFile.Size,
			ContentType: logoFile.ContentType,
		}
	}

	testCases := []struct {
		name        string
		request     testclient.CreateTrustcenterEntityInput
		logoFile    *graphql.Upload
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateTrustcenterEntityInput{
				Name: "Test Entity",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, full input",
			request: testclient.CreateTrustcenterEntityInput{
				Name: "Full Test Entity",
				URL:  lo.ToPtr("https://example.com"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, with logo file",
			request: testclient.CreateTrustcenterEntityInput{
				Name: "Entity With Logo",
				URL:  lo.ToPtr("https://example.com"),
			},
			logoFile: createLogoUpload(),
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateTrustcenterEntityInput{
				Name: "API Token Entity",
				URL:  lo.ToPtr("https://example.com"),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateTrustcenterEntityInput{
				Name: "PAT Entity",
				URL:  lo.ToPtr("https://example.com"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "not authorized, view only user",
			request: testclient.CreateTrustcenterEntityInput{
				Name: "Unauthorized Entity",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "invalid URL",
			request: testclient.CreateTrustcenterEntityInput{
				Name: "Invalid URL Entity",
				URL:  lo.ToPtr("not-a-valid-url"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: invalidInputErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.logoFile != nil {
				expectUpload(t, suite.client.mockProvider, []graphql.Upload{*tc.logoFile})
			}

			resp, err := tc.client.CreateTrustcenterEntity(tc.ctx, tc.request, tc.logoFile)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.CreateTrustcenterEntity.TrustcenterEntity.ID != "")
			assert.Check(t, is.Equal(tc.request.Name, resp.CreateTrustcenterEntity.TrustcenterEntity.Name))

			if tc.request.URL != nil {
				assert.Check(t, resp.CreateTrustcenterEntity.TrustcenterEntity.URL != nil)
				assert.Check(t, is.Equal(*tc.request.URL, *resp.CreateTrustcenterEntity.TrustcenterEntity.URL))
			}

			if tc.logoFile != nil {
				assert.Check(t, resp.CreateTrustcenterEntity.TrustcenterEntity.LogoFile != nil)
				assert.Check(t, resp.CreateTrustcenterEntity.TrustcenterEntity.LogoFile.ID != "")
			}

			assert.Check(t, resp.CreateTrustcenterEntity.TrustcenterEntity.EntityTypeID != nil)

			(&Cleanup[*generated.TrustcenterEntityDeleteOne]{client: suite.client.db.TrustcenterEntity, ID: resp.CreateTrustcenterEntity.TrustcenterEntity.ID}).MustDelete(tc.ctx, t)
		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateTrustcenterEntity(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustcenterEntity := (&TrustcenterEntityBuilder{
		client:        suite.client,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser1.UserCtx, t)

	createLogoUpload := func() *graphql.Upload {
		logoFile, err := storage.NewUploadFile("testdata/uploads/logo.png")
		assert.NilError(t, err)
		return &graphql.Upload{
			File:        logoFile.RawFile,
			Filename:    logoFile.OriginalName,
			Size:        logoFile.Size,
			ContentType: logoFile.ContentType,
		}
	}

	testCases := []struct {
		name        string
		request     testclient.UpdateTrustcenterEntityInput
		logoFile    *graphql.Upload
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:    "happy path, minimal input",
			request: testclient.UpdateTrustcenterEntityInput{},
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name: "happy path, full input",
			request: testclient.UpdateTrustcenterEntityInput{
				URL: lo.ToPtr("https://example.com"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, with logo file",
			request: testclient.UpdateTrustcenterEntityInput{
				URL: lo.ToPtr("https://example.com"),
			},
			logoFile: createLogoUpload(),
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
		},
		{
			name: "happy path, using api token",
			request: testclient.UpdateTrustcenterEntityInput{
				URL: lo.ToPtr("https://example.com"),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using pat",
			request: testclient.UpdateTrustcenterEntityInput{
				URL: lo.ToPtr("https://example.com"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:        "not authorized, view only user",
			request:     testclient.UpdateTrustcenterEntityInput{},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "invalid URL",
			request: testclient.UpdateTrustcenterEntityInput{
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

			resp, err := tc.client.UpdateTrustcenterEntity(tc.ctx, trustcenterEntity.ID, tc.request, tc.logoFile)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.UpdateTrustcenterEntity.TrustcenterEntity.ID != "")
			assert.Check(t, is.Equal(trustcenterEntity.ID, resp.UpdateTrustcenterEntity.TrustcenterEntity.ID))

			if tc.request.URL != nil {
				assert.Check(t, resp.UpdateTrustcenterEntity.TrustcenterEntity.URL != nil)
				assert.Check(t, is.Equal(*tc.request.URL, *resp.UpdateTrustcenterEntity.TrustcenterEntity.URL))
			}

			if tc.logoFile != nil {
				assert.Check(t, resp.UpdateTrustcenterEntity.TrustcenterEntity.LogoFile != nil)
				assert.Check(t, resp.UpdateTrustcenterEntity.TrustcenterEntity.LogoFile.ID != "")
			}
		})
	}

	(&Cleanup[*generated.TrustcenterEntityDeleteOne]{client: suite.client.db.TrustcenterEntity, ID: trustcenterEntity.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteTrustcenterEntity(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustcenterEntity1 := (&TrustcenterEntityBuilder{
		client:        suite.client,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser1.UserCtx, t)
	trustcenterEntity2 := (&TrustcenterEntityBuilder{
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
			idToDelete: trustcenterEntity1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:       "happy path, using api token",
			idToDelete: trustcenterEntity2.ID,
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
			idToDelete:  trustcenterEntity1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteTrustcenterEntity(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteTrustcenterEntity.DeletedID))
		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestTrustcenterEntityHookCustomerEntityType(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.CreateTrustcenterEntityInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "creates customer entity type if it doesn't exist",
			request: testclient.CreateTrustcenterEntityInput{
				Name: "Test Entity",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "uses existing customer entity type if it exists",
			request: testclient.CreateTrustcenterEntityInput{
				Name: "Test Entity 2",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := setContext(tc.ctx, suite.client.db)

			resp, err := tc.client.CreateTrustcenterEntity(tc.ctx, tc.request, nil)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, resp.CreateTrustcenterEntity.TrustcenterEntity.EntityTypeID != nil)

			entityType, err := suite.client.db.EntityType.Get(ctx, *resp.CreateTrustcenterEntity.TrustcenterEntity.EntityTypeID)
			assert.NilError(t, err)
			assert.Check(t, is.Equal("customer", entityType.Name))

			(&Cleanup[*generated.TrustcenterEntityDeleteOne]{client: suite.client.db.TrustcenterEntity, ID: resp.CreateTrustcenterEntity.TrustcenterEntity.ID}).MustDelete(tc.ctx, t)
		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}
