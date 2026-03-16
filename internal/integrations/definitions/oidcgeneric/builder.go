package oidcgeneric

import (
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Generic OIDC definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Family:      "oidc",
				DisplayName: "Generic OIDC",
				Description: "Connect Openlane to standards-based OpenID Connect identity providers for federated authentication and claim inspection.",
				Category:    "sso",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/oidc_generic/overview",
				Labels:      map[string]string{"protocol": "oidc"},
				Active:      true,
				Visible:     false,
			},
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: providerkit.SchemaFrom[Config](),
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			Auth: &types.AuthRegistration{
				StartPath:    "/v1/integrations/oauth/start",
				CallbackPath: "/v1/integrations/oauth/callback",
				OAuth: &types.OAuthPublicConfig{
					ClientID:    cfg.ClientID,
					RedirectURI: cfg.RedirectURL,
					Scopes:      oidcGenericScopes,
				},
				ClientSecret: cfg.ClientSecret,
				DiscoveryURL: cfg.DiscoveryURL,
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         OIDCClient.ID(),
					Description: "OIDC userinfo HTTP client",
					Build:       Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call the configured userinfo endpoint to validate the OIDC token",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   OIDCClient.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      HealthCheck{}.Handle(Client{}),
				},
				{
					Name:        ClaimsInspectOperation.Name(),
					Description: "Expose stored ID token claims for downstream checks",
					Topic:       ClaimsInspectOperation.Topic(Slug),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      ClaimsInspect{}.Handle(),
				},
			},
		}, nil
	})
}
