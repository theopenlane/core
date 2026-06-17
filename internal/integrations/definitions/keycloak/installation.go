package keycloak

import (
	"context"

	gocloak "github.com/Nerzal/gocloak/v13"
	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// resolveInstallationMetadata derives Keycloak realm metadata from the persisted credential
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error resolving credentials")

		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if cred.BaseURL == "" || cred.Realm == "" || cred.ClientID == "" || cred.ClientSecret == "" {
		return InstallationMetadata{}, false, nil
	}

	client, err := Client{}.Build(ctx, types.ClientBuildRequest{Credentials: req.Credentials})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error building keycloak client")

		return InstallationMetadata{}, false, err
	}

	gc := client.(*gocloak.GoCloak)

	token, err := gc.LoginClient(ctx, cred.ClientID, cred.ClientSecret, cred.Realm)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error logging in keycloak client")

		return InstallationMetadata{}, false, ErrInstallationResolveFailed
	}

	realm, err := gc.GetRealm(ctx, token.AccessToken, cred.Realm)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error fetching keycloak realm")

		return InstallationMetadata{}, false, ErrInstallationResolveFailed
	}

	return InstallationMetadata{
		RealmID:         lo.FromPtr(realm.ID),
		RealmName:       lo.FromPtr(realm.Realm),
		DisplayName:     lo.FromPtr(realm.DisplayName),
		KeycloakVersion: lo.FromPtr(realm.KeycloakVersion),
	}, true, nil
}
