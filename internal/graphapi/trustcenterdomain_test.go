package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/utils/ulids"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/jobspec"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/core/v2/internal/testutils"
)

func TestMutationCreatePreviewDomain(t *testing.T) {
	org := th.CreateFreshOrgWithTrustCenter(t, th.WithAllUserTypes(), th.WithSupportUser())
	secondOrg := th.CreateFreshOrgWithTrustCenter(t)

	defer func() {
		th.CleanupOrganizationDataWithContext(org.Owner.UserCtx, t)
		th.CleanupOrganizationDataWithContext(secondOrg.Owner.UserCtx, t)
	}()

	tt := []struct {
		name          string
		ctx           context.Context
		trustCenterID string
		expectedErr   string
	}{
		{
			name:          "owner can create a preview domain",
			ctx:           org.Owner.UserCtx,
			trustCenterID: org.TrustCenter.ID,
		},
		{
			name:          "support user can create a preview domain",
			ctx:           org.SupportCtx,
			trustCenterID: org.TrustCenter.ID,
		},
		{
			name:          "view only user can create a preview domain",
			ctx:           org.Member.UserCtx,
			trustCenterID: org.TrustCenter.ID,
			expectedErr:   th.NotAuthorizedErrorMsg,
		},

		{
			name:          "no access to another organization",
			ctx:           secondOrg.Owner.UserCtx,
			trustCenterID: org.TrustCenter.ID,
			expectedErr:   th.NotFoundErrorMsg,
		},
		{
			name:          "cannot create preview domain for non existent trustcenter",
			ctx:           org.Owner.UserCtx,
			trustCenterID: ulids.New().String(),
			expectedErr:   th.NotFoundErrorMsg,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert.NilError(t, suite.Client.DB.Job.TruncateRiverTables(tc.ctx))

			_, err := suite.Client.API.CreateTrustCenterPreviewDomain(tc.ctx, tc.trustCenterID)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				rivertest.RequireNotInserted(tc.ctx, t, riverpgxv5.New(suite.Client.DB.Job.GetPool()), &jobspec.CreatePreviewDomainArgs{}, nil)
				return
			}

			assert.NilError(t, err)

			rivertest.RequireManyInserted(tc.ctx, t, riverpgxv5.New(suite.Client.DB.Job.GetPool()), []rivertest.ExpectedJob{
				{
					Args: jobspec.CreatePreviewDomainArgs{
						TrustCenterID:            tc.trustCenterID,
						TrustCenterPreviewZoneID: testutils.TrustCenterPreviewZoneID,
						TrustCenterCnameTarget:   testutils.TrustCenterCnameTarget,
					},
				},
			})
		})
	}
}

func TestMutationSetupTrustCenterDomains(t *testing.T) {
	org := th.CreateFreshOrgWithTrustCenter(t)

	defer th.CleanupOrganizationDataWithContext(org.Owner.UserCtx, t)

	tt := []struct {
		name          string
		ctx           context.Context
		trustCenterID string
		expectedErr   string
	}{
		{
			name:          "owner can setup domains",
			ctx:           org.Owner.UserCtx,
			trustCenterID: org.TrustCenter.ID,
		},
		{
			name:          "trust center not found",
			ctx:           org.Owner.UserCtx,
			trustCenterID: ulids.New().String(),
			expectedErr:   th.NotFoundErrorMsg,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert.NilError(t, suite.Client.DB.Job.TruncateRiverTables(tc.ctx))

			resp, err := suite.Client.API.SetupTrustCenterDomains(tc.ctx, tc.trustCenterID)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				rivertest.RequireNotInserted(tc.ctx, t, riverpgxv5.New(suite.Client.DB.Job.GetPool()), &jobspec.CreatePreviewDomainArgs{}, nil)
				return
			}

			assert.NilError(t, err)
			assert.Check(t, resp.SetupTrustCenterDomains.Success)

			setting, err := suite.Client.API.GetTrustCenterSettingByID(org.Owner.UserCtx, org.TrustCenter.Edges.PreviewSetting.ID)
			assert.NilError(t, err)
			assert.Assert(t, setting.TrustCenterSetting.DisableDefaultDomain != nil)
			assert.Check(t, !lo.FromPtr(setting.TrustCenterSetting.DisableDefaultDomain))
			rivertest.RequireManyInserted(tc.ctx, t, riverpgxv5.New(suite.Client.DB.Job.GetPool()), []rivertest.ExpectedJob{
				{
					Args: jobspec.CreatePreviewDomainArgs{
						TrustCenterID:            tc.trustCenterID,
						TrustCenterPreviewZoneID: testutils.TrustCenterPreviewZoneID,
						TrustCenterCnameTarget:   testutils.TrustCenterCnameTarget,
					},
				},
			})
		})
	}
}

func TestMutationCreateTrustCenterDomain(t *testing.T) {
	t.Parallel()

	tcOrg := th.CreateFreshOrgWithTrustCenter(t)
	mappableDomain := (&th.MappableDomainBuilder{Client: suite.Client, Name: testutils.TrustCenterCnameTarget}).MustNew(tcOrg.Admin.UserCtx, t)

	t.Run("happy path, do not require TrustCenterID", func(t *testing.T) {
		domain := gofakeit.DomainName()
		resp, err := suite.Client.API.CreateTrustCenterDomain(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord: domain,
		})
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		assert.Check(t, is.Equal(domain, resp.CreateTrustCenterDomain.CustomDomain.CnameRecord))
		assert.Check(t, is.Equal(enums.CustomDomainTypeExternal, resp.CreateTrustCenterDomain.CustomDomain.DomainType))
		assert.Check(t, resp.CreateTrustCenterDomain.CustomDomain.TrustCenterID != nil)
		assert.Check(t, is.Equal(tcOrg.TrustCenter.ID, *resp.CreateTrustCenterDomain.CustomDomain.TrustCenterID))
		(&th.Cleanup[*generated.CustomDomainDeleteOne]{Client: suite.Client.DB.CustomDomain, ID: resp.CreateTrustCenterDomain.CustomDomain.ID}).MustDelete(tcOrg.Owner.UserCtx, t)
	})

	t.Run("normalizes cname record input", func(t *testing.T) {
		inputDomain := "https://Trust.Example.com/path"
		domainType := enums.CustomDomainTypePreview
		resp, err := suite.Client.API.CreateTrustCenterDomain(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord: inputDomain,
			DomainType:  &domainType,
		})
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal("trust.example.com", resp.CreateTrustCenterDomain.CustomDomain.CnameRecord))
		previewMappableDomain, err := suite.Client.DB.MappableDomain.Get(th.SetInternalContext(tcOrg.Owner.UserCtx, suite.Client.DB), resp.CreateTrustCenterDomain.CustomDomain.MappableDomainID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(testutils.TrustCenterCnameTarget, previewMappableDomain.Name))
		assert.Check(t, is.Equal(mappableDomain.ID, previewMappableDomain.ID))
		assert.Check(t, is.Equal(enums.CustomDomainTypePreview, resp.CreateTrustCenterDomain.CustomDomain.DomainType))
		(&th.Cleanup[*generated.CustomDomainDeleteOne]{Client: suite.Client.DB.CustomDomain, ID: resp.CreateTrustCenterDomain.CustomDomain.ID}).MustDelete(tcOrg.Owner.UserCtx, t)
	})

	t.Run("trust center not found", func(t *testing.T) {
		_, err := suite.Client.API.CreateTrustCenterDomain(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			TrustCenterID: "non-existent-id",
		})
		assert.ErrorContains(t, err, th.NotFoundErrorMsg)
	})

	t.Run("view only user cannot create domain", func(t *testing.T) {
		// Create a new user and trust center to avoid slug conflicts
		testUserForViewOnly := suite.UserBuilder(t.Context(), t)

		// Add viewOnlyUser to this new organization as a member (view-only)
		suite.AddUserToOrganization(testUserForViewOnly.UserCtx, t, &th.SharedViewOnlyUser, enums.RoleMember, tcOrg.OrganizationID)

		_, err := suite.Client.API.CreateTrustCenterDomain(th.SharedViewOnlyUser.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			TrustCenterID: tcOrg.TrustCenter.ID,
		})
		assert.ErrorContains(t, err, th.NotAuthorizedErrorMsg)

		th.CleanupOrganizationDataWithContext(testUserForViewOnly.UserCtx, t)
	})

	t.Run("user from different organization cannot access trust center", func(t *testing.T) {

		_, err := suite.Client.API.CreateTrustCenterDomain(th.SharedTestUser2.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			TrustCenterID: tcOrg.Owner.ID,
		})
		assert.ErrorContains(t, err, th.NotFoundErrorMsg)
	})

	t.Run("only one domain type can be added per trustcenter", func(t *testing.T) {
		freshOrg := th.CreateFreshOrgWithTrustCenter(t)

		resp, err := suite.Client.API.CreateTrustCenterDomain(freshOrg.Owner.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			TrustCenterID: freshOrg.TrustCenter.ID,
		})
		assert.NilError(t, err)
		assert.Check(t, is.Equal(enums.CustomDomainTypeExternal, resp.CreateTrustCenterDomain.CustomDomain.DomainType))

		previewDomainType := enums.CustomDomainTypePreview
		previewResp, err := suite.Client.API.CreateTrustCenterDomain(freshOrg.Owner.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			DomainType:    &previewDomainType,
			TrustCenterID: freshOrg.TrustCenter.ID,
		})
		assert.NilError(t, err)
		assert.Check(t, is.Equal(enums.CustomDomainTypePreview, previewResp.CreateTrustCenterDomain.CustomDomain.DomainType))

		_, err = suite.Client.API.CreateTrustCenterDomain(freshOrg.Owner.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			TrustCenterID: freshOrg.TrustCenter.ID,
		})
		assert.ErrorContains(t, err, "domain already exists for this trust center")

		_, err = suite.Client.API.CreateTrustCenterDomain(freshOrg.Owner.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			DomainType:    &previewDomainType,
			TrustCenterID: freshOrg.TrustCenter.ID,
		})
		assert.ErrorContains(t, err, "domain already exists for this trust center")

		th.CleanupOrganizationDataWithContext(freshOrg.Owner.UserCtx, t)
	})

	t.Run("trust center already has a domain", func(t *testing.T) {
		// Create trust center in testUser2's org to avoid slug conflicts
		testUserDomainExists := suite.UserBuilder(t.Context(), t)
		trustCenter4 := (&th.TrustCenterBuilder{Client: suite.Client}).MustNew(testUserDomainExists.UserCtx, t)

		// Create a custom domain and associate it with the trust center using the builder
		existingDomain := (&th.CustomDomainBuilder{Client: suite.Client, MappableDomainID: mappableDomain.ID}).MustNew(testUserDomainExists.UserCtx, t)

		// Update trust center to have the custom domain using proper context
		ctx := th.SetInternalContext(testUserDomainExists.UserCtx, suite.Client.DB)
		_, err := suite.Client.DB.TrustCenter.UpdateOneID(trustCenter4.ID).SetCustomDomainID(existingDomain.ID).Save(ctx)
		assert.NilError(t, err)

		_, err = suite.Client.API.CreateTrustCenterDomain(testUserDomainExists.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			TrustCenterID: trustCenter4.ID,
		})
		assert.ErrorContains(t, err, "domain already exists for this trust center")

		// th.Cleanup
		th.CleanupOrganizationDataWithContext(testUserDomainExists.UserCtx, t)
	})

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	(&th.Cleanup[*generated.MappableDomainDeleteOne]{Client: suite.Client.DB.MappableDomain, IDs: []string{mappableDomain.ID}}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)

}

func TestMutationCreateTrustCenterDomainMappableDomainNotExists(t *testing.T) {
	// Create a new user to avoid slug conflicts
	testUser := suite.UserBuilder(t.Context(), t)
	trustCenter := (&th.TrustCenterBuilder{Client: suite.Client}).MustNew(testUser.UserCtx, t)

	t.Run("mappable domain does not exist", func(t *testing.T) {
		_, err := suite.Client.API.CreateTrustCenterDomain(testUser.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			TrustCenterID: trustCenter.ID,
		})
		assert.ErrorContains(t, err, th.NotFoundErrorMsg)
	})

	th.CleanupOrganizationDataWithContext(testUser.UserCtx, t)
}
