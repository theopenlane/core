package oidcgeneric

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrationsv2/definition"
	"github.com/theopenlane/core/internal/integrationsv2/types"
	"github.com/theopenlane/core/pkg/gala"
)

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label string `json:"label,omitempty" jsonschema:"title=Installation Label"`
}

var (
	definitionSpec  = types.DefinitionSpec{
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
	}

	configSchema    = providerkit.SchemaFrom[Config]()
	userInputSchema = providerkit.SchemaFrom[userInput]()
)

// def implements definition.Assembler for the Generic OIDC integration
type def struct {
	cfg Config
}

// Builder returns the Generic OIDC definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.FromAssembler(&def{cfg: cfg})
}

func (d *def) Spec() types.DefinitionSpec { return definitionSpec }

func (d *def) OperatorConfig() *types.OperatorConfigRegistration {
	return &types.OperatorConfigRegistration{Schema: configSchema}
}

func (d *def) UserInput() *types.UserInputRegistration {
	return &types.UserInputRegistration{Schema: userInputSchema}
}

func (d *def) Credentials() *types.CredentialRegistration { return nil }

func (d *def) Auth() *types.AuthRegistration {
	return &types.AuthRegistration{
		Start:    startInstallAuth,
		Complete: completeInstallAuth,
	}
}

func (d *def) Clients() []types.ClientRegistration {
	return []types.ClientRegistration{
		{
			Name:        "api",
			Description: "OIDC userinfo HTTP client",
			Build:       buildOIDCClient,
		},
	}
}

func (d *def) Operations() []types.OperationRegistration {
	return []types.OperationRegistration{
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
	}
}

func (d *def) Mappings() []types.MappingRegistration { return nil }
func (d *def) Webhooks() []types.WebhookRegistration { return nil }
