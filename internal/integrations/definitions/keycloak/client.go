package keycloak

import (
	"context"
	"time"

	gocloak "github.com/Nerzal/gocloak/v13"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

const (
	// keycloakRequestTimeout is the per-request timeout for Keycloak API calls
	keycloakRequestTimeout = 30 * time.Second
	// keycloakDefaultPageSize is the number of records requested per Keycloak API page
	keycloakDefaultPageSize = 100
)

// buildClient constructs the Keycloak API client for one installation
func buildClient(_ context.Context, req types.ConnectionRequest[CredentialSchema]) (Client, error) {
	cred := req.Credential

	switch {
	case cred.BaseURL == "":
		return Client{}, ErrBaseURLMissing
	case cred.Realm == "":
		return Client{}, ErrRealmMissing
	case cred.ClientID == "":
		return Client{}, ErrClientIDMissing
	case cred.ClientSecret == "":
		return Client{}, ErrClientSecretMissing
	}

	gc := gocloak.NewClient(cred.BaseURL)
	gc.RestyClient().SetTimeout(keycloakRequestTimeout)

	return Client{GoCloak: gc, Realm: cred.Realm, ClientID: cred.ClientID, ClientSecret: cred.ClientSecret}, nil
}
