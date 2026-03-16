package oidcgeneric

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck identifies the default health check operation
type HealthCheck struct{}

// ClaimsInspect identifies the claims inspection operation
type ClaimsInspect struct{}

var (
	DefinitionID           = types.NewDefinitionRef("def_01K0OIDCGEN00000000000000001")
	HealthDefaultOperation = types.NewOperationRef[HealthCheck]("health.default")
	ClaimsInspectOperation = types.NewOperationRef[ClaimsInspect]("claims.inspect")
)

const Slug = "oidc_generic"

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label string `json:"label,omitempty" jsonschema:"title=Installation Label"`
}

// Builder returns the Generic OIDC definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.Builder(func(_ context.Context) (types.Definition, error) {
		clientRef := types.NewClientRef[any]()

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Version:     "v1",
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
				Schema: providerkit.SchemaFrom[userInput](),
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
					Ref:         clientRef.ID(),
					Description: "OIDC userinfo HTTP client",
					Build:       buildOIDCClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call the configured userinfo endpoint to validate the OIDC token",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   clientRef.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
				{
					Name:        ClaimsInspectOperation.Name(),
					Description: "Expose stored ID token claims for downstream checks",
					Topic:       ClaimsInspectOperation.Topic(Slug),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runClaimsInspectOperation,
				},
			},
		}, nil
	})
}
