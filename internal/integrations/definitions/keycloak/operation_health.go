package keycloak

import (
	"context"
	"encoding/json"

	gocloak "github.com/Nerzal/gocloak/v13"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck holds the result of a Keycloak health check
type HealthCheck struct {
	// Realm is the Keycloak realm that was checked
	Realm string `json:"realm"`
	// RealmID is the stable UUID of the Keycloak realm
	RealmID string `json:"realmId"`
	// KeycloakVersion is the version of the Keycloak instance
	KeycloakVersion string `json:"keycloakVersion"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClientRequest(keycloakClient, func(ctx context.Context, req types.OperationRequest, gc *gocloak.GoCloak) (json.RawMessage, error) {
		return h.Run(ctx, gc, req)
	})
}

// Run executes the Keycloak health check
func (HealthCheck) Run(ctx context.Context, gc *gocloak.GoCloak, req types.OperationRequest) (json.RawMessage, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		return nil, ErrHealthCheckFailed
	}

	token, err := gc.LoginClient(ctx, cred.ClientID, cred.ClientSecret, cred.Realm)
	if err != nil {
		return nil, ErrHealthCheckFailed
	}

	realm, err := gc.GetRealm(ctx, token.AccessToken, cred.Realm)
	if err != nil {
		return nil, ErrHealthCheckFailed
	}

	return providerkit.EncodeResult(HealthCheck{
		Realm:           lo.FromPtr(realm.Realm),
		RealmID:         lo.FromPtr(realm.ID),
		KeycloakVersion: lo.FromPtr(realm.KeycloakVersion),
	}, ErrResultEncode)
}
