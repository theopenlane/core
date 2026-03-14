package okta

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrationsv2/definition"
	"github.com/theopenlane/core/internal/integrationsv2/types"
	"github.com/theopenlane/core/pkg/gala"
)

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label  string `json:"label,omitempty"  jsonschema:"title=Installation Label"`
	OrgURL string `json:"orgUrl,omitempty" jsonschema:"title=Org URL"`
}

// credential holds the Okta tenant credentials for one installation
type credential struct {
	OrgURL   string `json:"orgUrl"   jsonschema:"required,title=Org URL"`
	APIToken string `json:"apiToken" jsonschema:"required,title=API Token"`
}

var (
	definitionSpec = types.DefinitionSpec{
		ID:          "def_01K0OKTA0000000000000000001",
		Slug:        "okta",
		Version:     "v1",
		Family:      "okta",
		DisplayName: "Okta",
		Description: "Collect Okta tenant and sign-on policy metadata for identity posture and access governance analysis.",
		Category:    "sso",
		DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/okta/overview",
		Labels:      map[string]string{"vendor": "okta", "product": "identity"},
		Active:      false,
		Visible:     true,
	}

	userInputSchema  = providerkit.SchemaFrom[userInput]()
	credentialSchema = providerkit.SchemaFrom[credential]()
)

// def implements definition.Assembler for the Okta integration
type def struct{}

// Builder returns the Okta definition builder
func Builder() definition.Builder {
	return definition.FromAssembler(&def{})
}

func (d *def) Spec() types.DefinitionSpec { return definitionSpec }

func (d *def) OperatorConfig() *types.OperatorConfigRegistration { return nil }

func (d *def) UserInput() *types.UserInputRegistration {
	return &types.UserInputRegistration{Schema: userInputSchema}
}

func (d *def) Credentials() *types.CredentialRegistration {
	return &types.CredentialRegistration{
		Schema:  credentialSchema,
		Persist: types.CredentialPersistModeKeystore,
	}
}

func (d *def) Auth() *types.AuthRegistration { return nil }

func (d *def) Clients() []types.ClientRegistration {
	return []types.ClientRegistration{
		{
			Name:        "api",
			Description: "Okta API client",
			Build:       buildOktaClient,
		},
	}
}

func (d *def) Operations() []types.OperationRegistration {
	return []types.OperationRegistration{
		{
			Name:        "health.default",
			Kind:        types.OperationKindHealth,
			Description: "Call Okta user API to verify API token",
			Topic:       gala.TopicName("integration.okta.health.default"),
			Client:      "api",
			Policy:      types.ExecutionPolicy{Idempotent: true},
			Handle:      runHealthOperation,
		},
		{
			Name:        "policies.collect",
			Kind:        types.OperationKindCollect,
			Description: "Collect sign-on policy metadata for posture analysis",
			Topic:       gala.TopicName("integration.okta.policies.collect"),
			Client:      "api",
			Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
			Handle:      runPoliciesCollectOperation,
		},
	}
}

func (d *def) Mappings() []types.MappingRegistration  { return nil }
func (d *def) Webhooks() []types.WebhookRegistration  { return nil }
