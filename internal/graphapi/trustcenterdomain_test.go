package graphapi_test

import (
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/brianvoe/gofakeit/v7"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/core/v2/internal/testutils"
)

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
		ctx := th.SetContext(testUserDomainExists.UserCtx, suite.Client.DB)
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
