package keycloak

import (
	"context"
	"time"

	gocloak "github.com/Nerzal/gocloak/v13"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

const (
	// keycloakRequestTimeout is the per-request timeout for Keycloak API calls
	keycloakRequestTimeout = 30 * time.Second
	// keycloakDefaultPageSize is the number of records requested per Keycloak API page
	keycloakDefaultPageSize = 100
)

// clientBuilder returns the Keycloak client constructor, allowPrivate permits non-public base urls for local development
func clientBuilder(allowPrivate bool) func(context.Context, types.ConnectionRequest[CredentialSchema]) (Client, error) {
	return func(_ context.Context, req types.ConnectionRequest[CredentialSchema]) (Client, error) {
		return buildClient(req, allowPrivate)
	}
}

// buildClient constructs the Keycloak API client for one installation
func buildClient(req types.ConnectionRequest[CredentialSchema], allowPrivate bool) (Client, error) {
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

	httpClient, err := urlx.NewCustomerHTTPClient(cred.BaseURL, allowPrivate)
	if err != nil {
		return Client{}, err
	}

	gc := gocloak.NewClient(cred.BaseURL)
	gc.RestyClient().SetTransport(httpClient.Transport)
	gc.RestyClient().SetTimeout(keycloakRequestTimeout)

	return Client{GoCloak: gc, Realm: cred.Realm, ClientID: cred.ClientID, ClientSecret: cred.ClientSecret}, nil
}
