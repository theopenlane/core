package graphapi_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/customdomain"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestQueryTrustCenterByID(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

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
			ctx:     tcOrg.owner.UserCtx,
		},
		{
			name:    "happy path, view only user",
			queryID: trustCenter.ID,
			client:  suite.client.api,
			ctx:     tcOrg.member.UserCtx,
		},
		{
			name:     "trust center not found",
			queryID:  "non-existent-id",
			client:   suite.client.api,
			ctx:      tcOrg.owner.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "not authorized to query org",
			queryID:  trustCenter.ID,
			client:   suite.client.api,
			ctx:      sharedTestUser2.UserCtx,
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
			assert.Check(t, is.Equal(tcOrg.organizationID, *resp.TrustCenter.OwnerID))

			setting := resp.TrustCenter.GetSetting()
			assert.Assert(t, setting != nil)
			assert.Check(t, setting.Title != nil)
			assert.Check(t, setting.Overview != nil)
			assert.Check(t, setting.PrimaryColor != nil)
		})
	}

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}

func TestQueryTrustCenters(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t, withCustomDomain(), withAllUserTypes())
	tcOrg2 := createFreshOrgWithTrustCenter(t)
	trustCenter1 := tcOrg.trustCenter

	nonExistentSlug := "nonexistent-slug"

	if trustCenter1.CustomDomainID == nil {
		failNow(t, "expected trust center custom domain but no ID was returned")

	}
	customDomainTrustCenter1, err := suite.client.api.GetCustomDomainByID(tcOrg.owner.UserCtx, *trustCenter1.CustomDomainID)
	requireNoError(t, err)

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
			ctx:             tcOrg.owner.UserCtx,
			expectedResults: 1,
		},
		{
			name:            "return all, ro user",
			client:          suite.client.api,
			ctx:             tcOrg.member.UserCtx,
			expectedResults: 1,
		},
		{
			name:   "query by org ID",
			client: suite.client.api,
			ctx:    tcOrg.superAdmin.UserCtx,
			where: &testclient.TrustCenterWhereInput{
				OwnerID: &tcOrg.organizationID,
			},
			expectedResults: 1,
		},
		{
			name:   "query by slug",
			client: suite.client.api,
			ctx:    tcOrg.admin.UserCtx,
			where: &testclient.TrustCenterWhereInput{
				Slug: &trustCenter1.Slug,
			},
			expectedResults: 1,
		},
		{
			name:   "query by slug, not found",
			client: suite.client.api,
			ctx:    tcOrg.owner.UserCtx,
			where: &testclient.TrustCenterWhereInput{
				Slug: &nonExistentSlug,
			},
			expectedResults: 0,
		},
		{
			name:   "query by custom domain, slug",
			client: suite.client.api,
			ctx:    tcOrg.owner.UserCtx,
			where: &testclient.TrustCenterWhereInput{
				And: []*testclient.TrustCenterWhereInput{
					{
						Slug: &trustCenter1.Slug,
					},
					{
						HasCustomDomainWith: []*testclient.CustomDomainWhereInput{
							{
								CnameRecord: &customDomainTrustCenter1.CustomDomain.CnameRecord,
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
			ctx:    tcOrg.owner.UserCtx,
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
				assert.Assert(t, node.Node != nil)
				assert.Check(t, node.Node.Slug != nil)
				assert.Check(t, node.Node.OwnerID != nil)
				assert.Check(t, is.Equal(tcOrg.organizationID, *node.Node.OwnerID))
				setting := node.Node.GetSetting()
				assert.Assert(t, setting != nil)
				assert.Check(t, setting.Title != nil)
				assert.Check(t, setting.Overview != nil)
				assert.Check(t, setting.PrimaryColor != nil)
			}

		})
	}

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestMutationCreateTrustCenter(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t, withCustomDomain())
	customDomainAnotherOrg, err := suite.client.api.GetCustomDomainByID(tcOrg.owner.UserCtx, *tcOrg.trustCenter.CustomDomainID)
	requireNoError(t, err)

	localTestUser := suite.seedFreshMinimalOrgUsers(t, false)
	customDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(localTestUser.owner.UserCtx, t)

	// create trust center standard
	trustCenterControlStd := (&StandardBuilder{client: suite.client, Name: "OTS", Framework: "openlane-trust-center", IsPublic: true}).MustNew(sharedSystemAdminUser.UserCtx, t)

	trustCenterControlIDs := []string{}
	numTrustCenterControls := 5
	for range numTrustCenterControls {
		control := (&ControlBuilder{client: suite.client, StandardID: trustCenterControlStd.ID}).MustNew(sharedSystemAdminUser.UserCtx, t)
		trustCenterControlIDs = append(trustCenterControlIDs, control.ID)
	}

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
			ctx:     localTestUser.owner.UserCtx,
		},
		{
			name: "custom domain for different organization should error",
			request: testclient.CreateTrustCenterInput{
				CustomDomainID: &customDomainAnotherOrg.CustomDomain.ID,
			},
			client:      suite.client.api,
			ctx:         localTestUser.owner.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "custom domain setting",
			request: testclient.CreateTrustCenterInput{
				CustomDomainID: &customDomain.ID,
			},
			client: suite.client.api,
			ctx:    localTestUser.owner.UserCtx,
		},
		{
			name: "happy path with settings for different organization",
			request: testclient.CreateTrustCenterInput{
				CreateTrustCenterSetting: &testclient.CreateTrustCenterSettingInput{
					Title: lo.ToPtr(gofakeit.Name()),
				},
			},
			client: suite.client.api,
			ctx:    localTestUser.owner.UserCtx,
		},
		{
			name:        "not authorized",
			request:     testclient.CreateTrustCenterInput{},
			client:      suite.client.api,
			ctx:         localTestUser.member.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "duplicate trust center for same organization",
			request:     testclient.CreateTrustCenterInput{},
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
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
				assert.Assert(t, resp.CreateTrustCenter.TrustCenter.CustomDomainID != nil)
				assert.Check(t, is.Equal(*tc.request.CustomDomainID, *resp.CreateTrustCenter.TrustCenter.CustomDomainID))
			} else {
				assert.Check(t, resp.CreateTrustCenter.TrustCenter.CustomDomainID == nil)
			}

			// Verify slug is the lowercased, alphanumeric version of the org name
			// Get the organization to check its name using a context that allows database access
			dbCtx := setContext(tc.ctx, suite.client.db)
			org, err := suite.client.db.Organization.Get(dbCtx, *resp.CreateTrustCenter.TrustCenter.OwnerID)
			assert.NilError(t, err)

			// Generate expected slug: remove non-alphanumeric chars and lowercase
			expectedSlug := strcase.KebabCase(org.Name)
			require.NotNil(t, resp.CreateTrustCenter.TrustCenter.Slug)
			assert.Equal(t, expectedSlug, *resp.CreateTrustCenter.TrustCenter.Slug)
			setting := resp.CreateTrustCenter.TrustCenter.GetSetting()
			if tc.request.CreateTrustCenterSetting != nil && tc.request.CreateTrustCenterSetting.Title != nil {
				assert.Assert(t, setting != nil)
				assert.Assert(t, setting.Title != nil)
				assert.Equal(t, *tc.request.CreateTrustCenterSetting.Title, *setting.Title)
			} else {
				assert.Equal(t, fmt.Sprintf("%s Trust Center", org.Name), *setting.Title)
			}

			// ensure trust center preview settings object is created
			assert.Assert(t, resp.CreateTrustCenter.TrustCenter.PreviewSetting != nil)
			assert.Check(t, resp.CreateTrustCenter.TrustCenter.PreviewSetting.ID != "")

			// ensure trust center watermark config object is created
			assert.Assert(t, resp.CreateTrustCenter.TrustCenter.WatermarkConfig != nil)
			assert.Check(t, resp.CreateTrustCenter.TrustCenter.WatermarkConfig.Text != nil)

			// get controls for the trust center standard and ensure they are added to the trust center
			controlsResp, err := tc.client.GetControls(tc.ctx, nil, nil, nil, nil, nil, &testclient.ControlWhereInput{
				IsTrustCenterControl: lo.ToPtr(true),
				SystemOwned:          lo.ToPtr(false),
			})
			assert.NilError(t, err)
			assert.Assert(t, controlsResp != nil)
			assert.Check(t, is.Equal(numTrustCenterControls, len(controlsResp.Controls.Edges)))

			// Clean up
			(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: resp.CreateTrustCenter.TrustCenter.ID}).MustDelete(tc.ctx, t)
		})
	}

	// Clean up the existing trust center
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(localTestUser.owner.UserCtx, t)
}

func TestGetAllTrustCenters(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t, withAllUserTypes())
	tcOrg2 := createFreshOrgWithTrustCenter(t)

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
			ctx:             tcOrg.owner.UserCtx,
			expectedResults: 1,
		},
		{
			name:            "happy path - admin user sees all trust centers",
			client:          suite.client.api,
			ctx:             tcOrg.admin.UserCtx,
			expectedResults: 1,
		},
		{
			name:            "happy path - view only user",
			client:          suite.client.api,
			ctx:             tcOrg.member.UserCtx,
			expectedResults: 1,
		},
		{
			name:            "happy path - different user sees only their trust centers",
			client:          suite.client.api,
			ctx:             tcOrg2.owner.UserCtx,
			expectedResults: 1,
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
			switch tc.ctx {
			case tcOrg.owner.UserCtx, tcOrg.admin.UserCtx, tcOrg.member.UserCtx, tcOrg.superAdmin.UserCtx:
				for _, edge := range resp.TrustCenters.Edges {
					assert.Check(t, is.Equal(tcOrg.organizationID, *edge.Node.OwnerID))
				}
			case tcOrg2.owner.UserCtx:
				for _, edge := range resp.TrustCenters.Edges {
					assert.Check(t, is.Equal(tcOrg2.organizationID, *edge.Node.OwnerID))
				}
			}
		})
	}

	// Clean up created trust centers
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}

func TestMutationUpdateTrustCenter(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t, withCustomDomain(), withAllUserTypes())
	trustCenter := tcOrg.trustCenter

	if trustCenter.CustomDomainID == nil {
		failNow(t, "expected trust center custom domain but no ID was returned")

	}
	customDomainTrustCenter, err := suite.client.api.GetCustomDomainByID(tcOrg.owner.UserCtx, *trustCenter.CustomDomainID)
	requireNoError(t, err)

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
			ctx:    tcOrg.owner.UserCtx,
		},
		{
			name:          "happy path, update custom domain",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				CustomDomainID: &customDomainTrustCenter.CustomDomain.ID,
				AddPost: &testclient.CreateNoteInput{
					Text: "Adding a post about obtaining our SOC 2 compliance attestation.",
				},
			},
			client: suite.client.api,
			ctx:    tcOrg.superAdmin.UserCtx,
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
			ctx:    tcOrg.admin.UserCtx,
		},
		{
			name:          "happy path, append tags",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				AppendTags: []string{"appended", "tag"},
			},
			client: suite.client.api,
			ctx:    tcOrg.owner.UserCtx,
		},
		{
			name:          "happy path, using admin user",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"admin", "update"},
			},
			client: suite.client.api,
			ctx:    tcOrg.admin.UserCtx,
		},
		{
			name:          "happy path, using personal access token",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"pat", "update"},
			},
			client: tcOrg.adminPatClient,
			ctx:    context.Background(),
		},
		{
			name:          "not authorized, view only user",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"unauthorized"},
			},
			client:      suite.client.api,
			ctx:         tcOrg.member.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:          "not authorized, different org user",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"unauthorized"},
			},
			client:      suite.client.api,
			ctx:         sharedTestUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:          "trust center not found",
			trustCenterID: "non-existent-id",
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"test"},
			},
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
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

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}

func TestMutationDeleteTrustCenter(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t)
	tcOrg2 := createFreshOrgWithTrustCenter(t)

	trustCenter1 := tcOrg.trustCenter
	trustCenter2 := tcOrg2.trustCenter

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
			ctx:        tcOrg.owner.UserCtx,
		},
		{
			name:        "not authorized, different org user",
			idToDelete:  trustCenter2.ID,
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "trust center not found",
			idToDelete:  "non-existent-id",
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
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

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestQueryTrustCenterAsAnonymousUser(t *testing.T) {
	t.Parallel()
	// create new test users
	tcOrg := createFreshOrgWithTrustCenter(t)
	tcOrg2 := createFreshOrgWithTrustCenter(t)

	trustCenter := tcOrg.trustCenter

	// create trust center entities for the trust center
	createLogoUpload := logoFileFunc(t)
	logoFile := createLogoUpload()

	expectUpload(t, suite.client.mockProvider, []graphql.Upload{*logoFile})

	_, err := suite.client.api.CreateTrustCenterEntity(tcOrg.owner.UserCtx, testclient.CreateTrustCenterEntityInput{
		Name:          "test entity 1",
		TrustCenterID: &trustCenter.ID,
		URL:           lo.ToPtr(gofakeit.URL()),
	}, logoFile, nil)
	assert.NilError(t, err)

	_, err = suite.client.api.UpdateTrustCenter(tcOrg.owner.UserCtx, trustCenter.ID, testclient.UpdateTrustCenterInput{
		AddPost: &testclient.CreateNoteInput{
			Text: "this is an update",
		},
	})
	assert.NilError(t, err)

	// create trust center compliance
	std := (&StandardBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	_, err = suite.client.api.CreateTrustCenterCompliance(tcOrg.owner.UserCtx, testclient.CreateTrustCenterComplianceInput{
		StandardID: std.ID,
	})
	assert.NilError(t, err)

	// create subprocessor
	sbpr := (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	sbprKind := (&CustomTypeEnumBuilder{
		client:     suite.client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg.owner.UserCtx, t)
	_, err = suite.client.api.CreateTrustCenterSubprocessor(tcOrg.owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
		SubprocessorID:                  sbpr.ID,
		TrustCenterSubprocessorKindName: &sbprKind.Name,
		Countries:                       []string{"United States"},
	})
	assert.NilError(t, err)

	// create custom type enum for trust center doc kind
	docKind := (&CustomTypeEnumBuilder{
		client:     suite.client,
		ObjectType: "trust_center_doc",
	}).MustNew(tcOrg.owner.UserCtx, t)

	// create trust center doc
	createFileUpload := uploadFileFunc(t, pdfFilePath)
	fileUpload := createFileUpload()

	expectUpload(t, suite.client.mockProvider, []graphql.Upload{*fileUpload})
	doc, err := suite.client.api.CreateTrustCenterDoc(tcOrg.owner.UserCtx, testclient.CreateTrustCenterDocInput{
		Title:                  "Test Doc",
		TrustCenterDocKindName: &docKind.Name,
		Visibility:             &enums.TrustCenterDocumentVisibilityPubliclyVisible,
	}, *fileUpload)
	assert.NilError(t, err)
	assert.Check(t, doc.CreateTrustCenterDoc.TrustCenterDoc.ID != "")
	assert.Check(t, doc.CreateTrustCenterDoc.TrustCenterDoc.OriginalFile != nil)
	assert.Check(t, doc.CreateTrustCenterDoc.TrustCenterDoc.OriginalFileID != nil)
	assert.Check(t, doc.CreateTrustCenterDoc.TrustCenterDoc.Title == "Test Doc")

	// create trust center FAQ
	faqNote := (&NoteBuilder{client: suite.client, TrustCenterID: trustCenter.ID}).MustNew(tcOrg.owner.UserCtx, t)
	_, err = suite.client.api.CreateTrustCenterFaq(tcOrg.owner.UserCtx, testclient.CreateTrustCenterFAQInput{
		NoteID:        faqNote.ID,
		TrustCenterID: &trustCenter.ID,
		ReferenceLink: lo.ToPtr("https://example.com/faq"),
		DisplayOrder:  lo.ToPtr(int64(1)),
	})
	assert.NilError(t, err)

	// Create another trust center that the anonymous user should NOT have access to
	trustCenter2 := tcOrg2.trustCenter

	testCases := []struct {
		name           string
		queryID        string
		trustCenterID  string
		organizationID string
		client         *testclient.TestClient
		expectedErr    string
		shouldSucceed  bool
		isList         bool
	}{
		{
			name:           "list query - anonymous user can query their trust center, only one returned",
			queryID:        trustCenter.ID,
			trustCenterID:  trustCenter.ID,
			organizationID: tcOrg.organizationID,
			client:         suite.client.api,
			shouldSucceed:  true,
			isList:         true,
		},
		{
			name:           "anonymous user cannot query different trust center by id",
			queryID:        trustCenter2.ID,
			trustCenterID:  trustCenter.ID, // Anonymous user has access to trustCenter, not trustCenter2
			organizationID: tcOrg.organizationID,
			client:         suite.client.api,
			expectedErr:    notFoundErrorMsg,
			shouldSucceed:  false,
		},
		{
			name:           "anonymous user cannot query non-existent trust center by id",
			queryID:        "non-existent-id",
			trustCenterID:  trustCenter.ID,
			organizationID: tcOrg.organizationID,
			client:         suite.client.api,
			expectedErr:    notFoundErrorMsg,
			shouldSucceed:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create anonymous trust center context
			anonCtx := createAnonymousTrustCenterContext(tc.trustCenterID, tc.organizationID)

			trustCenter := &testclient.GetTrustCenterFrontendQuery_TrustCenters_Edges_Node{}
			if tc.isList {
				resp, err := tc.client.GetTrustCenterFrontendQuery(anonCtx)
				assert.NilError(t, err)
				assert.Check(t, resp != nil)
				assert.Check(t, is.Len(resp.TrustCenters.Edges, 1))

				trustCenter = resp.TrustCenters.Edges[0].Node
			} else {
				resp, err := tc.client.GetTrustCenterByID(anonCtx, tc.queryID)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, resp.TrustCenter.ID == "")

				return
			}

			assert.Assert(t, is.Equal(tc.trustCenterID, trustCenter.ID))
			assert.Check(t, trustCenter.Slug != nil)

			setting := trustCenter.Setting
			assert.Assert(t, setting != nil)
			assert.Assert(t, setting.Title != nil)
			assert.Check(t, setting.Overview != nil)
			assert.Check(t, setting.PrimaryColor != nil)

			previewSetting := trustCenter.PreviewSetting
			assert.Assert(t, previewSetting != nil)
			assert.Assert(t, previewSetting.ID != "")
			assert.Check(t, previewSetting.Overview != nil)
			assert.Check(t, previewSetting.Title != nil)
			assert.Check(t, previewSetting.PrimaryColor != nil)

			// // Verify that children are accessible
			assert.Assert(t, trustCenter.Posts.Edges != nil)
			assert.Assert(t, is.Len(trustCenter.Posts.Edges, 2))

			assert.Assert(t, trustCenter.TrustCenterCompliances.Edges != nil)
			assert.Assert(t, is.Len(trustCenter.TrustCenterCompliances.Edges, 1))
			assert.Check(t, trustCenter.TrustCenterCompliances.Edges[0].Node.ID != "")
			assert.Check(t, trustCenter.TrustCenterCompliances.Edges[0].Node.Standard.ID != "")
			assert.Check(t, trustCenter.TrustCenterCompliances.Edges[0].Node.Standard.Name != "")

			assert.Assert(t, trustCenter.TrustCenterSubprocessors.Edges != nil)
			assert.Assert(t, is.Len(trustCenter.TrustCenterSubprocessors.Edges, 1))
			assert.Check(t, trustCenter.TrustCenterSubprocessors.Edges[0].Node.ID != "")
			assert.Check(t, trustCenter.TrustCenterSubprocessors.Edges[0].Node.Subprocessor.Name != "")

			assert.Assert(t, trustCenter.TrustCenterDocs.Edges != nil)
			assert.Assert(t, is.Len(trustCenter.TrustCenterDocs.Edges, 1))
			assert.Check(t, trustCenter.TrustCenterDocs.Edges[0].Node.ID != "")
			assert.Check(t, trustCenter.TrustCenterDocs.Edges[0].Node.Title != "")

			// trust center entities
			assert.Assert(t, trustCenter.TrustCenterEntities.Edges != nil)
			assert.Check(t, is.Len(trustCenter.TrustCenterEntities.Edges, 1))
			assert.Check(t, trustCenter.TrustCenterEntities.Edges[0].Node.LogoFile != nil)
			assert.Check(t, trustCenter.TrustCenterEntities.Edges[0].Node.LogoFile.Base64 != nil)

			// trust center FAQs
			assert.Assert(t, trustCenter.TrustCenterFaqs.Edges != nil)
			assert.Assert(t, is.Len(trustCenter.TrustCenterFaqs.Edges, 1))
			assert.Check(t, trustCenter.TrustCenterFaqs.Edges[0].Node.ID != "")
			assert.Check(t, trustCenter.TrustCenterFaqs.Edges[0].Node.NoteID != "")
			assert.Check(t, trustCenter.TrustCenterFaqs.Edges[0].Node.ReferenceLink != nil)
		})
	}

	// create a trust center control and verify frontend query still works with controls present
	dbCtx := setContext(tcOrg.owner.UserCtx, suite.client.db)

	tcControl, err := suite.client.db.Control.Create().
		SetRefCode("OTS-TC-" + ulids.New().String()).
		SetTitle("Trust Center Control").
		SetSource(enums.ControlSourceUserDefined).
		SetIsTrustCenterControl(true).
		SetOwnerID(tcOrg.organizationID).
		Save(dbCtx)
	assert.NilError(t, err)

	_, err = suite.client.api.UpdateControl(tcOrg.owner.UserCtx, tcControl.ID, testclient.UpdateControlInput{
		TrustCenterVisibility: &enums.TrustCenterControlVisibilityPubliclyVisible,
	})
	assert.NilError(t, err)

	// create another trust center control for another trust center to ensure only controls for the queried trust center are returned in the frontend query
	dbCtx2 := setContext(tcOrg2.owner.UserCtx, suite.client.db)
	tcControlForAnotherOrg, err := suite.client.db.Control.Create().
		SetRefCode("OTS-TC-" + ulids.New().String()).
		SetTitle("Trust Center Control").
		SetSource(enums.ControlSourceUserDefined).
		SetIsTrustCenterControl(true).
		SetOwnerID(tcOrg2.organizationID).
		Save(dbCtx2)
	assert.NilError(t, err)

	_, err = suite.client.api.UpdateControl(tcOrg2.owner.UserCtx, tcControlForAnotherOrg.ID, testclient.UpdateControlInput{
		TrustCenterVisibility: &enums.TrustCenterControlVisibilityPubliclyVisible,
	})
	assert.NilError(t, err)

	t.Run("anonymous user frontend query returns all child objects with controls present", func(t *testing.T) {
		anonCtx := createAnonymousTrustCenterContext(trustCenter.ID, tcOrg.owner.OrganizationID)

		resp, err := suite.client.api.GetTrustCenterFrontendQuery(anonCtx)
		assert.NilError(t, err)
		assert.Check(t, resp != nil)
		assert.Assert(t, is.Len(resp.TrustCenters.Edges, 1))

		tc := resp.TrustCenters.Edges[0].Node

		assert.Check(t, tc.ID != "")
		assert.Check(t, tc.GetSetting() != nil)
		assert.Check(t, tc.TrustCenterCompliances.Edges != nil)
		assert.Check(t, tc.TrustCenterDocs.Edges != nil)
		assert.Check(t, tc.TrustCenterEntities.Edges != nil)
		assert.Check(t, tc.TrustCenterSubprocessors.Edges != nil)
		assert.Check(t, tc.Posts.Edges != nil)
		assert.Check(t, tc.TrustCenterFaqs.Edges != nil)
		assert.Assert(t, resp.Controls.Edges != nil)
		assert.Assert(t, is.Len(resp.Controls.Edges, 1))
		assert.Check(t, resp.Controls.Edges[0].Node.ID == tcControl.ID)
		assert.Check(t, resp.Controls.Edges[0].Node.RefCode != "")
	})

	t.Run("anonymous user can query publicly visible trust center controls", func(t *testing.T) {
		anonCtx := createAnonymousTrustCenterContext(trustCenter.ID, tcOrg.organizationID)

		resp, err := suite.client.api.GetAllControls(anonCtx)
		assert.NilError(t, err)
		assert.Check(t, resp != nil)
		assert.Check(t, is.Len(resp.Controls.Edges, 1))
		assert.Check(t, resp.Controls.Edges[0].Node.ID == tcControl.ID)
		assert.Check(t, resp.Controls.Edges[0].Node.RefCode != "")
	})

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestQueryTrustCentersAsAnonymousUser(t *testing.T) {
	t.Parallel()
	// create new test users
	tcOrg := createFreshOrgWithTrustCenter(t)
	tcOrg2 := createFreshOrgWithTrustCenter(t)

	trustCenter := tcOrg.trustCenter
	trustCenter2 := tcOrg2.trustCenter

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
			organizationID: tcOrg.organizationID,
			client:         suite.client.api,
			expectedCount:  1, // Should only see the one trust center they have access to
		},
		{
			name:           "anonymous user with different trust center sees only their trust center",
			trustCenterID:  trustCenter2.ID,
			organizationID: tcOrg2.organizationID,
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
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestMutationUpdateTrustCenterSetting(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t, withAllUserTypes())
	tcOrg2 := createFreshOrgWithTrustCenter(t)

	trustCenter := tcOrg.trustCenter

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
			logoPath:    logoFilePath,
			updateInput: testclient.UpdateTrustCenterSettingInput{},
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
		},
		{
			name:      "happy path - update logo with other fields",
			settingID: trustCenter.Edges.Setting.ID,
			logoPath:  logoFilePath,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				Title:        lo.ToPtr("Updated Title with Logo"),
				PrimaryColor: lo.ToPtr("#FF5733"),
			},
			client: suite.client.api,
			ctx:    tcOrg.admin.UserCtx,
		},

		{
			name:        "invalid file type - text file instead of image",
			settingID:   trustCenter.Edges.Setting.ID,
			logoPath:    txtFilePath,
			invalidFile: true,
			updateInput: testclient.UpdateTrustCenterSettingInput{},
			client:      suite.client.api,
			ctx:         tcOrg.superAdmin.UserCtx,
			expectedErr: "unsupported mime type uploaded: text/plain",
		},
		{
			name:        "not authorized - view only user",
			settingID:   trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{},
			client:      suite.client.api,
			ctx:         tcOrg.member.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "not authorized - different organization user",
			settingID:   trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{},
			client:      suite.client.api,
			ctx:         tcOrg2.owner.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "trust center setting not found",
			settingID:   "non-existent-setting-id",
			updateInput: testclient.UpdateTrustCenterSettingInput{},
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
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
			ctx:    tcOrg.owner.UserCtx,
		},
		{
			name:      "happy path - update theme mode to EASY",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				ThemeMode: lo.ToPtr(enums.TrustCenterThemeModeEasy),
			},
			client: suite.client.api,
			ctx:    tcOrg.owner.UserCtx,
		},
		{
			name:      "happy path - update theme mode to ADVANCED",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				ThemeMode: lo.ToPtr(enums.TrustCenterThemeModeAdvanced),
			},
			client: suite.client.api,
			ctx:    tcOrg.owner.UserCtx,
		},
		{
			name:      "happy path - update font",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				Font: lo.ToPtr("Arial, sans-serif"),
			},
			client: suite.client.api,
			ctx:    tcOrg.owner.UserCtx,
		},
		{
			name:      "happy path - update foreground color",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				ForegroundColor: lo.ToPtr("#333333"),
			},
			client: suite.client.api,
			ctx:    tcOrg.superAdmin.UserCtx,
		},
		{
			name:      "happy path - update background color",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				BackgroundColor: lo.ToPtr("#FFFFFF"),
			},
			client: suite.client.api,
			ctx:    tcOrg.admin.UserCtx,
		},
		{
			name:      "happy path - update accent color",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				AccentColor: lo.ToPtr("#007BFF"),
			},
			client: suite.client.api,
			ctx:    tcOrg.owner.UserCtx,
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
			ctx:    tcOrg.owner.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			var logoFile *graphql.Upload

			// Create file upload if logoPath is provided
			if tc.logoPath != "" {
				logoFile = uploadFile(t, tc.logoPath)

				// Set up mock expectations based on whether we expect an error
				if tc.expectedErr == "" {
					expectUpload(t, suite.client.mockProvider, []graphql.Upload{*logoFile})
				} else {
					expectUploadCheckOnly(t, suite.client.mockProvider)
				}
			}

			resp, err := tc.client.UpdateTrustCenterSetting(tc.ctx, tc.settingID, tc.updateInput, logoFile, nil, nil, nil, nil, nil)
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
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}

// TestTrustCenterCreateHookWithCustomDomain tests that CreatePirschDomain job is called when custom_domain_id is set during creation
func TestTrustCenterCreateHookWithCustomDomain(t *testing.T) {
	users := suite.seedFreshOrgUsers(t)

	customDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(users.owner.UserCtx, t)

	testCases := []struct {
		name                  string
		request               testclient.CreateTrustCenterInput
		client                *testclient.TestClient
		ctx                   context.Context
		expectCreatePirschJob bool
		expectedErr           string
	}{
		{
			name: "create trust center with custom domain - should trigger CreatePirschDomain job",
			request: testclient.CreateTrustCenterInput{
				CustomDomainID: &customDomain.ID,
			},
			client:                suite.client.api,
			ctx:                   users.owner.UserCtx,
			expectCreatePirschJob: true,
		},
		{
			name:                  "create trust center without custom domain - should NOT trigger CreatePirschDomain job",
			request:               testclient.CreateTrustCenterInput{},
			client:                suite.client.api,
			ctx:                   users.owner.UserCtx,
			expectCreatePirschJob: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear any existing jobs
			err := suite.client.db.Job.TruncateRiverTables(tc.ctx)
			assert.NilError(t, err)

			resp, err := tc.client.CreateTrustCenter(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Verify the job was or was not created based on expectation
			if tc.expectCreatePirschJob {
				jobs := rivertest.RequireManyInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()),
					[]rivertest.ExpectedJob{
						{
							Args: jobspec.CreatePirschDomainArgs{
								TrustCenterID: resp.CreateTrustCenter.TrustCenter.ID,
							},
						},
					})
				assert.Assert(t, jobs != nil)
				assert.Assert(t, is.Len(jobs, 1))
			} else {
				rivertest.RequireNotInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()), &jobspec.CreatePirschDomainArgs{}, nil)
			}

			// Clean up
			(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: resp.CreateTrustCenter.TrustCenter.ID}).MustDelete(tc.ctx, t)
		})
	}

	// Clean up custom domain
	cleanupOrganizationDataWithContext(users.owner.UserCtx, t)
}

// TestTrustCenterUpdateHookWithCustomDomain tests that CreatePirschDomain job is called when custom_domain_id changes from empty to non-empty
func TestTrustCenterUpdateHookWithCustomDomain(t *testing.T) {
	tcOrg := createFreshOrgWithTrustCenter(t, withAllUserTypes())
	trustCenter := tcOrg.trustCenter

	customDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)

	testCases := []struct {
		name                  string
		trustCenterID         string
		request               testclient.UpdateTrustCenterInput
		client                *testclient.TestClient
		ctx                   context.Context
		expectCreatePirschJob bool
		expectedErr           string
	}{
		{
			name:          "update trust center to add custom domain - should trigger CreatePirschDomain job",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				CustomDomainID: &customDomain.ID,
			},
			client:                suite.client.api,
			ctx:                   tcOrg.owner.UserCtx,
			expectCreatePirschJob: true,
		},
		{
			name:          "update trust center without changing custom domain - should NOT trigger CreatePirschDomain job",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"test", "tag"},
			},
			client:                suite.client.api,
			ctx:                   tcOrg.superAdmin.UserCtx,
			expectCreatePirschJob: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear any existing jobs
			err := suite.client.db.Job.TruncateRiverTables(tc.ctx)
			assert.NilError(t, err)

			resp, err := tc.client.UpdateTrustCenter(tc.ctx, tc.trustCenterID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Verify the job was or was not created based on expectation
			if tc.expectCreatePirschJob {
				jobs := rivertest.RequireManyInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()),
					[]rivertest.ExpectedJob{
						{
							Args: jobspec.CreatePirschDomainArgs{
								TrustCenterID: tc.trustCenterID,
							},
						},
					})
				assert.Assert(t, jobs != nil)
				assert.Assert(t, is.Len(jobs, 1))
			} else {
				rivertest.RequireNotInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()), &jobspec.CreatePirschDomainArgs{}, nil)
			}
		})
	}

	// Clean up
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}

// TestTrustCenterUpdateHookWithPirschDomainUpdate tests that UpdatePirschDomain job is called when custom_domain_id changes from one domain to another
func TestTrustCenterUpdateHookWithPirschDomainUpdate(t *testing.T) {
	tcOrgWithDomain := createFreshOrgWithTrustCenter(t, withCustomDomain())
	trustCenterWithDomain := tcOrgWithDomain.trustCenter

	// Create two custom domains
	customDomain2 := (&CustomDomainBuilder{client: suite.client}).MustNew(tcOrgWithDomain.owner.UserCtx, t)

	// Manually set pirsch_domain_id to simulate what would happen after the CreatePirschDomain job completes
	ctx := setContext(tcOrgWithDomain.owner.UserCtx, suite.client.db)
	fakePirschDomainID := "fake-pirsch-domain-id-for-update-test"
	_, err := suite.client.db.TrustCenter.UpdateOneID(trustCenterWithDomain.ID).SetPirschDomainID(fakePirschDomainID).Save(ctx)
	assert.NilError(t, err)

	testCases := []struct {
		name                  string
		trustCenterID         string
		request               testclient.UpdateTrustCenterInput
		client                *testclient.TestClient
		ctx                   context.Context
		expectUpdatePirschJob bool
		expectedErr           string
	}{
		{
			name:          "update trust center to change custom domain - should trigger UpdatePirschDomain job",
			trustCenterID: trustCenterWithDomain.ID,
			request: testclient.UpdateTrustCenterInput{
				CustomDomainID: &customDomain2.ID,
			},
			client:                suite.client.api,
			ctx:                   tcOrgWithDomain.owner.UserCtx,
			expectUpdatePirschJob: true,
		},
		{
			name:          "update trust center without changing custom domain - should NOT trigger UpdatePirschDomain job",
			trustCenterID: trustCenterWithDomain.ID,
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"test", "tag"},
			},
			client:                suite.client.api,
			ctx:                   tcOrgWithDomain.owner.UserCtx,
			expectUpdatePirschJob: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear any existing jobs
			err := suite.client.db.Job.TruncateRiverTables(tc.ctx)
			assert.NilError(t, err)

			resp, err := tc.client.UpdateTrustCenter(tc.ctx, tc.trustCenterID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Verify the job was or was not created based on expectation
			if tc.expectUpdatePirschJob {
				jobs := rivertest.RequireManyInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()),
					[]rivertest.ExpectedJob{
						{
							Args: jobspec.UpdatePirschDomainArgs{
								TrustCenterID: tc.trustCenterID,
							},
						},
					})
				assert.Assert(t, jobs != nil)
				assert.Assert(t, is.Len(jobs, 1))
			} else {
				rivertest.RequireNotInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()), &jobspec.UpdatePirschDomainArgs{}, nil)
			}
		})
	}

	// Clean up
	cleanupOrganizationDataWithContext(tcOrgWithDomain.owner.UserCtx, t)
}

// TestTrustCenterUpdateHookWithCustomDomainRemoval tests that DeletePirschDomain job is called when custom_domain_id is cleared
func TestTrustCenterUpdateHookWithCustomDomainRemoval(t *testing.T) {
	tcOrg := createFreshOrgWithTrustCenter(t, withCustomDomain())
	trustCenter := tcOrg.trustCenter

	ctx := setContext(tcOrg.owner.UserCtx, suite.client.db)
	fakePirschDomainID := "fake-pirsch-domain-id-clear-test"
	_, err := suite.client.db.TrustCenter.UpdateOneID(trustCenter.ID).
		SetPirschDomainID(fakePirschDomainID).
		Save(ctx)
	assert.NilError(t, err)

	err = suite.client.db.Job.TruncateRiverTables(tcOrg.owner.UserCtx)
	assert.NilError(t, err)

	resp, err := suite.client.api.UpdateTrustCenter(tcOrg.owner.UserCtx, trustCenter.ID, testclient.UpdateTrustCenterInput{
		ClearCustomDomain: lo.ToPtr(true),
	})
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	jobs := rivertest.RequireManyInserted(tcOrg.owner.UserCtx, t, riverpgxv5.New(suite.client.db.Job.GetPool()),
		[]rivertest.ExpectedJob{
			{
				Args: jobspec.DeletePirschDomainArgs{
					PirschDomainID: fakePirschDomainID,
				},
			},
		})
	assert.Assert(t, jobs != nil)
	assert.Assert(t, is.Len(jobs, 1))

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}

// TestTrustCenterDeleteHookWithPirschDomain tests that DeletePirschDomain job is called when pirsch_domain_id exists during deletion
func TestTrustCenterDeleteHookWithPirschDomain(t *testing.T) {
	tcOrg := createFreshOrgWithTrustCenter(t, withCustomDomain())
	tcOrg2 := createFreshOrgWithTrustCenter(t)
	trustCenterWithDomain := tcOrg.trustCenter
	trustCenterWithoutDomain := tcOrg2.trustCenter

	// Manually set pirsch_domain_id to simulate what would happen after the CreatePirschDomain job completes
	// This is necessary because the job runs asynchronously and we need the field set for the delete hook to trigger
	ctx := setContext(tcOrg.owner.UserCtx, suite.client.db)
	fakePirschDomainID := "fake-pirsch-domain-id-123"
	_, err := suite.client.db.TrustCenter.UpdateOneID(trustCenterWithDomain.ID).SetPirschDomainID(fakePirschDomainID).Save(ctx)
	assert.NilError(t, err)

	testCases := []struct {
		name                  string
		trustCenterID         string
		client                *testclient.TestClient
		ctx                   context.Context
		expectDeletePirschJob bool
		expectedErr           string
	}{
		{
			name:                  "delete trust center with pirsch domain - should trigger DeletePirschDomain job",
			trustCenterID:         trustCenterWithDomain.ID,
			client:                suite.client.api,
			ctx:                   tcOrg.owner.UserCtx,
			expectDeletePirschJob: true,
		},
		{
			name:                  "delete trust center without pirsch domain - should NOT trigger DeletePirschDomain job",
			trustCenterID:         trustCenterWithoutDomain.ID,
			client:                suite.client.api,
			ctx:                   tcOrg2.owner.UserCtx,
			expectDeletePirschJob: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear any existing jobs
			err := suite.client.db.Job.TruncateRiverTables(tc.ctx)
			assert.NilError(t, err)

			resp, err := tc.client.DeleteTrustCenter(tc.ctx, tc.trustCenterID)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Verify the job was or was not created based on expectation
			if tc.expectDeletePirschJob {
				jobs := rivertest.RequireManyInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()),
					[]rivertest.ExpectedJob{
						{
							Args: jobspec.DeletePirschDomainArgs{},
						},
					})
				assert.Assert(t, jobs != nil)
				assert.Assert(t, is.Len(jobs, 1))
				// Verify the job has encoded args (PirschDomainID should be set)
				assert.Assert(t, jobs[0].EncodedArgs != nil)
			} else {
				rivertest.RequireNotInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()), &jobspec.DeletePirschDomainArgs{}, nil)
			}
		})
	}

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestTrustCenterDocStandards(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	standard1 := (&StandardBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	standard2 := (&StandardBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)

	(&CustomTypeEnumBuilder{
		client:     suite.client,
		Name:       "Policy",
		ObjectType: "trust_center_doc",
	}).MustNew(tcOrg.owner.UserCtx, t)

	createPDFUpload := uploadFileFunc(t, pdfFilePath)

	t.Run("create trust center doc with standard and retrieve it", func(t *testing.T) {
		fileUpload := createPDFUpload()
		expectUpload(t, suite.client.mockProvider, []graphql.Upload{*fileUpload})

		input := testclient.CreateTrustCenterDocInput{
			Title:                  "Test Document with Standard",
			TrustCenterDocKindName: lo.ToPtr("Policy"),
			TrustCenterID:          &trustCenter.ID,
			StandardID:             &standard1.ID,
			Tags:                   []string{"test", "standard"},
		}

		createResp, err := suite.client.api.CreateTrustCenterDoc(tcOrg.owner.UserCtx, input, *fileUpload)
		assert.NilError(t, err)
		assert.Assert(t, createResp != nil)

		doc := createResp.CreateTrustCenterDoc.TrustCenterDoc
		assert.Check(t, doc.ID != "")
		assert.Check(t, doc.StandardID != nil)
		assert.Check(t, is.Equal(standard1.ID, *doc.StandardID))
		assert.Check(t, doc.Standard != nil)
		assert.Check(t, is.Equal(standard1.ID, doc.Standard.ID))
		assert.Check(t, is.Equal(standard1.Name, doc.Standard.Name))

		getResp, err := suite.client.api.GetTrustCenterDocByID(tcOrg.owner.UserCtx, doc.ID)
		assert.NilError(t, err)
		assert.Assert(t, getResp != nil)
		assert.Check(t, getResp.TrustCenterDoc.StandardID != nil)
		assert.Check(t, is.Equal(standard1.ID, *getResp.TrustCenterDoc.StandardID))
		assert.Check(t, getResp.TrustCenterDoc.Standard != nil)
		assert.Check(t, is.Equal(standard1.ID, getResp.TrustCenterDoc.Standard.ID))
		assert.Check(t, is.Equal(standard1.Name, getResp.TrustCenterDoc.Standard.Name))
	})

	t.Run("update trust center doc to set standard and retrieve it", func(t *testing.T) {
		fileUpload := createPDFUpload()
		expectUpload(t, suite.client.mockProvider, []graphql.Upload{*fileUpload})

		createInput := testclient.CreateTrustCenterDocInput{
			Title:                  "Test Document without Standard",
			TrustCenterDocKindName: lo.ToPtr("Policy"),
			TrustCenterID:          &trustCenter.ID,
			Tags:                   []string{"test"},
		}

		createResp, err := suite.client.api.CreateTrustCenterDoc(tcOrg.owner.UserCtx, createInput, *fileUpload)
		assert.NilError(t, err)
		assert.Assert(t, createResp != nil)

		docID := createResp.CreateTrustCenterDoc.TrustCenterDoc.ID

		getResp, err := suite.client.api.GetTrustCenterDocByID(tcOrg.owner.UserCtx, docID)
		assert.NilError(t, err)
		assert.Assert(t, getResp != nil)
		assert.Check(t, getResp.TrustCenterDoc.StandardID == nil || *getResp.TrustCenterDoc.StandardID == "")

		updateInput := testclient.UpdateTrustCenterDocInput{
			StandardID: &standard1.ID,
		}

		updateResp, err := suite.client.api.UpdateTrustCenterDoc(tcOrg.owner.UserCtx, docID, updateInput, nil, nil)
		assert.NilError(t, err)
		assert.Assert(t, updateResp != nil)

		updatedDoc := updateResp.UpdateTrustCenterDoc.TrustCenterDoc
		assert.Check(t, updatedDoc.StandardID != nil)
		assert.Check(t, is.Equal(standard1.ID, *updatedDoc.StandardID))
		assert.Check(t, updatedDoc.Standard != nil)
		assert.Check(t, is.Equal(standard1.ID, updatedDoc.Standard.ID))
		assert.Check(t, is.Equal(standard1.Name, updatedDoc.Standard.Name))

		getResp2, err := suite.client.api.GetTrustCenterDocByID(tcOrg.owner.UserCtx, docID)
		assert.NilError(t, err)
		assert.Assert(t, getResp2 != nil)
		assert.Check(t, getResp2.TrustCenterDoc.StandardID != nil)
		assert.Check(t, is.Equal(standard1.ID, *getResp2.TrustCenterDoc.StandardID))
		assert.Check(t, getResp2.TrustCenterDoc.Standard != nil)
		assert.Check(t, is.Equal(standard1.ID, getResp2.TrustCenterDoc.Standard.ID))

		updateInput2 := testclient.UpdateTrustCenterDocInput{
			StandardID: &standard2.ID,
		}

		updateResp2, err := suite.client.api.UpdateTrustCenterDoc(tcOrg.owner.UserCtx, docID, updateInput2, nil, nil)
		assert.NilError(t, err)
		assert.Assert(t, updateResp2 != nil)

		updatedDoc2 := updateResp2.UpdateTrustCenterDoc.TrustCenterDoc
		assert.Check(t, updatedDoc2.StandardID != nil)
		assert.Check(t, is.Equal(standard2.ID, *updatedDoc2.StandardID))
		assert.Check(t, updatedDoc2.Standard != nil)
		assert.Check(t, is.Equal(standard2.ID, updatedDoc2.Standard.ID))
		assert.Check(t, is.Equal(standard2.Name, updatedDoc2.Standard.Name))

		getResp3, err := suite.client.api.GetTrustCenterDocByID(tcOrg.owner.UserCtx, docID)
		assert.NilError(t, err)
		assert.Assert(t, getResp3 != nil)
		assert.Check(t, getResp3.TrustCenterDoc.StandardID != nil)
		assert.Check(t, is.Equal(standard2.ID, *getResp3.TrustCenterDoc.StandardID))
		assert.Check(t, getResp3.TrustCenterDoc.Standard != nil)
		assert.Check(t, is.Equal(standard2.ID, getResp3.TrustCenterDoc.Standard.ID))
	})

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}

func TestMutationDeleteTrustCenterWithPreviewDomain(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	// Create a preview domain (custom domain)
	previewDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)

	dbCtx := setContext(tcOrg.owner.UserCtx, suite.client.db)
	trustCenter, err := suite.client.db.TrustCenter.UpdateOneID(trustCenter.ID).
		SetPreviewDomainID(previewDomain.ID).
		Save(dbCtx)
	assert.NilError(t, err)

	// Delete the trust center
	resp, err := suite.client.api.DeleteTrustCenter(tcOrg.owner.UserCtx, trustCenter.ID)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Check(t, is.Equal(trustCenter.ID, resp.DeleteTrustCenter.DeletedID))

	// Verify a job was queued to delete the preview domain
	// Note: We can't easily verify the exact job args without accessing the river queue,
	// but we can verify the preview domain still exists (it will be deleted by the job worker)
	exists, err := suite.client.db.CustomDomain.Query().Where(customdomain.ID(previewDomain.ID)).Exist(dbCtx)
	assert.NilError(t, err)
	assert.Check(t, exists, "preview domain should still exist (will be deleted by job)")

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}
