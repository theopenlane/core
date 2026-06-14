package keycloak

import (
	"context"
	"time"

	gocloak "github.com/Nerzal/gocloak/v13"

	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// keycloakRequestTimeout is the per-request timeout for Keycloak API calls
	keycloakRequestTimeout = 30 * time.Second
	// keycloakDefaultPageSize is the number of records requested per Keycloak API page
	keycloakDefaultPageSize = 100
)

// Client builds Keycloak API clients for one installation
type Client struct{}

// Build constructs the Keycloak API client for one installation
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		return nil, err
	}

	if cred.BaseURL == "" {
		return nil, ErrBaseURLMissing
	}

	if cred.Realm == "" {
		return nil, ErrRealmMissing
	}

	if cred.ClientID == "" {
		return nil, ErrClientIDMissing
	}

	if cred.ClientSecret == "" {
		return nil, ErrClientSecretMissing
	}

	gc := gocloak.NewClient(cred.BaseURL)
	gc.RestyClient().SetTimeout(keycloakRequestTimeout)

	return gc, nil
}

// resolveCredential extracts the CredentialSchema from the provided credential bindings
func resolveCredential(bindings types.CredentialBindings) (CredentialSchema, error) {
	cred, ok, err := keycloakCredential.Resolve(bindings)
	if err != nil {
		return CredentialSchema{}, ErrCredentialDecode
	}

	if !ok {
		return CredentialSchema{}, ErrCredentialDecode
	}

	return cred, nil
}

// derefString safely dereferences a string pointer returning empty string if nil
func derefString(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}
