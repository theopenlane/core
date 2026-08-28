package graphapi_test

import (
	"context"
	"fmt"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

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
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/customdomain"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestQueryTrustCenterByID(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.TrustCenter

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
			client:  suite.Client.API,
			ctx:     tcOrg.Owner.UserCtx,
		},
		{
			name:    "happy path, view only user",
			queryID: trustCenter.ID,
			client:  suite.Client.API,
			ctx:     tcOrg.Member.UserCtx,
		},
		{
			name:     "trust center not found",
			queryID:  "non-existent-id",
			client:   suite.Client.API,
			ctx:      tcOrg.Owner.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "not authorized to query org",
			queryID:  trustCenter.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
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
			assert.Check(t, is.Equal(tcOrg.OrganizationID, *resp.TrustCenter.OwnerID))

			setting := resp.TrustCenter.GetSetting()
			assert.Assert(t, setting != nil)
			assert.Check(t, setting.Title != nil)
			assert.Check(t, setting.Overview != nil)
			assert.Check(t, setting.PrimaryColor != nil)
		})
	}

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestQueryTrustCenters(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithCustomDomain(), th.WithAllUserTypes())
	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t)
	trustCenter1 := tcOrg.TrustCenter

	nonExistentSlug := "nonexistent-slug"

	if trustCenter1.CustomDomainID == nil {
		th.FailNow(t, "expected trust center custom domain but no ID was returned")

	}
	customDomainTrustCenter1, err := suite.Client.API.GetCustomDomainByID(tcOrg.Owner.UserCtx, *trustCenter1.CustomDomainID)
	th.RequireNoError(t, err)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int64
		where           *testclient.TrustCenterWhereInput
	}{
		{
			name:            "return all",
			client:          suite.Client.API,
			ctx:             tcOrg.Owner.UserCtx,
			expectedResults: 1,
		},
		{
			name:            "return all, ro user",
			client:          suite.Client.API,
			ctx:             tcOrg.Member.UserCtx,
			expectedResults: 1,
		},
		{
			name:   "query by org ID",
			client: suite.Client.API,
			ctx:    tcOrg.SuperAdmin.UserCtx,
			where: &testclient.TrustCenterWhereInput{
				OwnerID: &tcOrg.OrganizationID,
			},
			expectedResults: 1,
		},
		{
			name:   "query by slug",
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
			where: &testclient.TrustCenterWhereInput{
				Slug: &trustCenter1.Slug,
			},
			expectedResults: 1,
		},
		{
			name:   "query by slug, not found",
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
			where: &testclient.TrustCenterWhereInput{
				Slug: &nonExistentSlug,
			},
			expectedResults: 0,
		},
		{
			name:   "query by custom domain, slug",
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
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
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
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
				assert.Check(t, is.Equal(tcOrg.OrganizationID, *node.Node.OwnerID))
				setting := node.Node.GetSetting()
				assert.Assert(t, setting != nil)
				assert.Check(t, setting.Title != nil)
				assert.Check(t, setting.Overview != nil)
				assert.Check(t, setting.PrimaryColor != nil)
			}

		})
	}

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(tcOrg2.Owner.UserCtx, t)
}

func TestMutationCreateTrustCenter(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithCustomDomain())
	customDomainAnotherOrg, err := suite.Client.API.GetCustomDomainByID(tcOrg.Owner.UserCtx, *tcOrg.TrustCenter.CustomDomainID)
	th.RequireNoError(t, err)

	localTestUser := suite.SeedFreshMinimalOrgUsers(t, false)
	customDomain := (&th.CustomDomainBuilder{Client: suite.Client}).MustNew(localTestUser.Owner.UserCtx, t)

	// create trust center standard
	trustCenterControlStd := (&th.StandardBuilder{Client: suite.Client, Name: "OTS", Framework: "openlane-trust-center", IsPublic: true}).MustNew(th.SharedSystemAdminUser.UserCtx, t)

	trustCenterControlIDs := []string{}
	numTrustCenterControls := 5
	for range numTrustCenterControls {
		control := (&th.ControlBuilder{Client: suite.Client, StandardID: trustCenterControlStd.ID}).MustNew(th.SharedSystemAdminUser.UserCtx, t)
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
			client:  suite.Client.API,
			ctx:     localTestUser.Owner.UserCtx,
		},
		{
			name: "custom domain for different organization should error",
			request: testclient.CreateTrustCenterInput{
				CustomDomainID: &customDomainAnotherOrg.CustomDomain.ID,
			},
			client:      suite.Client.API,
			ctx:         localTestUser.Owner.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "custom domain setting",
			request: testclient.CreateTrustCenterInput{
				CustomDomainID: &customDomain.ID,
			},
			client: suite.Client.API,
			ctx:    localTestUser.Owner.UserCtx,
		},
		{
			name: "happy path with settings for different organization",
			request: testclient.CreateTrustCenterInput{
				CreateTrustCenterSetting: &testclient.CreateTrustCenterSettingInput{
					Title: lo.ToPtr(gofakeit.Name()),
				},
			},
			client: suite.Client.API,
			ctx:    localTestUser.Owner.UserCtx,
		},
		{
			name:        "not authorized",
			request:     testclient.CreateTrustCenterInput{},
			client:      suite.Client.API,
			ctx:         localTestUser.Member.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:        "duplicate trust center for same organization",
			request:     testclient.CreateTrustCenterInput{},
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
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
			dbCtx := th.SetContext(tc.ctx, suite.Client.DB)
			org, err := suite.Client.DB.Organization.Get(dbCtx, *resp.CreateTrustCenter.TrustCenter.OwnerID)
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
			(&th.Cleanup[*generated.TrustCenterDeleteOne]{Client: suite.Client.DB.TrustCenter, ID: resp.CreateTrustCenter.TrustCenter.ID}).MustDelete(tc.ctx, t)
		})
	}

	// Clean up the existing trust center
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(localTestUser.Owner.UserCtx, t)
}

func TestGetAllTrustCenters(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAllUserTypes())
	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int64
		expectedErr     string
	}{
		{
			name:            "happy path - regular user sees only their trust centers",
			client:          suite.Client.API,
			ctx:             tcOrg.Owner.UserCtx,
			expectedResults: 1,
		},
		{
			name:            "happy path - admin user sees all trust centers",
			client:          suite.Client.API,
			ctx:             tcOrg.Admin.UserCtx,
			expectedResults: 1,
		},
		{
			name:            "happy path - view only user",
			client:          suite.Client.API,
			ctx:             tcOrg.Member.UserCtx,
			expectedResults: 1,
		},
		{
			name:            "happy path - different user sees only their trust centers",
			client:          suite.Client.API,
			ctx:             tcOrg2.Owner.UserCtx,
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
			case tcOrg.Owner.UserCtx, tcOrg.Admin.UserCtx, tcOrg.Member.UserCtx, tcOrg.SuperAdmin.UserCtx:
				for _, edge := range resp.TrustCenters.Edges {
					assert.Check(t, is.Equal(tcOrg.OrganizationID, *edge.Node.OwnerID))
				}
			case tcOrg2.Owner.UserCtx:
				for _, edge := range resp.TrustCenters.Edges {
					assert.Check(t, is.Equal(tcOrg2.OrganizationID, *edge.Node.OwnerID))
				}
			}
		})
	}

	// Clean up created trust centers
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestMutationUpdateTrustCenter(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithCustomDomain(), th.WithAllUserTypes())
	trustCenter := tcOrg.TrustCenter

	if trustCenter.CustomDomainID == nil {
		th.FailNow(t, "expected trust center custom domain but no ID was returned")

	}
	customDomainTrustCenter, err := suite.Client.API.GetCustomDomainByID(tcOrg.Owner.UserCtx, *trustCenter.CustomDomainID)
	th.RequireNoError(t, err)

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
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
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
			client: suite.Client.API,
			ctx:    tcOrg.SuperAdmin.UserCtx,
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
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
		},
		{
			name:          "happy path, append tags",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				AppendTags: []string{"appended", "tag"},
			},
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
		},
		{
			name:          "happy path, using admin user",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"admin", "update"},
			},
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
		},
		{
			name:          "happy path, using personal access token",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"pat", "update"},
			},
			client: tcOrg.AdminPatClient,
			ctx:    context.Background(),
		},
		{
			name:          "not authorized, view only user",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"unauthorized"},
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Member.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:          "not authorized, different org user",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"unauthorized"},
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:          "trust center not found",
			trustCenterID: "non-existent-id",
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"test"},
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestMutationDeleteTrustCenter(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t)
	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t)

	trustCenter1 := tcOrg.TrustCenter
	trustCenter2 := tcOrg2.TrustCenter

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
			client:     suite.Client.API,
			ctx:        tcOrg.Owner.UserCtx,
		},
		{
			name:        "not authorized, different org user",
			idToDelete:  trustCenter2.ID,
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "trust center not found",
			idToDelete:  "non-existent-id",
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(tcOrg2.Owner.UserCtx, t)
}

func TestQueryTrustCenterAsAnonymousUser(t *testing.T) {
	t.Parallel()
	// create new test users
	tcOrg := th.CreateFreshOrgWithTrustCenter(t)
	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t)

	trustCenter := tcOrg.TrustCenter

	// create trust center entities for the trust center
	createLogoUpload := th.LogoFileFunc(t)
	logoFile := createLogoUpload()

	th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*logoFile})

	_, err := suite.Client.API.CreateTrustCenterEntity(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterEntityInput{
		Name:          "test entity 1",
		TrustCenterID: &trustCenter.ID,
		URL:           lo.ToPtr(gofakeit.URL()),
	}, logoFile, nil)
	assert.NilError(t, err)

	_, err = suite.Client.API.UpdateTrustCenter(tcOrg.Owner.UserCtx, trustCenter.ID, testclient.UpdateTrustCenterInput{
		AddPost: &testclient.CreateNoteInput{
			Text: "this is an update",
		},
	})
	assert.NilError(t, err)

	// create trust center compliance
	std := (&th.StandardBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
	_, err = suite.Client.API.CreateTrustCenterCompliance(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterComplianceInput{
		StandardID: std.ID,
	})
	assert.NilError(t, err)

	// create subprocessor
	sbpr := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
	sbprKind := (&th.CustomTypeEnumBuilder{
		Client:     suite.Client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg.Owner.UserCtx, t)
	_, err = suite.Client.API.CreateTrustCenterSubprocessor(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
		SubprocessorID:                  sbpr.ID,
		TrustCenterSubprocessorKindName: &sbprKind.Name,
		Countries:                       []string{"United States"},
	})
	assert.NilError(t, err)

	// create custom type enum for trust center doc kind
	docKind := (&th.CustomTypeEnumBuilder{
		Client:     suite.Client,
		ObjectType: "trust_center_doc",
	}).MustNew(tcOrg.Owner.UserCtx, t)

	// create trust center doc
	createFileUpload := th.UploadFileFunc(t, th.PdfFilePath)
	fileUpload := createFileUpload()

	th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*fileUpload})
	doc, err := suite.Client.API.CreateTrustCenterDoc(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterDocInput{
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
	faqNote := (&th.NoteBuilder{Client: suite.Client, TrustCenterID: trustCenter.ID}).MustNew(tcOrg.Owner.UserCtx, t)
	_, err = suite.Client.API.CreateTrustCenterFaq(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterFAQInput{
		NoteID:        faqNote.ID,
		TrustCenterID: &trustCenter.ID,
		ReferenceLink: lo.ToPtr("https://example.com/faq"),
		DisplayOrder:  lo.ToPtr(int64(1)),
	})
	assert.NilError(t, err)

	// Create another trust center that the anonymous user should NOT have access to
	trustCenter2 := tcOrg2.TrustCenter

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
			organizationID: tcOrg.OrganizationID,
			client:         suite.Client.API,
			shouldSucceed:  true,
			isList:         true,
		},
		{
			name:           "anonymous user cannot query different trust center by id",
			queryID:        trustCenter2.ID,
			trustCenterID:  trustCenter.ID, // Anonymous user has access to trustCenter, not trustCenter2
			organizationID: tcOrg.OrganizationID,
			client:         suite.Client.API,
			expectedErr:    th.NotFoundErrorMsg,
			shouldSucceed:  false,
		},
		{
			name:           "anonymous user cannot query non-existent trust center by id",
			queryID:        "non-existent-id",
			trustCenterID:  trustCenter.ID,
			organizationID: tcOrg.OrganizationID,
			client:         suite.Client.API,
			expectedErr:    th.NotFoundErrorMsg,
			shouldSucceed:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create anonymous trust center context
			anonCtx := th.CreateAnonymousTrustCenterContext(tc.trustCenterID, tc.organizationID)

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
	dbCtx := th.SetContext(tcOrg.Owner.UserCtx, suite.Client.DB)

	tcControl, err := suite.Client.DB.Control.Create().
		SetRefCode("OTS-TC-" + ulids.New().String()).
		SetTitle("Trust Center Control").
		SetSource(enums.ControlSourceUserDefined).
		SetIsTrustCenterControl(true).
		SetOwnerID(tcOrg.OrganizationID).
		Save(dbCtx)
	assert.NilError(t, err)

	_, err = suite.Client.API.UpdateControl(tcOrg.Owner.UserCtx, tcControl.ID, testclient.UpdateControlInput{
		TrustCenterVisibility: &enums.TrustCenterControlVisibilityPubliclyVisible,
	})
	assert.NilError(t, err)

	// create another trust center control for another trust center to ensure only controls for the queried trust center are returned in the frontend query
	dbCtx2 := th.SetContext(tcOrg2.Owner.UserCtx, suite.Client.DB)
	tcControlForAnotherOrg, err := suite.Client.DB.Control.Create().
		SetRefCode("OTS-TC-" + ulids.New().String()).
		SetTitle("Trust Center Control").
		SetSource(enums.ControlSourceUserDefined).
		SetIsTrustCenterControl(true).
		SetOwnerID(tcOrg2.OrganizationID).
		Save(dbCtx2)
	assert.NilError(t, err)

	_, err = suite.Client.API.UpdateControl(tcOrg2.Owner.UserCtx, tcControlForAnotherOrg.ID, testclient.UpdateControlInput{
		TrustCenterVisibility: &enums.TrustCenterControlVisibilityPubliclyVisible,
	})
	assert.NilError(t, err)

	t.Run("anonymous user frontend query returns all child objects with controls present", func(t *testing.T) {
		anonCtx := th.CreateAnonymousTrustCenterContext(trustCenter.ID, tcOrg.Owner.OrganizationID)

		resp, err := suite.Client.API.GetTrustCenterFrontendQuery(anonCtx)
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
		anonCtx := th.CreateAnonymousTrustCenterContext(trustCenter.ID, tcOrg.OrganizationID)

		resp, err := suite.Client.API.GetTrustCenterControls(anonCtx)
		assert.NilError(t, err)
		assert.Check(t, resp != nil)
		assert.Check(t, is.Len(resp.Controls.Edges, 1))
		assert.Check(t, resp.Controls.Edges[0].Node.ID == tcControl.ID)
		assert.Check(t, resp.Controls.Edges[0].Node.RefCode != "")
	})

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(tcOrg2.Owner.UserCtx, t)
}

func TestQueryTrustCentersAsAnonymousUser(t *testing.T) {
	t.Parallel()
	// create new test users
	tcOrg := th.CreateFreshOrgWithTrustCenter(t)
	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t)

	trustCenter := tcOrg.TrustCenter
	trustCenter2 := tcOrg2.TrustCenter

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
			organizationID: tcOrg.OrganizationID,
			client:         suite.Client.API,
			expectedCount:  1, // Should only see the one trust center they have access to
		},
		{
			name:           "anonymous user with different trust center sees only their trust center",
			trustCenterID:  trustCenter2.ID,
			organizationID: tcOrg2.OrganizationID,
			client:         suite.Client.API,
			expectedCount:  1, // Should only see the one trust center they have access to
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create anonymous trust center context
			anonCtx := th.CreateAnonymousTrustCenterContext(tc.trustCenterID, tc.organizationID)

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
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(tcOrg2.Owner.UserCtx, t)
}

func TestMutationUpdateTrustCenterSetting(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAllUserTypes())
	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t)

	trustCenter := tcOrg.TrustCenter

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
			logoPath:    th.LogoFilePath,
			updateInput: testclient.UpdateTrustCenterSettingInput{},
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
		},
		{
			name:      "happy path - update logo with other fields",
			settingID: trustCenter.Edges.Setting.ID,
			logoPath:  th.LogoFilePath,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				Title:        lo.ToPtr("Updated Title with Logo"),
				PrimaryColor: lo.ToPtr("#FF5733"),
			},
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
		},

		{
			name:        "invalid file type - text file instead of image",
			settingID:   trustCenter.Edges.Setting.ID,
			logoPath:    th.TxtFilePath,
			invalidFile: true,
			updateInput: testclient.UpdateTrustCenterSettingInput{},
			client:      suite.Client.API,
			ctx:         tcOrg.SuperAdmin.UserCtx,
			expectedErr: "unsupported mime type uploaded: text/plain",
		},
		{
			name:        "not authorized - view only user",
			settingID:   trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{},
			client:      suite.Client.API,
			ctx:         tcOrg.Member.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:        "not authorized - different organization user",
			settingID:   trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{},
			client:      suite.Client.API,
			ctx:         tcOrg2.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "trust center setting not found",
			settingID:   "non-existent-setting-id",
			updateInput: testclient.UpdateTrustCenterSettingInput{},
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
		},
		{
			name:      "happy path - update theme mode to EASY",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				ThemeMode: lo.ToPtr(enums.TrustCenterThemeModeEasy),
			},
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
		},
		{
			name:      "happy path - update theme mode to ADVANCED",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				ThemeMode: lo.ToPtr(enums.TrustCenterThemeModeAdvanced),
			},
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
		},
		{
			name:      "happy path - update font",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				Font: lo.ToPtr("Arial, sans-serif"),
			},
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
		},
		{
			name:      "happy path - update foreground color",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				ForegroundColor: lo.ToPtr("#333333"),
			},
			client: suite.Client.API,
			ctx:    tcOrg.SuperAdmin.UserCtx,
		},
		{
			name:      "happy path - update background color",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				BackgroundColor: lo.ToPtr("#FFFFFF"),
			},
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
		},
		{
			name:      "happy path - update accent color",
			settingID: trustCenter.Edges.Setting.ID,
			updateInput: testclient.UpdateTrustCenterSettingInput{
				AccentColor: lo.ToPtr("#007BFF"),
			},
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
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
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			var logoFile *graphql.Upload

			// Create file upload if logoPath is provided
			if tc.logoPath != "" {
				logoFile = th.UploadFile(t, tc.logoPath)

				// Set up mock expectations based on whether we expect an error
				if tc.expectedErr == "" {
					th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*logoFile})
				} else {
					th.ExpectUploadCheckOnly(t, suite.Client.MockProvider)
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
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

// TestTrustCenterCreateHookWithCustomDomain tests that CreatePirschDomain job is called when custom_domain_id is set during creation
func TestTrustCenterCreateHookWithCustomDomain(t *testing.T) {
	users := suite.SeedFreshOrgUsers(t)

	customDomain := (&th.CustomDomainBuilder{Client: suite.Client}).MustNew(users.Owner.UserCtx, t)

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
			client:                suite.Client.API,
			ctx:                   users.Owner.UserCtx,
			expectCreatePirschJob: true,
		},
		{
			name:                  "create trust center without custom domain - should NOT trigger CreatePirschDomain job",
			request:               testclient.CreateTrustCenterInput{},
			client:                suite.Client.API,
			ctx:                   users.Owner.UserCtx,
			expectCreatePirschJob: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear any existing jobs
			err := suite.Client.DB.Job.TruncateRiverTables(tc.ctx)
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
				jobs := rivertest.RequireManyInserted(tc.ctx, t, riverpgxv5.New(suite.Client.DB.Job.GetPool()),
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
				rivertest.RequireNotInserted(tc.ctx, t, riverpgxv5.New(suite.Client.DB.Job.GetPool()), &jobspec.CreatePirschDomainArgs{}, nil)
			}

			// Clean up
			(&th.Cleanup[*generated.TrustCenterDeleteOne]{Client: suite.Client.DB.TrustCenter, ID: resp.CreateTrustCenter.TrustCenter.ID}).MustDelete(tc.ctx, t)
		})
	}

	// Clean up custom domain
	th.CleanupOrganizationDataWithContext(users.Owner.UserCtx, t)
}

// TestTrustCenterUpdateHookWithCustomDomain tests that CreatePirschDomain job is called when custom_domain_id changes from empty to non-empty
func TestTrustCenterUpdateHookWithCustomDomain(t *testing.T) {
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAllUserTypes())
	trustCenter := tcOrg.TrustCenter

	customDomain := (&th.CustomDomainBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)

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
			client:                suite.Client.API,
			ctx:                   tcOrg.Owner.UserCtx,
			expectCreatePirschJob: true,
		},
		{
			name:          "update trust center without changing custom domain - should NOT trigger CreatePirschDomain job",
			trustCenterID: trustCenter.ID,
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"test", "tag"},
			},
			client:                suite.Client.API,
			ctx:                   tcOrg.SuperAdmin.UserCtx,
			expectCreatePirschJob: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear any existing jobs
			err := suite.Client.DB.Job.TruncateRiverTables(tc.ctx)
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
				jobs := rivertest.RequireManyInserted(tc.ctx, t, riverpgxv5.New(suite.Client.DB.Job.GetPool()),
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
				rivertest.RequireNotInserted(tc.ctx, t, riverpgxv5.New(suite.Client.DB.Job.GetPool()), &jobspec.CreatePirschDomainArgs{}, nil)
			}
		})
	}

	// Clean up
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

// TestTrustCenterUpdateHookWithPirschDomainUpdate tests that UpdatePirschDomain job is called when custom_domain_id changes from one domain to another
func TestTrustCenterUpdateHookWithPirschDomainUpdate(t *testing.T) {
	tcOrgWithDomain := th.CreateFreshOrgWithTrustCenter(t, th.WithCustomDomain())
	trustCenterWithDomain := tcOrgWithDomain.TrustCenter

	// Create two custom domains
	customDomain2 := (&th.CustomDomainBuilder{Client: suite.Client}).MustNew(tcOrgWithDomain.Owner.UserCtx, t)

	// Manually set pirsch_domain_id to simulate what would happen after the CreatePirschDomain job completes
	ctx := th.SetContext(tcOrgWithDomain.Owner.UserCtx, suite.Client.DB)
	fakePirschDomainID := "fake-pirsch-domain-id-for-update-test"
	_, err := suite.Client.DB.TrustCenter.UpdateOneID(trustCenterWithDomain.ID).SetPirschDomainID(fakePirschDomainID).Save(ctx)
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
			client:                suite.Client.API,
			ctx:                   tcOrgWithDomain.Owner.UserCtx,
			expectUpdatePirschJob: true,
		},
		{
			name:          "update trust center without changing custom domain - should NOT trigger UpdatePirschDomain job",
			trustCenterID: trustCenterWithDomain.ID,
			request: testclient.UpdateTrustCenterInput{
				Tags: []string{"test", "tag"},
			},
			client:                suite.Client.API,
			ctx:                   tcOrgWithDomain.Owner.UserCtx,
			expectUpdatePirschJob: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear any existing jobs
			err := suite.Client.DB.Job.TruncateRiverTables(tc.ctx)
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
				jobs := rivertest.RequireManyInserted(tc.ctx, t, riverpgxv5.New(suite.Client.DB.Job.GetPool()),
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
				rivertest.RequireNotInserted(tc.ctx, t, riverpgxv5.New(suite.Client.DB.Job.GetPool()), &jobspec.UpdatePirschDomainArgs{}, nil)
			}
		})
	}

	// Clean up
	th.CleanupOrganizationDataWithContext(tcOrgWithDomain.Owner.UserCtx, t)
}

// TestTrustCenterUpdateHookWithCustomDomainRemoval tests that DeletePirschDomain job is called when custom_domain_id is cleared
func TestTrustCenterUpdateHookWithCustomDomainRemoval(t *testing.T) {
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithCustomDomain())
	trustCenter := tcOrg.TrustCenter

	ctx := th.SetContext(tcOrg.Owner.UserCtx, suite.Client.DB)
	fakePirschDomainID := "fake-pirsch-domain-id-clear-test"
	_, err := suite.Client.DB.TrustCenter.UpdateOneID(trustCenter.ID).
		SetPirschDomainID(fakePirschDomainID).
		Save(ctx)
	assert.NilError(t, err)

	err = suite.Client.DB.Job.TruncateRiverTables(tcOrg.Owner.UserCtx)
	assert.NilError(t, err)

	resp, err := suite.Client.API.UpdateTrustCenter(tcOrg.Owner.UserCtx, trustCenter.ID, testclient.UpdateTrustCenterInput{
		ClearCustomDomain: lo.ToPtr(true),
	})
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	jobs := rivertest.RequireManyInserted(tcOrg.Owner.UserCtx, t, riverpgxv5.New(suite.Client.DB.Job.GetPool()),
		[]rivertest.ExpectedJob{
			{
				Args: jobspec.DeletePirschDomainArgs{
					PirschDomainID: fakePirschDomainID,
				},
			},
		})
	assert.Assert(t, jobs != nil)
	assert.Assert(t, is.Len(jobs, 1))

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

// TestTrustCenterDeleteHookWithPirschDomain tests that DeletePirschDomain job is called when pirsch_domain_id exists during deletion
func TestTrustCenterDeleteHookWithPirschDomain(t *testing.T) {
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithCustomDomain())
	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t)
	trustCenterWithDomain := tcOrg.TrustCenter
	trustCenterWithoutDomain := tcOrg2.TrustCenter

	// Manually set pirsch_domain_id to simulate what would happen after the CreatePirschDomain job completes
	// This is necessary because the job runs asynchronously and we need the field set for the delete hook to trigger
	ctx := th.SetContext(tcOrg.Owner.UserCtx, suite.Client.DB)
	fakePirschDomainID := "fake-pirsch-domain-id-123"
	_, err := suite.Client.DB.TrustCenter.UpdateOneID(trustCenterWithDomain.ID).SetPirschDomainID(fakePirschDomainID).Save(ctx)
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
			client:                suite.Client.API,
			ctx:                   tcOrg.Owner.UserCtx,
			expectDeletePirschJob: true,
		},
		{
			name:                  "delete trust center without pirsch domain - should NOT trigger DeletePirschDomain job",
			trustCenterID:         trustCenterWithoutDomain.ID,
			client:                suite.Client.API,
			ctx:                   tcOrg2.Owner.UserCtx,
			expectDeletePirschJob: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear any existing jobs
			err := suite.Client.DB.Job.TruncateRiverTables(tc.ctx)
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
				jobs := rivertest.RequireManyInserted(tc.ctx, t, riverpgxv5.New(suite.Client.DB.Job.GetPool()),
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
				rivertest.RequireNotInserted(tc.ctx, t, riverpgxv5.New(suite.Client.DB.Job.GetPool()), &jobspec.DeletePirschDomainArgs{}, nil)
			}
		})
	}

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(tcOrg2.Owner.UserCtx, t)
}

func TestTrustCenterDocStandards(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.TrustCenter

	standard1 := (&th.StandardBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
	standard2 := (&th.StandardBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)

	(&th.CustomTypeEnumBuilder{
		Client:     suite.Client,
		Name:       "Policy",
		ObjectType: "trust_center_doc",
	}).MustNew(tcOrg.Owner.UserCtx, t)

	createPDFUpload := th.UploadFileFunc(t, th.PdfFilePath)

	t.Run("create trust center doc with standard and retrieve it", func(t *testing.T) {
		fileUpload := createPDFUpload()
		th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*fileUpload})

		input := testclient.CreateTrustCenterDocInput{
			Title:                  "Test Document with Standard",
			TrustCenterDocKindName: lo.ToPtr("Policy"),
			TrustCenterID:          &trustCenter.ID,
			StandardID:             &standard1.ID,
			Tags:                   []string{"test", "standard"},
		}

		createResp, err := suite.Client.API.CreateTrustCenterDoc(tcOrg.Owner.UserCtx, input, *fileUpload)
		assert.NilError(t, err)
		assert.Assert(t, createResp != nil)

		doc := createResp.CreateTrustCenterDoc.TrustCenterDoc
		assert.Check(t, doc.ID != "")
		assert.Check(t, doc.StandardID != nil)
		assert.Check(t, is.Equal(standard1.ID, *doc.StandardID))
		assert.Check(t, doc.Standard != nil)
		assert.Check(t, is.Equal(standard1.ID, doc.Standard.ID))
		assert.Check(t, is.Equal(standard1.Name, doc.Standard.Name))

		getResp, err := suite.Client.API.GetTrustCenterDocByID(tcOrg.Owner.UserCtx, doc.ID)
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
		th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*fileUpload})

		createInput := testclient.CreateTrustCenterDocInput{
			Title:                  "Test Document without Standard",
			TrustCenterDocKindName: lo.ToPtr("Policy"),
			TrustCenterID:          &trustCenter.ID,
			Tags:                   []string{"test"},
		}

		createResp, err := suite.Client.API.CreateTrustCenterDoc(tcOrg.Owner.UserCtx, createInput, *fileUpload)
		assert.NilError(t, err)
		assert.Assert(t, createResp != nil)

		docID := createResp.CreateTrustCenterDoc.TrustCenterDoc.ID

		getResp, err := suite.Client.API.GetTrustCenterDocByID(tcOrg.Owner.UserCtx, docID)
		assert.NilError(t, err)
		assert.Assert(t, getResp != nil)
		assert.Check(t, getResp.TrustCenterDoc.StandardID == nil || *getResp.TrustCenterDoc.StandardID == "")

		updateInput := testclient.UpdateTrustCenterDocInput{
			StandardID: &standard1.ID,
		}

		updateResp, err := suite.Client.API.UpdateTrustCenterDoc(tcOrg.Owner.UserCtx, docID, updateInput, nil, nil)
		assert.NilError(t, err)
		assert.Assert(t, updateResp != nil)

		updatedDoc := updateResp.UpdateTrustCenterDoc.TrustCenterDoc
		assert.Check(t, updatedDoc.StandardID != nil)
		assert.Check(t, is.Equal(standard1.ID, *updatedDoc.StandardID))
		assert.Check(t, updatedDoc.Standard != nil)
		assert.Check(t, is.Equal(standard1.ID, updatedDoc.Standard.ID))
		assert.Check(t, is.Equal(standard1.Name, updatedDoc.Standard.Name))

		getResp2, err := suite.Client.API.GetTrustCenterDocByID(tcOrg.Owner.UserCtx, docID)
		assert.NilError(t, err)
		assert.Assert(t, getResp2 != nil)
		assert.Check(t, getResp2.TrustCenterDoc.StandardID != nil)
		assert.Check(t, is.Equal(standard1.ID, *getResp2.TrustCenterDoc.StandardID))
		assert.Check(t, getResp2.TrustCenterDoc.Standard != nil)
		assert.Check(t, is.Equal(standard1.ID, getResp2.TrustCenterDoc.Standard.ID))

		updateInput2 := testclient.UpdateTrustCenterDocInput{
			StandardID: &standard2.ID,
		}

		updateResp2, err := suite.Client.API.UpdateTrustCenterDoc(tcOrg.Owner.UserCtx, docID, updateInput2, nil, nil)
		assert.NilError(t, err)
		assert.Assert(t, updateResp2 != nil)

		updatedDoc2 := updateResp2.UpdateTrustCenterDoc.TrustCenterDoc
		assert.Check(t, updatedDoc2.StandardID != nil)
		assert.Check(t, is.Equal(standard2.ID, *updatedDoc2.StandardID))
		assert.Check(t, updatedDoc2.Standard != nil)
		assert.Check(t, is.Equal(standard2.ID, updatedDoc2.Standard.ID))
		assert.Check(t, is.Equal(standard2.Name, updatedDoc2.Standard.Name))

		getResp3, err := suite.Client.API.GetTrustCenterDocByID(tcOrg.Owner.UserCtx, docID)
		assert.NilError(t, err)
		assert.Assert(t, getResp3 != nil)
		assert.Check(t, getResp3.TrustCenterDoc.StandardID != nil)
		assert.Check(t, is.Equal(standard2.ID, *getResp3.TrustCenterDoc.StandardID))
		assert.Check(t, getResp3.TrustCenterDoc.Standard != nil)
		assert.Check(t, is.Equal(standard2.ID, getResp3.TrustCenterDoc.Standard.ID))
	})

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestMutationDeleteTrustCenterWithPreviewDomain(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.TrustCenter

	// Create a preview domain (custom domain)
	previewDomain := (&th.CustomDomainBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)

	dbCtx := th.SetContext(tcOrg.Owner.UserCtx, suite.Client.DB)
	trustCenter, err := suite.Client.DB.TrustCenter.UpdateOneID(trustCenter.ID).
		SetPreviewDomainID(previewDomain.ID).
		Save(dbCtx)
	assert.NilError(t, err)

	// Delete the trust center
	resp, err := suite.Client.API.DeleteTrustCenter(tcOrg.Owner.UserCtx, trustCenter.ID)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Check(t, is.Equal(trustCenter.ID, resp.DeleteTrustCenter.DeletedID))

	// Verify a job was queued to delete the preview domain
	// Note: We can't easily verify the exact job args without accessing the river queue,
	// but we can verify the preview domain still exists (it will be deleted by the job worker)
	exists, err := suite.Client.DB.CustomDomain.Query().Where(customdomain.ID(previewDomain.ID)).Exist(dbCtx)
	assert.NilError(t, err)
	assert.Check(t, exists, "preview domain should still exist (will be deleted by job)")

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}
