package hooks_test

import (
	"testing"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/iam/auth"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestValidateIdentityProviderConfig(t *testing.T) {
	s := suite.setupTest(t)

	user := s.seedUser()
	ctx := auth.NewTestContextWithOrgID(user.ID, user.Edges.OrgMemberships[0].ID)
	ctx = generated.NewContext(ctx, s.client)
	ctx = auth.WithInternalOperationContext(ctx)

	t.Run("create missing fields", func(t *testing.T) {
		m := s.client.OrganizationSetting.Create().
			SetIdentityProviderLoginEnforced(true).
			Mutation()

		err := hooks.ValidateIdentityProviderConfig(ctx, m)
		assert.Assert(t, err != nil)
	})

	t.Run("create with fields", func(t *testing.T) {
		m := s.client.OrganizationSetting.Create().
			SetIdentityProviderLoginEnforced(true).
			SetIdentityProvider(enums.SSOProviderOkta).
			SetIdentityProviderClientID("id").
			SetIdentityProviderClientSecret("secret").
			SetOidcDiscoveryEndpoint("https://example.com").
			Mutation()

		err := hooks.ValidateIdentityProviderConfig(ctx, m)
		assert.Assert(t, err != nil)
		assert.Assert(t, is.Equal(err, hooks.ErrSSONotEnforceable))
	})

	t.Run("update with existing fields", func(t *testing.T) {
		setting, err := s.client.OrganizationSetting.Create().
			SetIdentityProvider(enums.SSOProviderOkta).
			SetIdentityProviderClientID("id").
			SetIdentityProviderClientSecret("secret").
			SetOidcDiscoveryEndpoint("https://example.com").
			Save(ctx)
		assert.NilError(t, err)

		m := s.client.OrganizationSetting.UpdateOneID(setting.ID).
			SetIdentityProviderLoginEnforced(true).Mutation()

		err = hooks.ValidateIdentityProviderConfig(ctx, m)
		assert.Assert(t, err != nil)
		assert.Assert(t, is.Equal(err, hooks.ErrSSONotEnforceable))
	})

	t.Run("update with missing fields", func(t *testing.T) {
		setting, err := s.client.OrganizationSetting.Create().Save(ctx)
		assert.NilError(t, err)

		m := s.client.OrganizationSetting.UpdateOneID(setting.ID).
			SetIdentityProviderLoginEnforced(true).Mutation()

		err = hooks.ValidateIdentityProviderConfig(ctx, m)
		assert.Assert(t, err != nil)
	})

	t.Run("update with tested connection", func(t *testing.T) {
		setting, err := s.client.OrganizationSetting.Create().
			SetIdentityProvider(enums.SSOProviderOkta).
			SetIdentityProviderClientID("id").
			SetIdentityProviderClientSecret("secret").
			SetOidcDiscoveryEndpoint("https://example.com").
			Save(ctx)
		assert.NilError(t, err)

		// verify the hook makes sure auth tested is false on idp update of any value
		assert.Assert(t, !setting.IdentityProviderAuthTested)

		testedSetting, err := s.client.OrganizationSetting.UpdateOneID(setting.ID).
			SetIdentityProviderAuthTested(true).
			Save(ctx)
		assert.NilError(t, err)
		assert.Assert(t, testedSetting.IdentityProviderAuthTested)

		// enforce sso login for all users now that we have
		// enforced it earlier
		m := s.client.OrganizationSetting.UpdateOneID(setting.ID).
			SetIdentityProviderLoginEnforced(true).Mutation()

		err = hooks.ValidateIdentityProviderConfig(ctx, m)
		assert.NilError(t, err)
	})
}
