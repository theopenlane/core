package graphapi_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
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
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: trustCenter.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, view only user",
			queryID: trustCenter.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:     "trust center not found",
			queryID:  "non-existent-id",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "not authorized to query org",
			queryID:  trustCenter.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenterByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.TrustCenter.ID))
			assert.Check(t, resp.TrustCenter.Slug != nil)
			assert.Check(t, resp.TrustCenter.OwnerID != nil)
			assert.Check(t, is.Equal(testUser1.OrganizationID, *resp.TrustCenter.OwnerID))

			setting := resp.TrustCenter.GetSetting()
			assert.Check(t, setting != nil)
			assert.Check(t, setting.Title != nil)
			assert.Check(t, setting.Overview != nil)
			assert.Check(t, setting.LogoURL != nil)
			assert.Check(t, setting.FaviconURL != nil)
			assert.Check(t, setting.PrimaryColor != nil)
		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryTrustCenters(t *testing.T) {
	cnameRecord := gofakeit.DomainName()
	customDomain := (&CustomDomainBuilder{client: suite.client, CnameRecord: cnameRecord}).MustNew(testUser1.UserCtx, t)
	trustCenter1 := (&TrustCenterBuilder{client: suite.client, CustomDomainID: customDomain.ID}).MustNew(testUser1.UserCtx, t)
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

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
			expectedResults: 1,
		},
		{
			name:            "return all, ro user",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 1,
		},
		{
			name:   "query by org ID",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &openlaneclient.TrustCenterWhereInput{
				OwnerID: &testUser1.OrganizationID,
			},
			expectedResults: 1,
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
			for _, node := range resp.TrustCenters.Edges {
				assert.Check(t, node.Node != nil)
				assert.Check(t, node.Node.Slug != nil)
				assert.Check(t, node.Node.OwnerID != nil)
				assert.Check(t, is.Equal(testUser1.OrganizationID, *node.Node.OwnerID))
				setting := node.Node.GetSetting()
				assert.Check(t, setting != nil)
				assert.Check(t, setting.Title != nil)
				assert.Check(t, setting.Overview != nil)
				assert.Check(t, setting.LogoURL != nil)
				assert.Check(t, setting.FaviconURL != nil)
				assert.Check(t, setting.PrimaryColor != nil)
			}

		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter1.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationCreateTrustCenter(t *testing.T) {
	customDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// Create a trust center first to test the duplicate constraint
	existingTrustCenter := (&TrustCenterBuilder{client: suite.client, OwnerID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.CreateTrustCenterInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path for different organization",
			request: openlaneclient.CreateTrustCenterInput{
				OwnerID: lo.ToPtr(testUser2.OrganizationID),
			},
			client: suite.client.api,
			ctx:    testUser2.UserCtx,
		},
		{
			name: "happy path with custom domain for different organization",
			request: openlaneclient.CreateTrustCenterInput{
				CustomDomainID: &customDomain.ID,
				OwnerID:        lo.ToPtr(testUser2.OrganizationID),
			},
			client: suite.client.api,
			ctx:    testUser2.UserCtx,
		},
		{
			name: "happy path with settings for different organization",
			request: openlaneclient.CreateTrustCenterInput{
				OwnerID: lo.ToPtr(testUser2.OrganizationID),
				CreateTrustCenterSetting: &openlaneclient.CreateTrustCenterSettingInput{
					Title: lo.ToPtr(gofakeit.Name()),
				},
			},
			client: suite.client.api,
			ctx:    testUser2.UserCtx,
		},
		{
			name: "not authorized",
			request: openlaneclient.CreateTrustCenterInput{
				OwnerID: lo.ToPtr(testUser1.OrganizationID),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "duplicate trust center for same organization",
			request: openlaneclient.CreateTrustCenterInput{
				OwnerID: lo.ToPtr(testUser1.OrganizationID),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "trustcenter already exists", // This will be the error when trying to create a duplicate
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

			if tc.request.CustomDomainID != nil {
				assert.Check(t, is.Equal(*tc.request.CustomDomainID, *resp.CreateTrustCenter.TrustCenter.CustomDomainID))
			}

			// Verify slug is the lowercased, alphanumeric version of the org name
			// Get the organization to check its name using a context that allows database access
			dbCtx := setContext(tc.ctx, suite.client.db)
			org, err := suite.client.db.Organization.Get(dbCtx, *resp.CreateTrustCenter.TrustCenter.OwnerID)
			assert.NilError(t, err)

			// Generate expected slug: remove non-alphanumeric chars and lowercase
			expectedSlug := strings.ToLower(regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(org.Name, ""))
			require.NotNil(t, resp.CreateTrustCenter.TrustCenter.Slug)

			assert.Equal(t, expectedSlug, *resp.CreateTrustCenter.TrustCenter.Slug)
			setting := resp.CreateTrustCenter.TrustCenter.GetSetting()
			if tc.request.CreateTrustCenterSetting != nil {
				assert.Equal(t, *tc.request.CreateTrustCenterSetting.Title, *setting.Title)
			} else {
				assert.Equal(t, fmt.Sprintf("%s Trust Center", org.Name), *setting.Title)
				assert.Equal(t, *setting.LogoURL, *org.AvatarRemoteURL)
			}

			// Clean up
			(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: resp.CreateTrustCenter.TrustCenter.ID}).MustDelete(tc.ctx, t)
		})
	}

	// Clean up the existing trust center
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: existingTrustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: customDomain.MappableDomainID}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: customDomain.ID}).MustDelete(systemAdminUser.UserCtx, t)
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
	// Each organization can only have one trust center
	trustCenter1 := (&TrustCenterBuilder{
		client:  suite.client,
		OwnerID: testUser1.OrganizationID,
	}).MustNew(testUser1.UserCtx, t)

	trustCenter2 := (&TrustCenterBuilder{
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
			expectedResults: 1, // Should see only trust centers owned by testUser1
		},
		{
			name:            "happy path - admin user sees all trust centers",
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: 1, // Should see all owned by testUser
		},
		{
			name:            "happy path - view only user",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 1, // Should see only trust centers from their organization
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
				assert.Check(t, len(*firstNode.Slug) != 0)
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
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter1.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationUpdateTrustCenter(t *testing.T) {
	customDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name          string
		trustCenterID string
		request       openlaneclient.UpdateTrustCenterInput
		client        *openlaneclient.OpenlaneClient
		ctx           context.Context
		expectedErr   string
	}{
		{
			name:          "happy path, update tags",
			trustCenterID: trustCenter.ID,
			request: openlaneclient.UpdateTrustCenterInput{
				Tags: []string{"updated", "test"},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:          "happy path, update custom domain",
			trustCenterID: trustCenter.ID,
			request: openlaneclient.UpdateTrustCenterInput{
				CustomDomainID: &customDomain.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:          "happy path, update settings",
			trustCenterID: trustCenter.ID,
			request: openlaneclient.UpdateTrustCenterInput{
				UpdateTrustCenterSetting: &openlaneclient.UpdateTrustCenterSettingInput{
					Title:        lo.ToPtr("Updated Trust Center Title"),
					Overview:     lo.ToPtr("Updated Trust Center Overview"),
					LogoURL:      lo.ToPtr("https://example.com/new-logo.png"),
					FaviconURL:   lo.ToPtr("https://example.com/new-favicon.ico"),
					PrimaryColor: lo.ToPtr("#FF5733"),
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:          "happy path, append tags",
			trustCenterID: trustCenter.ID,
			request: openlaneclient.UpdateTrustCenterInput{
				AppendTags: []string{"appended", "tag"},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:          "happy path, using admin user",
			trustCenterID: trustCenter.ID,
			request: openlaneclient.UpdateTrustCenterInput{
				Tags: []string{"admin", "update"},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:          "happy path, using personal access token",
			trustCenterID: trustCenter.ID,
			request: openlaneclient.UpdateTrustCenterInput{
				Tags: []string{"pat", "update"},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:          "not authorized, view only user",
			trustCenterID: trustCenter.ID,
			request: openlaneclient.UpdateTrustCenterInput{
				Tags: []string{"unauthorized"},
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:          "not authorized, different org user",
			trustCenterID: trustCenter.ID,
			request: openlaneclient.UpdateTrustCenterInput{
				Tags: []string{"unauthorized"},
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:          "trust center not found",
			trustCenterID: "non-existent-id",
			request: openlaneclient.UpdateTrustCenterInput{
				Tags: []string{"test"},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateTrustCenter(tc.ctx, tc.trustCenterID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.trustCenterID, resp.UpdateTrustCenter.TrustCenter.ID))

			// Check updated fields
			if tc.request.Tags != nil {
				assert.Check(t, is.DeepEqual(tc.request.Tags, resp.UpdateTrustCenter.TrustCenter.Tags))
			}

			if tc.request.CustomDomainID != nil {
				assert.Check(t, is.Equal(*tc.request.CustomDomainID, *resp.UpdateTrustCenter.TrustCenter.CustomDomainID))
			}

			if tc.request.UpdateTrustCenterSetting != nil {
				assert.Check(t, is.Equal(*tc.request.UpdateTrustCenterSetting.Title, *resp.UpdateTrustCenter.TrustCenter.GetSetting().Title))
				assert.Check(t, is.Equal(*tc.request.UpdateTrustCenterSetting.Overview, *resp.UpdateTrustCenter.TrustCenter.GetSetting().Overview))
				assert.Check(t, is.Equal(*tc.request.UpdateTrustCenterSetting.LogoURL, *resp.UpdateTrustCenter.TrustCenter.GetSetting().LogoURL))
				assert.Check(t, is.Equal(*tc.request.UpdateTrustCenterSetting.FaviconURL, *resp.UpdateTrustCenter.TrustCenter.GetSetting().FaviconURL))
				assert.Check(t, is.Equal(*tc.request.UpdateTrustCenterSetting.PrimaryColor, *resp.UpdateTrustCenter.TrustCenter.GetSetting().PrimaryColor))
			}
		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: customDomain.MappableDomainID}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: customDomain.ID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationDeleteTrustCenter(t *testing.T) {
	// create objects to be deleted
	trustCenter1 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:       "happy path, delete trust center",
			idToDelete: trustCenter1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "not authorized, different org user",
			idToDelete:  trustCenter2.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "trust center not found",
			idToDelete:  "non-existent-id",
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteTrustCenter(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteTrustCenter.DeletedID))
		})
	}
}
