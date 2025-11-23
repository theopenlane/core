package graphapi_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/customdomain"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryTrustCenterByID(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
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

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.TrustCenter.ID))
			assert.Check(t, resp.TrustCenter.Slug != nil)
			assert.Check(t, resp.TrustCenter.OwnerID != nil)
			assert.Check(t, is.Equal(testUser1.OrganizationID, *resp.TrustCenter.OwnerID))

			setting := resp.TrustCenter.GetSetting()
			assert.Assert(t, setting != nil)
			assert.Check(t, setting.Title != nil)
			assert.Check(t, setting.Overview != nil)
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
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int64
		where           *testclient.TrustCenterWhereInput
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
			where: &testclient.TrustCenterWhereInput{
				OwnerID: &testUser1.OrganizationID,
			},
			expectedResults: 1,
		},
		{
			name:   "query by slug",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &testclient.TrustCenterWhereInput{
				Slug: &trustCenter1.Slug,
			},
			expectedResults: 1,
		},
		{
			name:   "query by slug, not found",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &testclient.TrustCenterWhereInput{
				Slug: &nonExistentSlug,
			},
			expectedResults: 0,
		},
		{
			name:   "query by custom domain, slug",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &testclient.TrustCenterWhereInput{
				And: []*testclient.TrustCenterWhereInput{
					{
						Slug: &trustCenter1.Slug,
					},
					{
						HasCustomDomainWith: []*testclient.CustomDomainWhereInput{
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
			where: &testclient.TrustCenterWhereInput{
				And: []*testclient.TrustCenterWhereInput{
					{
						Slug: &trustCenter1.Slug,
					},
					{
						HasCustomDomainWith: []*testclient.CustomDomainWhereInput{
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
				assert.Check(t, setting.PrimaryColor != nil)
			}

		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter1.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationCreateTrustCenter(t *testing.T) {
	customDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	customDomainAnotherOrg := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// Create a trust center first to test the duplicate constraint
	existingTrustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.CreateTrustCenterInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:    "happy path for different organization",
			request: testclient.CreateTrustCenterInput{},
			client:  suite.client.api,
			ctx:     testUser2.UserCtx,
		},
		{
			name: "custom domain for different organization should error",
			request: testclient.CreateTrustCenterInput{
				CustomDomainID: &customDomainAnotherOrg.ID,
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "custom domain setting",
			request: testclient.CreateTrustCenterInput{
				CustomDomainID: &customDomain.ID,
			},
			client: suite.client.api,
			ctx:    testUser2.UserCtx,
		},
		{
			name: "happy path with settings for different organization",
			request: testclient.CreateTrustCenterInput{
				CreateTrustCenterSetting: &testclient.CreateTrustCenterSettingInput{
					Title: lo.ToPtr(gofakeit.Name()),
				},
			},
			client: suite.client.api,
			ctx:    testUser2.UserCtx,
		},
		{
			name:        "not authorized",
			request:     testclient.CreateTrustCenterInput{},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "duplicate trust center for same organization",
			request:     testclient.CreateTrustCenterInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "one trust center at a time",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateTrustCenter(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

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
			if tc.request.CreateTrustCenterSetting != nil && tc.request.CreateTrustCenterSetting.Title != nil {
				require.NotNil(t, setting)
				require.NotNil(t, setting.Title)
				assert.Equal(t, *tc.request.CreateTrustCenterSetting.Title, *setting.Title)
			} else {
				assert.Equal(t, fmt.Sprintf("%s Trust Center", org.Name), *setting.Title)
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
		client: suite.client,
	}).MustNew(testUser1.UserCtx, t)

	trustCenter2 := (&TrustCenterBuilder{
		client: suite.client,
	}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
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
		request       testclient.UpdateTrustCenterInput
		client        *testclient.TestClient
		ctx           context.Context
		expectedErr   string
	}{
		{
			name:          "happy path, update tags",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"updated", "test"},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:          "happy path, update custom domain",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				CustomDomainID: &customDomain.ID,
				AddPost: &testclient.CreateNoteInput{
					Text: "Adding a post about obtaining our SOC 2 compliance attestation.",
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:          "happy path, update settings",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				UpdateTrustCenterSetting: &testclient.UpdateTrustCenterSettingInput{
					Title:        lo.ToPtr("Updated Trust Center Title"),
					Overview:     lo.ToPtr("Updated Trust Center Overview"),
					PrimaryColor: lo.ToPtr("#FF5733"),
				},
				AddPost: &testclient.CreateNoteInput{
					Text: "Adding a post about obtaining our FedRamp Moderate compliance attestation.",
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:          "happy path, append tags",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				AppendTags: []string{"appended", "tag"},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:          "happy path, using admin user",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"admin", "update"},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:          "happy path, using personal access token",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"pat", "update"},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:          "not authorized, view only user",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"unauthorized"},
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:          "not authorized, different org user",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"unauthorized"},
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:          "trust center not found",
			trustCenterID: "non-existent-id",
			request: testclient.UpdateTrustCenterInput{
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

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.trustCenterID, resp.UpdateTrustCenter.TrustCenter.ID))

			// Check updated fields
			if tc.request.Tags != nil {
				assert.Check(t, is.DeepEqual(tc.request.Tags, resp.UpdateTrustCenter.TrustCenter.Tags))

				tagDefs, err := tc.client.GetTagDefinitions(tc.ctx, nil, nil, &testclient.TagDefinitionWhereInput{
					NameIn: tc.request.Tags,
				})

				assert.NilError(t, err)
				assert.Check(t, is.Len(tagDefs.TagDefinitions.Edges, len(tc.request.Tags)))
			}

			if tc.request.CustomDomainID != nil {
				assert.Check(t, is.Equal(*tc.request.CustomDomainID, *resp.UpdateTrustCenter.TrustCenter.CustomDomainID))
			}

			if tc.request.UpdateTrustCenterSetting != nil {
				assert.Check(t, is.Equal(*tc.request.UpdateTrustCenterSetting.Title, *resp.UpdateTrustCenter.TrustCenter.GetSetting().Title))
				assert.Check(t, is.Equal(*tc.request.UpdateTrustCenterSetting.Overview, *resp.UpdateTrustCenter.TrustCenter.GetSetting().Overview))
				assert.Check(t, is.Equal(*tc.request.UpdateTrustCenterSetting.PrimaryColor, *resp.UpdateTrustCenter.TrustCenter.GetSetting().PrimaryColor))
			}

			if tc.request.AddPost != nil {
				assert.Check(t, resp.UpdateTrustCenter.TrustCenter.Posts.Edges != nil)
				assert.Check(t, len(resp.UpdateTrustCenter.TrustCenter.Posts.Edges) > 0)
				found := false
				for _, edge := range resp.UpdateTrustCenter.TrustCenter.Posts.Edges {
					if edge.Node.Text == tc.request.AddPost.Text {
						found = true
						break
					}
				}
				assert.Check(t, found, "expected post text not found in trust center posts")
			}
		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: customDomain.MappableDomainID}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: customDomain.ID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationDeleteTrustCenter(t *testing.T) {
	t.Parallel()
	// Create new test users
	testUser := suite.userBuilder(context.Background(), t)
	testUserOther := suite.userBuilder(testUser.UserCtx, t)

	// create objects to be deleted
	trustCenter1 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUserOther.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:       "happy path, delete trust center",
			idToDelete: trustCenter1.ID,
			client:     suite.client.api,
			ctx:        testUser.UserCtx,
		},
		{
			name:        "not authorized, different org user",
			idToDelete:  trustCenter2.ID,
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "trust center not found",
			idToDelete:  "non-existent-id",
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteTrustCenter(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteTrustCenter.DeletedID))
		})
	}
}

func TestMutationDeleteTrustCenterWithCustomDomain(t *testing.T) {
	t.Parallel()

	// Create a new test user
	testUser := suite.userBuilder(context.Background(), t)

	// Create a custom domain
	customDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	// Create a trust center with the custom domain
	trustCenter := (&TrustCenterBuilder{client: suite.client, CustomDomainID: customDomain.ID}).MustNew(testUser.UserCtx, t)

	// Delete the trust center
	resp, err := suite.client.api.DeleteTrustCenter(testUser.UserCtx, trustCenter.ID)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Check(t, is.Equal(trustCenter.ID, resp.DeleteTrustCenter.DeletedID))

	// Verify the custom domain was also deleted
	dbCtx := setContext(testUser.UserCtx, suite.client.db)
	exists, err := suite.client.db.CustomDomain.Query().Where(customdomain.ID(customDomain.ID)).Exist(dbCtx)
	assert.NilError(t, err)
	assert.Check(t, !exists, "custom domain should have been deleted")

	// Clean up the mappable domain
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: customDomain.MappableDomainID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationDeleteTrustCenterWithPreviewDomain(t *testing.T) {
	t.Parallel()

	// Create a new test user
	testUser := suite.userBuilder(context.Background(), t)

	// Create a preview domain (custom domain)
	previewDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	// Create a trust center and manually set the preview domain ID
	// We need to use the database directly to set the preview domain
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	dbCtx := setContext(testUser.UserCtx, suite.client.db)
	trustCenter, err := suite.client.db.TrustCenter.UpdateOneID(trustCenter.ID).
		SetPreviewDomainID(previewDomain.ID).
		Save(dbCtx)
	assert.NilError(t, err)

	// Delete the trust center
	resp, err := suite.client.api.DeleteTrustCenter(testUser.UserCtx, trustCenter.ID)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Check(t, is.Equal(trustCenter.ID, resp.DeleteTrustCenter.DeletedID))

	// Verify a job was queued to delete the preview domain
	// Note: We can't easily verify the exact job args without accessing the river queue,
	// but we can verify the preview domain still exists (it will be deleted by the job worker)
	exists, err := suite.client.db.CustomDomain.Query().Where(customdomain.ID(previewDomain.ID)).Exist(dbCtx)
	assert.NilError(t, err)
	assert.Check(t, exists, "preview domain should still exist (will be deleted by job)")

	// Clean up the preview domain and mappable domain
	(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: previewDomain.ID}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: previewDomain.MappableDomainID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationDeleteTrustCenterWithBothDomains(t *testing.T) {
	t.Parallel()

	// Create a new test user
	testUser := suite.userBuilder(context.Background(), t)

	// Create a custom domain
	customDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	// Create a preview domain
	previewDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	// Create a trust center with the custom domain
	trustCenter := (&TrustCenterBuilder{client: suite.client, CustomDomainID: customDomain.ID}).MustNew(testUser.UserCtx, t)

	// Set the preview domain ID
	dbCtx := setContext(testUser.UserCtx, suite.client.db)
	trustCenter, err := suite.client.db.TrustCenter.UpdateOneID(trustCenter.ID).
		SetPreviewDomainID(previewDomain.ID).
		Save(dbCtx)
	assert.NilError(t, err)

	// Delete the trust center
	resp, err := suite.client.api.DeleteTrustCenter(testUser.UserCtx, trustCenter.ID)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Check(t, is.Equal(trustCenter.ID, resp.DeleteTrustCenter.DeletedID))

	// Verify the custom domain was deleted
	customDomainExists, err := suite.client.db.CustomDomain.Query().Where(customdomain.ID(customDomain.ID)).Exist(dbCtx)
	assert.NilError(t, err)
	assert.Check(t, !customDomainExists, "custom domain should have been deleted")

	// Verify the preview domain still exists (will be deleted by job)
	previewDomainExists, err := suite.client.db.CustomDomain.Query().Where(customdomain.ID(previewDomain.ID)).Exist(dbCtx)
	assert.NilError(t, err)
	assert.Check(t, previewDomainExists, "preview domain should still exist (will be deleted by job)")

	// Clean up the preview domain and mappable domains
	(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: previewDomain.ID}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: customDomain.MappableDomainID}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: previewDomain.MappableDomainID}).MustDelete(systemAdminUser.UserCtx, t)
}

// createAnonymousTrustCenterContext creates a context for an anonymous trust center user
func createAnonymousTrustCenterContext(trustCenterID, organizationID string) context.Context {
	anonUserID := fmt.Sprintf("%s%s", authmanager.AnonTrustcenterJWTPrefix, ulids.New().String())

	anonUser := &auth.AnonymousTrustCenterUser{
		SubjectID:          anonUserID,
		SubjectName:        "Anonymous User",
		OrganizationID:     organizationID,
		AuthenticationType: auth.JWTAuthentication,
		TrustCenterID:      trustCenterID,
	}

	ctx := context.Background()
	return auth.WithAnonymousTrustCenterUser(ctx, anonUser)
}

func TestQueryTrustCenterAsAnonymousUser(t *testing.T) {
	t.Parallel()

	// create new test users
	testUser := suite.userBuilder(context.Background(), t)
	testUserOther := suite.userBuilder(context.Background(), t)

	// Create a trust center for testing
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	// Create another trust center that the anonymous user should NOT have access to
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUserOther.UserCtx, t)

	testCases := []struct {
		name           string
		queryID        string
		trustCenterID  string
		organizationID string
		client         *testclient.TestClient
		expectedErr    string
		shouldSucceed  bool
	}{
		{
			name:           "happy path - anonymous user can query their trust center by ID",
			queryID:        trustCenter.ID,
			trustCenterID:  trustCenter.ID,
			organizationID: testUser.OrganizationID,
			client:         suite.client.api,
			shouldSucceed:  true,
		},
		{
			name:           "anonymous user cannot query different trust center",
			queryID:        trustCenter2.ID,
			trustCenterID:  trustCenter.ID, // Anonymous user has access to trustCenter, not trustCenter2
			organizationID: testUser.OrganizationID,
			client:         suite.client.api,
			expectedErr:    notFoundErrorMsg,
			shouldSucceed:  false,
		},
		{
			name:           "anonymous user cannot query non-existent trust center",
			queryID:        "non-existent-id",
			trustCenterID:  trustCenter.ID,
			organizationID: testUser.OrganizationID,
			client:         suite.client.api,
			expectedErr:    notFoundErrorMsg,
			shouldSucceed:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create anonymous trust center context
			anonCtx := createAnonymousTrustCenterContext(tc.trustCenterID, tc.organizationID)

			resp, err := tc.client.GetTrustCenterByID(anonCtx, tc.queryID)

			if !tc.shouldSucceed {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.queryID, resp.TrustCenter.ID))
			assert.Check(t, resp.TrustCenter.Slug != nil)
			assert.Check(t, resp.TrustCenter.OwnerID != nil)
			assert.Check(t, is.Equal(tc.organizationID, *resp.TrustCenter.OwnerID))

			setting := resp.TrustCenter.GetSetting()
			assert.Check(t, setting != nil)
			assert.Check(t, setting.Title != nil)
			assert.Check(t, setting.Overview != nil)
			assert.Check(t, setting.PrimaryColor != nil)
		})
	}

	// Clean up
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUserOther.UserCtx, t)
}

func TestQueryTrustCentersAsAnonymousUser(t *testing.T) {
	t.Parallel()

	// create new test users
	testUser := suite.userBuilder(context.Background(), t)
	testUserOther := suite.userBuilder(context.Background(), t)

	// Create a trust center for testing
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	// Create another trust center that the anonymous user should NOT have access to
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUserOther.UserCtx, t)

	testCases := []struct {
		name           string
		trustCenterID  string
		organizationID string
		client         *testclient.TestClient
		expectedCount  int
	}{
		{
			name:           "anonymous user can only see their trust center in list query",
			trustCenterID:  trustCenter.ID,
			organizationID: testUser.OrganizationID,
			client:         suite.client.api,
			expectedCount:  1, // Should only see the one trust center they have access to
		},
		{
			name:           "anonymous user with different trust center sees only their trust center",
			trustCenterID:  trustCenter2.ID,
			organizationID: testUserOther.OrganizationID,
			client:         suite.client.api,
			expectedCount:  1, // Should only see the one trust center they have access to
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create anonymous trust center context
			anonCtx := createAnonymousTrustCenterContext(tc.trustCenterID, tc.organizationID)

			resp, err := tc.client.GetAllTrustCenters(anonCtx)

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.expectedCount, len(resp.TrustCenters.Edges)))

			if len(resp.TrustCenters.Edges) > 0 {
				// Verify that the returned trust center is the one the anonymous user has access to
				returnedTrustCenter := resp.TrustCenters.Edges[0].Node
				assert.Check(t, is.Equal(tc.trustCenterID, returnedTrustCenter.ID))
				assert.Check(t, is.Equal(tc.organizationID, *returnedTrustCenter.OwnerID))
			}
		})
	}

	// Clean up
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUserOther.UserCtx, t)
}

func TestMutationUpdateTrustCenterSetting(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name        string
		settingID   string
		logoPath    string
		invalidFile bool
		updateInput testclient.UpdateTrustCenterSettingInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "happy path - update logo",
			settingID:   trustCenter.Edges.Setting.ID,
			logoPath:    "testdata/uploads/logo.png",
			updateInput: testclient.UpdateTrustCenterSettingInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
		},
		{
			name:      "happy path - update logo with other fields",
			settingID: trustCenter.Edges.Setting.ID,
			logoPath:  "testdata/uploads/logo.png",
			updateInput: testclient.UpdateTrustCenterSettingInput{
				Title:        lo.ToPtr("Updated Title with Logo"),
				PrimaryColor: lo.ToPtr("#FF5733"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},

		{
			name:        "invalid file type - text file instead of image",
			settingID:   trustCenter.Edges.Setting.ID,
			logoPath:    "testdata/uploads/hello.txt",
			invalidFile: true,
			updateInput: testclient.UpdateTrustCenterSettingInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "unsupported mime type uploaded: text/plain",
		},
		{
			name:        "not authorized - view only user",
			settingID:   trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "not authorized - different organization user",
			settingID:   trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "trust center setting not found",
			settingID:   "non-existent-setting-id",
			updateInput: testclient.UpdateTrustCenterSettingInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:      "update without logo file - should work",
			settingID: trustCenter.Edges.Setting.ID,
			logoPath:  "", // No logo file
			updateInput: testclient.UpdateTrustCenterSettingInput{
				Title:        lo.ToPtr("Updated Title Only"),
				Overview:     lo.ToPtr("Updated Overview"),
				PrimaryColor: lo.ToPtr("#00FF00"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "happy path - update theme mode to EASY",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				ThemeMode: lo.ToPtr(enums.TrustCenterThemeModeEasy),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "happy path - update theme mode to ADVANCED",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				ThemeMode: lo.ToPtr(enums.TrustCenterThemeModeAdvanced),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "happy path - update font",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				Font: lo.ToPtr("Arial, sans-serif"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "happy path - update foreground color",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				ForegroundColor: lo.ToPtr("#333333"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "happy path - update background color",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				BackgroundColor: lo.ToPtr("#FFFFFF"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "happy path - update accent color",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				AccentColor: lo.ToPtr("#007BFF"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "happy path - update all theme fields together",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				ThemeMode:       lo.ToPtr(enums.TrustCenterThemeModeAdvanced),
				PrimaryColor:    lo.ToPtr("#FF6B35"),
				Font:            lo.ToPtr("Roboto, sans-serif"),
				ForegroundColor: lo.ToPtr("#2C3E50"),
				BackgroundColor: lo.ToPtr("#F8F9FA"),
				AccentColor:     lo.ToPtr("#28A745"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			var logoFile *graphql.Upload

			// Create file upload if logoPath is provided
			if tc.logoPath != "" {
				uploadFile, err := storage.NewUploadFile(tc.logoPath)
				assert.NilError(t, err)

				logoFile = &graphql.Upload{
					File:        uploadFile.RawFile,
					Filename:    uploadFile.OriginalName,
					Size:        uploadFile.Size,
					ContentType: uploadFile.ContentType,
				}

				// Set up mock expectations based on whether we expect an error
				if tc.expectedErr == "" {
					expectUpload(t, suite.client.mockProvider, []graphql.Upload{*logoFile})
				} else {
					expectUploadCheckOnly(t, suite.client.mockProvider)
				}
			}

			resp, err := tc.client.UpdateTrustCenterSetting(tc.ctx, tc.settingID, tc.updateInput, logoFile, nil)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.settingID, resp.UpdateTrustCenterSetting.TrustCenterSetting.ID))

			// Check updated fields
			if tc.updateInput.Title != nil {
				assert.Check(t, is.Equal(*tc.updateInput.Title, *resp.UpdateTrustCenterSetting.TrustCenterSetting.Title))
			}

			if tc.updateInput.Overview != nil {
				assert.Check(t, is.Equal(*tc.updateInput.Overview, *resp.UpdateTrustCenterSetting.TrustCenterSetting.Overview))
			}

			if tc.updateInput.PrimaryColor != nil {
				assert.Check(t, is.Equal(*tc.updateInput.PrimaryColor, *resp.UpdateTrustCenterSetting.TrustCenterSetting.PrimaryColor))
			}

			if tc.updateInput.ThemeMode != nil {
				assert.Check(t, is.Equal(*tc.updateInput.ThemeMode, *resp.UpdateTrustCenterSetting.TrustCenterSetting.ThemeMode))
			}

			if tc.updateInput.Font != nil {
				assert.Check(t, is.Equal(*tc.updateInput.Font, *resp.UpdateTrustCenterSetting.TrustCenterSetting.Font))
			}

			if tc.updateInput.ForegroundColor != nil {
				assert.Check(t, is.Equal(*tc.updateInput.ForegroundColor, *resp.UpdateTrustCenterSetting.TrustCenterSetting.ForegroundColor))
			}

			if tc.updateInput.BackgroundColor != nil {
				assert.Check(t, is.Equal(*tc.updateInput.BackgroundColor, *resp.UpdateTrustCenterSetting.TrustCenterSetting.BackgroundColor))
			}

			if tc.updateInput.AccentColor != nil {
				assert.Check(t, is.Equal(*tc.updateInput.AccentColor, *resp.UpdateTrustCenterSetting.TrustCenterSetting.AccentColor))
			}
		})
	}

	// Clean up
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
}
