package graphapi_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryTrustCenterByID(t *testing.T) {
	t.Skip("TrustCenter GraphQL client methods not yet generated - this test will be enabled once the GraphQL client is updated")

	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// TODO: Implement test cases once GraphQL client methods are generated
	// This test should include:
	// - happy path query by ID
	// - view only user access
	// - not found scenarios
	// - unauthorized access scenarios

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryTrustCenters(t *testing.T) {
	cnameRecord := gofakeit.DomainName()
	customDomain := (&CustomDomainBuilder{client: suite.client, CnameRecord: cnameRecord}).MustNew(testUser1.UserCtx, t)
	trustCenter1 := (&TrustCenterBuilder{client: suite.client, CustomDomainID: customDomain.ID}).MustNew(testUser1.UserCtx, t)
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenter3 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	nonExistentSlug := "nonexistent-slug"

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		expectedResults int64
		where           *openlaneclient.TrustCenterWhereInput
	}{
		{
			name:            "return all",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 3,
		},
		{
			name:            "return all, ro user",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 3,
		},
		{
			name:   "query by org ID",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &openlaneclient.TrustCenterWhereInput{
				OwnerID: &testUser1.OrganizationID,
			},
			expectedResults: 2,
		},
		{
			name:   "query by slug",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &openlaneclient.TrustCenterWhereInput{
				Slug: &trustCenter1.Slug,
			},
			expectedResults: 1,
		},
		{
			name:   "query by slug, not found",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &openlaneclient.TrustCenterWhereInput{
				Slug: &nonExistentSlug,
			},
			expectedResults: 0,
		},
		{
			name:   "query by custom domain, slug",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &openlaneclient.TrustCenterWhereInput{
				And: []*openlaneclient.TrustCenterWhereInput{
					{
						Slug: &trustCenter1.Slug,
					},
					{
						HasCustomDomainWith: []*openlaneclient.CustomDomainWhereInput{
							{
								CnameRecord: &cnameRecord,
							},
						},
					},
				},
			},
			expectedResults: 1,
		},
		{
			name:   "query by non existent custom domain, slug",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &openlaneclient.TrustCenterWhereInput{
				And: []*openlaneclient.TrustCenterWhereInput{
					{
						Slug: &trustCenter1.Slug,
					},
					{
						HasCustomDomainWith: []*openlaneclient.CustomDomainWhereInput{
							{
								CnameRecord: lo.ToPtr("non-existent-domain.com"),
							},
						},
					},
				},
			},
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenters(tc.ctx, nil, nil, tc.where)

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.expectedResults, resp.TrustCenters.TotalCount))
		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, IDs: []string{trustCenter1.ID, trustCenter2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter3.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationCreateTrustCenter(t *testing.T) {
	customDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.CreateTrustCenterInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path",
			request: openlaneclient.CreateTrustCenterInput{
				Slug:    "test-trust-center",
				OwnerID: lo.ToPtr(testUser1.OrganizationID),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, adminUser",
			request: openlaneclient.CreateTrustCenterInput{
				Slug:    "admin-trust-center",
				OwnerID: lo.ToPtr(testUser1.OrganizationID),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "happy path with custom domain",
			request: openlaneclient.CreateTrustCenterInput{
				Slug:           "trust-center-with-domain",
				CustomDomainID: &customDomain.ID,
				OwnerID:        lo.ToPtr(testUser1.OrganizationID),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "not authorized",
			request: openlaneclient.CreateTrustCenterInput{
				Slug:    "unauthorized-trust-center",
				OwnerID: lo.ToPtr(testUser1.OrganizationID),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "missing slug",
			request: openlaneclient.CreateTrustCenterInput{
				OwnerID: lo.ToPtr(testUser1.OrganizationID),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "slug",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateTrustCenter(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.request.Slug, resp.CreateTrustCenter.TrustCenter.Slug))
			if tc.request.CustomDomainID != nil {
				assert.Check(t, is.Equal(*tc.request.CustomDomainID, *resp.CreateTrustCenter.TrustCenter.CustomDomainID))
			}

			// Clean up
			(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: resp.CreateTrustCenter.TrustCenter.ID}).MustDelete(tc.ctx, t)
		})
	}

	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: customDomain.MappableDomainID}).MustDelete(t.Context(), t)
	(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: customDomain.ID}).MustDelete(t.Context(), t)
}

func TestGetAllTrustCenters(t *testing.T) {
	// Clean up any existing trust centers
	deletectx := setContext(systemAdminUser.UserCtx, suite.client.db)
	d, err := suite.client.db.TrustCenter.Query().All(deletectx)
	require.Nil(t, err)
	for _, tc := range d {
		suite.client.db.TrustCenter.DeleteOneID(tc.ID).ExecX(deletectx)
	}

	// Create test trust centers with different users
	trustCenter1 := (&TrustCenterBuilder{
		client:  suite.client,
		OwnerID: testUser1.OrganizationID,
	}).MustNew(testUser1.UserCtx, t)

	trustCenter2 := (&TrustCenterBuilder{
		client:  suite.client,
		OwnerID: testUser1.OrganizationID,
	}).MustNew(testUser1.UserCtx, t)

	trustCenter3 := (&TrustCenterBuilder{
		client:  suite.client,
		OwnerID: testUser2.OrganizationID,
	}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		expectedResults int64
		expectedErr     string
	}{
		{
			name:            "happy path - regular user sees only their trust centers",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 2, // Should see only trust centers owned by testUser1
		},
		{
			name:            "happy path - admin user sees all trust centers",
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: 2, // Should see all owned by testUser
		},
		{
			name:            "happy path - view only user",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 2, // Should see only trust centers from their organization
		},
		{
			name:            "happy path - different user sees only their trust centers",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 1, // Should see only trust centers owned by testUser2
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenters(tc.ctx, nil, nil, nil)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.TrustCenters.Edges != nil)

			// Verify the number of results
			assert.Check(t, is.Len(resp.TrustCenters.Edges, int(tc.expectedResults)))
			assert.Check(t, is.Equal(tc.expectedResults, resp.TrustCenters.TotalCount))

			// Verify pagination info
			assert.Check(t, resp.TrustCenters.PageInfo.StartCursor != nil)

			// If we have results, verify the structure of the first result
			if tc.expectedResults > 0 {
				firstNode := resp.TrustCenters.Edges[0].Node
				assert.Check(t, len(firstNode.ID) != 0)
				assert.Check(t, len(firstNode.Slug) != 0)
				assert.Check(t, firstNode.OwnerID != nil)
				assert.Check(t, firstNode.CreatedAt != nil)
			}

			// Verify that users only see trust centers from their organization
			if tc.ctx == testUser1.UserCtx || tc.ctx == viewOnlyUser.UserCtx {
				for _, edge := range resp.TrustCenters.Edges {
					assert.Check(t, is.Equal(testUser1.OrganizationID, *edge.Node.OwnerID))
				}
			} else if tc.ctx == testUser2.UserCtx {
				for _, edge := range resp.TrustCenters.Edges {
					assert.Check(t, is.Equal(testUser2.OrganizationID, *edge.Node.OwnerID))
				}
			}
		})
	}

	// Clean up created trust centers
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, IDs: []string{trustCenter1.ID, trustCenter2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter3.ID}).MustDelete(testUser2.UserCtx, t)
}
