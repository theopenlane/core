package oidcgeneric

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label string `json:"label,omitempty" jsonschema:"title=Installation Label"`
}

// def holds operator config for the Generic OIDC integration
type def struct {
	cfg Config
}

// Builder returns the Generic OIDC definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	d := &def{cfg: cfg}
	return definition.BuilderFunc(func(_ context.Context) (types.Definition, error) {
		return types.Definition{
			Spec: types.DefinitionSpec{
				ID:          "def_01K0OIDCGEN00000000000000001",
				Slug:        "oidc_generic",
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
				Start:    d.startInstallAuth,
				Complete: d.completeInstallAuth,
			},
			Clients: []types.ClientRegistration{
				{
					Name:        "api",
					Description: "OIDC userinfo HTTP client",
					Build:       buildOIDCClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        "health.default",
					Kind:        types.OperationKindHealth,
					Description: "Call the configured userinfo endpoint to validate the OIDC token",
					Topic:       gala.TopicName("integration.oidc_generic.health.default"),
					Client:      "api",
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
				{
					Name:        "claims.inspect",
					Kind:        types.OperationKindCollect,
					Description: "Expose stored ID token claims for downstream checks",
					Topic:       gala.TopicName("integration.oidc_generic.claims.inspect"),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runClaimsInspectOperation,
				},
			},
		}, nil
	})
}
