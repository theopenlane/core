package hooks_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/iam/auth"
)

func (suite *HookTestSuite) TestValidateIdentityProviderConfig() {
	t := suite.T()

	user := suite.seedUser()
	ctx := auth.NewTestContextWithOrgID(user.ID, user.Edges.OrgMemberships[0].ID)
	ctx = generated.NewContext(ctx, suite.client)
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	t.Run("create missing fields", func(t *testing.T) {
		m := suite.client.OrganizationSetting.Create().
			SetIdentityProviderLoginEnforced(true).
			Mutation()

		err := hooks.ValidateIdentityProviderConfig(ctx, m)
		require.Error(t, err)
	})

	t.Run("create with fields", func(t *testing.T) {
		m := suite.client.OrganizationSetting.Create().
			SetIdentityProviderLoginEnforced(true).
			SetIdentityProvider(enums.SSOProviderOkta).
			SetIdentityProviderClientID("id").
			SetIdentityProviderClientSecret("secret").
			SetOidcDiscoveryEndpoint("https://example.com").
			Mutation()

		err := hooks.ValidateIdentityProviderConfig(ctx, m)
		require.Error(t, err)
		require.Equal(t, err, hooks.ErrSSONotEnforceable)
	})

	t.Run("update with existing fields", func(t *testing.T) {
		setting, err := suite.client.OrganizationSetting.Create().
			SetIdentityProvider(enums.SSOProviderOkta).
			SetIdentityProviderClientID("id").
			SetIdentityProviderClientSecret("secret").
			SetOidcDiscoveryEndpoint("https://example.com").
			Save(ctx)
		require.NoError(t, err)

		m := suite.client.OrganizationSetting.UpdateOneID(setting.ID).
			SetIdentityProviderLoginEnforced(true).Mutation()

		err = hooks.ValidateIdentityProviderConfig(ctx, m)
		require.Error(t, err)
		require.Equal(t, err, hooks.ErrSSONotEnforceable)
	})

	t.Run("update with missing fields", func(t *testing.T) {
		setting, err := suite.client.OrganizationSetting.Create().Save(ctx)
		require.NoError(t, err)

		m := suite.client.OrganizationSetting.UpdateOneID(setting.ID).
			SetIdentityProviderLoginEnforced(true).Mutation()

		err = hooks.ValidateIdentityProviderConfig(ctx, m)
		require.Error(t, err)
	})

	t.Run("update with tested connection", func(t *testing.T) {
		setting, err := suite.client.OrganizationSetting.Create().
			SetIdentityProvider(enums.SSOProviderOkta).
			SetIdentityProviderClientID("id").
			SetIdentityProviderClientSecret("secret").
			SetOidcDiscoveryEndpoint("https://example.com").
			Save(ctx)
		require.NoError(t, err)

		// verify the hook makes sure auth tested is false on idp update of any value
		require.False(t, setting.IdentityProviderAuthTested)

		testedSetting, err := suite.client.OrganizationSetting.UpdateOneID(setting.ID).
			SetIdentityProviderAuthTested(true).
			Save(ctx)
		require.NoError(t, err)
		require.True(t, testedSetting.IdentityProviderAuthTested)

		// enforce sso login for all users now that we have
		// enforced it earlier
		m := suite.client.OrganizationSetting.UpdateOneID(setting.ID).
			SetIdentityProviderLoginEnforced(true).Mutation()

		err = hooks.ValidateIdentityProviderConfig(ctx, m)
		require.NoError(t, err)
	})
}
