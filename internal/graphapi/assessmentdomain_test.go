package graphapi_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/theopenlane/core/common/enums"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/testutils"
)

func TestMutationCreateAssessmentDomain(t *testing.T) {
	t.Parallel()

	org := suite.SeedFreshMinimalOrgUsers(t, false)

	t.Run("mappable domain does not exist", func(t *testing.T) {
		_, err := suite.Client.API.CreateAssessmentDomain(org.Owner.UserCtx, testclient.CreateAssessmentDomainInput{
			CnameRecord: gofakeit.DomainName(),
		})
		assert.ErrorContains(t, err, th.NotFoundErrorMsg)
	})

	mappableDomain := (&th.MappableDomainBuilder{Client: suite.Client, Name: testutils.AssessmentCnameTarget}).MustNew(org.Admin.UserCtx, t)

	t.Run("happy path, can create an assessment domain", func(t *testing.T) {
		domain := gofakeit.DomainName()

		resp, err := suite.Client.API.CreateAssessmentDomain(org.Owner.UserCtx, testclient.CreateAssessmentDomainInput{
			CnameRecord: domain,
		})

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal(domain, resp.CreateAssessmentDomain.CustomDomain.CnameRecord))

		customDomain, err := suite.Client.API.GetCustomDomainByID(org.Owner.UserCtx, resp.CreateAssessmentDomain.CustomDomain.ID)
		assert.NilError(t, err)

		assert.Check(t, is.Equal(enums.DomainPurposeAssessment, *customDomain.CustomDomain.Purpose))
		assert.Check(t, is.Equal(enums.CustomDomainTypeExternal, customDomain.CustomDomain.DomainType))
		assert.Check(t, is.Equal(mappableDomain.ID, customDomain.CustomDomain.MappableDomainID))
	})

	t.Run("member does not have permission to create an assessment domain", func(t *testing.T) {
		newOrg := suite.SeedFreshMinimalOrgUsers(t, false)
		_, err := suite.Client.API.CreateAssessmentDomain(newOrg.Member.UserCtx, testclient.CreateAssessmentDomainInput{
			CnameRecord: gofakeit.DomainName(),
		})
		assert.ErrorContains(t, err, th.NotAuthorizedErrorMsg)

		th.CleanupOrganizationDataWithContext(newOrg.Owner.UserCtx, t)
	})

	t.Run("only one assessment domain can be added in an organization", func(t *testing.T) {
		newOrg := suite.SeedFreshMinimalOrgUsers(t, false)

		resp, err := suite.Client.API.CreateAssessmentDomain(newOrg.Owner.UserCtx, testclient.CreateAssessmentDomainInput{
			CnameRecord: gofakeit.DomainName(),
		})
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		_, err = suite.Client.API.CreateAssessmentDomain(newOrg.Owner.UserCtx, testclient.CreateAssessmentDomainInput{
			CnameRecord: gofakeit.DomainName(),
		})
		assert.ErrorContains(t, err, "domain already exists for this organization")

		th.CleanupOrganizationDataWithContext(newOrg.Owner.UserCtx, t)
	})

	th.CleanupOrganizationDataWithContext(org.Owner.UserCtx, t)
	(&th.Cleanup[*generated.MappableDomainDeleteOne]{Client: suite.Client.DB.MappableDomain, IDs: []string{mappableDomain.ID}}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
}
