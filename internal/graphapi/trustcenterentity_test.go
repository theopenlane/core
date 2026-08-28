package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestQueryTrustCenterEntity(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAPIClients())
	trustCenter := tcOrg.TrustCenter

	trustCenterEntity := (&th.TrustCenterEntityBuilder{
		Client:        suite.Client,
		TrustCenterID: trustCenter.ID,
	}).MustNew(tcOrg.Owner.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path as admin",
			queryID: trustCenterEntity.ID,
			client:  suite.Client.API,
			ctx:     tcOrg.Admin.UserCtx,
		},
		{
			name:    "happy path, using api token",
			queryID: trustCenterEntity.ID,
			client:  tcOrg.AdminAPIClient,
			ctx:     context.Background(),
		},
		{
			name:    "happy path, using personal access token",
			queryID: trustCenterEntity.ID,
			client:  tcOrg.AdminPatClient,
			ctx:     context.Background(),
		},
		{
			name:     "not found",
			queryID:  "non-existent-id",
			client:   suite.Client.API,
			ctx:      tcOrg.Owner.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "no access, different org user",
			queryID:  trustCenterEntity.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
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
			entityType, err := suite.Client.DB.EntityType.Get(tcOrg.Owner.UserCtx, *resp.TrustCenterEntity.EntityTypeID)
			assert.NilError(t, err)
			assert.Check(t, is.Equal("customer", entityType.Name))
		})
	}

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestQueryTrustCenterEntities(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAPIClients())

	trustCenter := tcOrg.TrustCenter

	(&th.TrustCenterEntityBuilder{
		Client:        suite.Client,
		TrustCenterID: trustCenter.ID,
	}).MustNew(tcOrg.Owner.UserCtx, t)
	(&th.TrustCenterEntityBuilder{
		Client:        suite.Client,
		TrustCenterID: trustCenter.ID,
	}).MustNew(tcOrg.Owner.UserCtx, t)

	createLogoUpload := th.LogoFileFunc(t)
	logoFile := createLogoUpload()

	th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*logoFile})

	entityWithFile, err := suite.Client.API.CreateTrustCenterEntity(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterEntityInput{
		Name: "Entity With File",
	}, logoFile, nil)
	th.RequireNoError(t, err)

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
			client:          suite.Client.API,
			ctx:             tcOrg.Admin.UserCtx,
			expectedResults: 3,
		},
		{
			name:            "happy path, using api token",
			client:          tcOrg.AdminAPIClient,
			ctx:             context.Background(),
			expectedResults: 3,
		},
		{
			name:            "happy path, using pat",
			client:          tcOrg.AdminPatClient,
			ctx:             context.Background(),
			expectedResults: 3,
		},
		{
			name:            "anonymous user can see trust center entities for trust center they have access to",
			client:          suite.Client.API,
			ctx:             th.CreateAnonymousTrustCenterContext(trustCenter.ID, tcOrg.OrganizationID),
			expectedResults: 3,
		},
		{
			name:            "another user, no entities should be returned",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser2.UserCtx,
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

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestMutationCreateTrustCenterEntity(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAllUserTypes())

	createLogoUpload := th.LogoFileFunc(t)

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
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
		},
		{
			name: "happy path, full input",
			request: testclient.CreateTrustCenterEntityInput{
				Name: "Full Test Entity",
				URL:  lo.ToPtr("https://example.com"),
			},
			client: suite.Client.API,
			ctx:    tcOrg.SuperAdmin.UserCtx,
		},
		{
			name: "happy path, with logo file",
			request: testclient.CreateTrustCenterEntityInput{
				Name: "Entity With Logo",
				URL:  lo.ToPtr("https://example.com"),
			},
			logoFile: createLogoUpload(),
			client:   suite.Client.API,
			ctx:      tcOrg.Admin.UserCtx,
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateTrustCenterEntityInput{
				Name: "API Token Entity",
				URL:  lo.ToPtr("https://example.com"),
			},
			client: tcOrg.AdminAPIClient,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateTrustCenterEntityInput{
				Name: "PAT Entity",
				URL:  lo.ToPtr("https://example.com"),
			},
			client: tcOrg.AdminPatClient,
			ctx:    context.Background(),
		},
		{
			name: "not authorized, view only user",
			request: testclient.CreateTrustCenterEntityInput{
				Name: "Unauthorized Entity",
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Member.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "invalid URL",
			request: testclient.CreateTrustCenterEntityInput{
				Name: "Invalid URL Entity",
				URL:  lo.ToPtr("not-a-valid-url"),
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Admin.UserCtx,
			expectedErr: th.InvalidInputErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.logoFile != nil {
				th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*tc.logoFile})
			}

			resp, err := tc.client.CreateTrustCenterEntity(tc.ctx, tc.request, tc.logoFile, nil)
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
		})
	}

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestMutationUpdateTrustCenterEntity(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAPIClients())
	trustCenter := tcOrg.TrustCenter

	trustCenterEntity := (&th.TrustCenterEntityBuilder{
		Client:        suite.Client,
		TrustCenterID: trustCenter.ID,
	}).MustNew(tcOrg.Owner.UserCtx, t)

	createLogoUpload := th.LogoFileFunc(t)

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
			client:  suite.Client.API,
			ctx:     tcOrg.Owner.UserCtx,
		},
		{
			name: "happy path, full input as admin",
			request: testclient.UpdateTrustCenterEntityInput{
				URL: lo.ToPtr("https://example.com"),
			},
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
		},
		{
			name: "happy path, with logo file",
			request: testclient.UpdateTrustCenterEntityInput{
				URL: lo.ToPtr("https://example.com"),
			},
			logoFile: createLogoUpload(),
			client:   suite.Client.API,
			ctx:      tcOrg.Admin.UserCtx,
		},
		{
			name: "happy path, using api token",
			request: testclient.UpdateTrustCenterEntityInput{
				URL: lo.ToPtr("https://example.com"),
			},
			client: tcOrg.AdminAPIClient,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using pat",
			request: testclient.UpdateTrustCenterEntityInput{
				URL: lo.ToPtr("https://example.com"),
			},
			client: tcOrg.AdminPatClient,
			ctx:    context.Background(),
		},
		{
			name:        "not authorized, view only user",
			request:     testclient.UpdateTrustCenterEntityInput{},
			client:      suite.Client.API,
			ctx:         tcOrg.Member.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "invalid URL",
			request: testclient.UpdateTrustCenterEntityInput{
				URL: lo.ToPtr("not-a-valid-url"),
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.InvalidInputErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			if tc.logoFile != nil {
				th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*tc.logoFile})
			}

			resp, err := tc.client.UpdateTrustCenterEntity(tc.ctx, trustCenterEntity.ID, tc.request, tc.logoFile, nil)
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

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestMutationDeleteTrustCenterEntity(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAPIClients())
	trustCenter := tcOrg.TrustCenter

	trustCenterEntity1 := (&th.TrustCenterEntityBuilder{
		Client:        suite.Client,
		TrustCenterID: trustCenter.ID,
	}).MustNew(tcOrg.Owner.UserCtx, t)
	trustCenterEntity2 := (&th.TrustCenterEntityBuilder{
		Client:        suite.Client,
		TrustCenterID: trustCenter.ID,
	}).MustNew(tcOrg.Owner.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:       "happy path, delete trustcenter as admin",
			idToDelete: trustCenterEntity1.ID,
			client:     suite.Client.API,
			ctx:        tcOrg.Admin.UserCtx,
		},
		{
			name:       "happy path, using api token",
			idToDelete: trustCenterEntity2.ID,
			client:     tcOrg.AdminAPIClient,
			ctx:        context.Background(),
		},
		{
			name:        "not found",
			idToDelete:  "non-existent-id",
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "no access, different org user",
			idToDelete:  trustCenterEntity1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestTrustCenterEntityHookCustomerEntityType(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAllUserTypes())

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
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
		},
		{
			name: "uses existing customer entity type if it exists",
			request: testclient.CreateTrustCenterEntityInput{
				Name: "Test Entity 2",
			},
			client: suite.Client.API,
			ctx:    tcOrg.SuperAdmin.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := th.SetContext(tc.ctx, suite.Client.DB)

			resp, err := tc.client.CreateTrustCenterEntity(tc.ctx, tc.request, nil, nil)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, resp.CreateTrustCenterEntity.TrustCenterEntity.EntityTypeID != nil)

			entityType, err := suite.Client.DB.EntityType.Get(ctx, *resp.CreateTrustCenterEntity.TrustCenterEntity.EntityTypeID)
			assert.NilError(t, err)
			assert.Check(t, is.Equal("customer", entityType.Name))

			(&th.Cleanup[*generated.TrustCenterEntityDeleteOne]{Client: suite.Client.DB.TrustCenterEntity, ID: resp.CreateTrustCenterEntity.TrustCenterEntity.ID}).MustDelete(tc.ctx, t)
		})
	}

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}
