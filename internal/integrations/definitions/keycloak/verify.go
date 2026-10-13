package keycloak

import (
	"context"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// verify resolves the Keycloak realm identity the credential is scoped to
func verify(ctx context.Context, _ types.ConnectionRequest[CredentialSchema], c Client) (InstallationMetadata, error) {
	token, err := c.ClientToken(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error logging in keycloak client")

		return InstallationMetadata{}, ErrInstallationResolveFailed
	}

	realm, err := c.GetRealm(ctx, token.AccessToken, c.Realm)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error fetching keycloak realm")

		return InstallationMetadata{}, ErrInstallationResolveFailed
	}

	return InstallationMetadata{
		RealmID:         lo.FromPtr(realm.ID),
		RealmName:       lo.FromPtr(realm.Realm),
		DisplayName:     lo.FromPtr(realm.DisplayName),
		KeycloakVersion: lo.FromPtr(realm.KeycloakVersion),
	}, nil
}
