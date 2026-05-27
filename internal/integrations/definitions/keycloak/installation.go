package keycloak

import (
	"context"

	gocloak "github.com/Nerzal/gocloak/v13"
	"github.com/theopenlane/core/internal/integrations/types"
)

// resolveInstallationMetadata derives Keycloak realm metadata from the persisted credential
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if cred.BaseURL == "" || cred.Realm == "" || cred.ClientID == "" || cred.ClientSecret == "" {
		return InstallationMetadata{}, false, nil
	}

	client, err := Client{}.Build(ctx, types.ClientBuildRequest{Credentials: req.Credentials})
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	gc := client.(*gocloak.GoCloak)

	token, err := gc.LoginClient(ctx, cred.ClientID, cred.ClientSecret, cred.Realm)
	if err != nil {
		return InstallationMetadata{}, false, ErrInstallationResolveFailed
	}

	realm, err := gc.GetRealm(ctx, token.AccessToken, cred.Realm)
	if err != nil {
		return InstallationMetadata{}, false, ErrInstallationResolveFailed
	}

	return InstallationMetadata{
		RealmID:         derefString(realm.ID),
		RealmName:       derefString(realm.Realm),
		DisplayName:     derefString(realm.DisplayName),
		KeycloakVersion: derefString(realm.KeycloakVersion),
	}, true, nil
}