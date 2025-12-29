package graphapi_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/testutils"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestMutationCreateTrustCenterDomain(t *testing.T) {
	testUser := suite.userBuilder(t.Context(), t)
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	mappableDomain := (&MappableDomainBuilder{client: suite.client, Name: testutils.TrustCenterCnameTarget}).MustNew(systemAdminUser.UserCtx, t)

	t.Run("happy path, do not require TrustCenterID", func(t *testing.T) {
		domain := gofakeit.DomainName()
		resp, err := suite.client.api.CreateTrustCenterDomain(testUser.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord: domain,
		})
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		assert.Check(t, is.Equal(domain, resp.CreateTrustCenterDomain.CustomDomain.CnameRecord))
		(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: resp.CreateTrustCenterDomain.CustomDomain.ID}).MustDelete(testUser.UserCtx, t)
	})

	t.Run("trust center not found", func(t *testing.T) {
		_, err := suite.client.api.CreateTrustCenterDomain(testUser.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			TrustCenterID: "non-existent-id",
		})
		assert.ErrorContains(t, err, notFoundErrorMsg)
	})

	t.Run("view only user cannot create domain", func(t *testing.T) {
		// Create a new user and trust center to avoid slug conflicts
		testUserForViewOnly := suite.userBuilder(t.Context(), t)

		// Add viewOnlyUser to this new organization as a member (view-only)
		suite.addUserToOrganization(testUserForViewOnly.UserCtx, t, &viewOnlyUser, enums.RoleMember, testUser.OrganizationID)

		_, err := suite.client.api.CreateTrustCenterDomain(viewOnlyUser.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			TrustCenterID: trustCenter.ID,
		})
		assert.ErrorContains(t, err, notAuthorizedErrorMsg)
	})

	t.Run("user from different organization cannot access trust center", func(t *testing.T) {

		_, err := suite.client.api.CreateTrustCenterDomain(testUser2.UserCtx, testclient.CreateTrustCenterDomainInput{
			CnameRecord:   gofakeit.DomainName(),
			TrustCenterID: trustCenter.ID,
		})
		assert.ErrorContains(t, err, notFoundErrorMsg)
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
		(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: existingDomain.ID}).MustDelete(testUserDomainExists.UserCtx, t)
		(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter4.ID}).MustDelete(testUserDomainExists.UserCtx, t)
	})

	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: mappableDomain.ID}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser.UserCtx, t)
}

func TestMutationCreateTrustCenterDomainMappableDomainNotExists(t *testing.T) {
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
}
