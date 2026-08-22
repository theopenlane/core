package graphapi_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/internal/testutils"
)

func TestMutationCreateTrustCenterDomain(t *testing.T) {
	t.Parallel()

	tcOrg := createFreshOrgWithTrustCenter(t)
	mappableDomain := (&MappableDomainBuilder{client: suite.client, Name: testutils.TrustCenterCnameTarget}).MustNew(tcOrg.admin.UserCtx, t)

	t.Run("happy path, do not require TrustCenterID", func(t *testing.T) {
		domain := gofakeit.DomainName()
		resp, err := suite.client.api.CreateTrustCenterDomain(tcOrg.owner.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord: domain,
		})
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		assert.Check(t, is.Equal(domain, resp.CreateTrustCenterDomain.CustomDomain.CnameRecord))
		assert.Check(t, is.Equal(enums.CustomDomainTypeExternal, resp.CreateTrustCenterDomain.CustomDomain.DomainType))
		assert.Check(t, resp.CreateTrustCenterDomain.CustomDomain.TrustCenterID != nil)
		assert.Check(t, is.Equal(tcOrg.trustCenter.ID, *resp.CreateTrustCenterDomain.CustomDomain.TrustCenterID))
		(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: resp.CreateTrustCenterDomain.CustomDomain.ID}).MustDelete(tcOrg.owner.UserCtx, t)
	})

	t.Run("normalizes cname record input", func(t *testing.T) {
		inputDomain := "https://Trust.Example.com/path"
		domainType := enums.CustomDomainTypePreview
		resp, err := suite.client.api.CreateTrustCenterDomain(tcOrg.owner.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord: inputDomain,
			DomainType:  &domainType,
		})
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal("trust.example.com", resp.CreateTrustCenterDomain.CustomDomain.CnameRecord))
		assert.Check(t, is.Equal(enums.CustomDomainTypePreview, resp.CreateTrustCenterDomain.CustomDomain.DomainType))
		(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: resp.CreateTrustCenterDomain.CustomDomain.ID}).MustDelete(tcOrg.owner.UserCtx, t)
	})

	t.Run("trust center not found", func(t *testing.T) {
		_, err := suite.client.api.CreateTrustCenterDomain(tcOrg.owner.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			TrustCenterID: "non-existent-id",
		})
		assert.ErrorContains(t, err, notFoundErrorMsg)
	})

	t.Run("view only user cannot create domain", func(t *testing.T) {
		// Create a new user and trust center to avoid slug conflicts
		testUserForViewOnly := suite.userBuilder(t.Context(), t)

		// Add viewOnlyUser to this new organization as a member (view-only)
		suite.addUserToOrganization(testUserForViewOnly.UserCtx, t, &sharedViewOnlyUser, enums.RoleMember, tcOrg.organizationID)

		_, err := suite.client.api.CreateTrustCenterDomain(sharedViewOnlyUser.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			TrustCenterID: tcOrg.trustCenter.ID,
		})
		assert.ErrorContains(t, err, notAuthorizedErrorMsg)

		cleanupOrganizationDataWithContext(testUserForViewOnly.UserCtx, t)
	})

	t.Run("user from different organization cannot access trust center", func(t *testing.T) {

		_, err := suite.client.api.CreateTrustCenterDomain(sharedTestUser2.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			TrustCenterID: tcOrg.owner.ID,
		})
		assert.ErrorContains(t, err, notFoundErrorMsg)
	})

	t.Run("only one domain type can be added per trustcenter", func(t *testing.T) {
		freshOrg := createFreshOrgWithTrustCenter(t)

		resp, err := suite.client.api.CreateTrustCenterDomain(freshOrg.owner.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			TrustCenterID: freshOrg.trustCenter.ID,
		})
		assert.NilError(t, err)
		assert.Check(t, is.Equal(enums.CustomDomainTypeExternal, resp.CreateTrustCenterDomain.CustomDomain.DomainType))

		previewDomainType := enums.CustomDomainTypePreview
		previewResp, err := suite.client.api.CreateTrustCenterDomain(freshOrg.owner.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			DomainType:    &previewDomainType,
			TrustCenterID: freshOrg.trustCenter.ID,
		})
		assert.NilError(t, err)
		assert.Check(t, is.Equal(enums.CustomDomainTypePreview, previewResp.CreateTrustCenterDomain.CustomDomain.DomainType))

		_, err = suite.client.api.CreateTrustCenterDomain(freshOrg.owner.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			TrustCenterID: freshOrg.trustCenter.ID,
		})
		assert.ErrorContains(t, err, "domain already exists for this trust center")

		_, err = suite.client.api.CreateTrustCenterDomain(freshOrg.owner.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			DomainType:    &previewDomainType,
			TrustCenterID: freshOrg.trustCenter.ID,
		})
		assert.ErrorContains(t, err, "domain already exists for this trust center")

		cleanupOrganizationDataWithContext(freshOrg.owner.UserCtx, t)
	})

	t.Run("trust center already has a domain", func(t *testing.T) {
		// Create trust center in testUser2's org to avoid slug conflicts
		testUserDomainExists := suite.userBuilder(t.Context(), t)
		trustCenter4 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUserDomainExists.UserCtx, t)

		// Create a custom domain and associate it with the trust center using the builder
		existingDomain := (&CustomDomainBuilder{client: suite.client, MappableDomainID: mappableDomain.ID}).MustNew(testUserDomainExists.UserCtx, t)

		// Update trust center to have the custom domain using proper context
		ctx := setContext(testUserDomainExists.UserCtx, suite.client.db)
		_, err := suite.client.db.TrustCenter.UpdateOneID(trustCenter4.ID).SetCustomDomainID(existingDomain.ID).Save(ctx)
		assert.NilError(t, err)

		_, err = suite.client.api.CreateTrustCenterDomain(testUserDomainExists.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			TrustCenterID: trustCenter4.ID,
		})
		assert.ErrorContains(t, err, "domain already exists for this trust center")

		// Cleanup
		cleanupOrganizationDataWithContext(testUserDomainExists.UserCtx, t)
	})

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, IDs: []string{mappableDomain.ID}}).MustDelete(sharedSystemAdminUser.UserCtx, t)

}

func TestMutationCreateTrustCenterDomainMappableDomainNotExists(t *testing.T) {
	t.Parallel()
	// Create a new user to avoid slug conflicts
	testUser := suite.userBuilder(t.Context(), t)
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	t.Run("mappable domain does not exist", func(t *testing.T) {
		_, err := suite.client.api.CreateTrustCenterDomain(testUser.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			TrustCenterID: trustCenter.ID,
		})
		assert.ErrorContains(t, err, notFoundErrorMsg)
	})

	cleanupOrganizationDataWithContext(testUser.UserCtx, t)
}
