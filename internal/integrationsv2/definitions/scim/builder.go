package scim

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrationsv2/definition"
	"github.com/theopenlane/core/internal/integrationsv2/types"
	"github.com/theopenlane/core/pkg/gala"
)

// operatorConfig holds operator-owned defaults that apply across all SCIM installations
type operatorConfig struct {
	DefaultProvisioningMode string `json:"defaultProvisioningMode,omitempty" jsonschema:"title=Default Provisioning Mode"`
	DefaultBasePath         string `json:"defaultBasePath,omitempty"         jsonschema:"title=Default Base Path"`
}

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label            string `json:"label,omitempty"            jsonschema:"title=Tenant Label"`
	TenantKey        string `json:"tenantKey,omitempty"        jsonschema:"title=Tenant Key"`
	MappingMode      string `json:"mappingMode,omitempty"      jsonschema:"title=Mapping Mode"`
	ProvisioningMode string `json:"provisioningMode,omitempty" jsonschema:"title=Provisioning Mode"`
}

// credential holds the inbound or outbound authentication material for one SCIM installation
type credential struct {
	BaseURL       string `json:"baseUrl,omitempty"       jsonschema:"title=SCIM Base URL"`
	Token         string `json:"token,omitempty"         jsonschema:"title=Bearer Token"`
	InboundSecret string `json:"inboundSecret,omitempty" jsonschema:"title=Inbound Secret"`
}

var (
	definitionSpec       = types.DefinitionSpec{
		ID:          "def_01K0SCIM000000000000000001",
		Slug:        "scim_directory_sync",
		Version:     "v1",
		Family:      "scim",
		DisplayName: "SCIM Directory Sync",
		Description: "Synchronize directory objects through SCIM",
		Category:    "directory",
		DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/scim/overview",
		Labels:      map[string]string{"protocol": "scim"},
		Active:      true,
		Visible:     true,
	}

	operatorConfigSchema = providerkit.SchemaFrom[operatorConfig]()
	userInputSchema      = providerkit.SchemaFrom[userInput]()
	credentialSchema     = providerkit.SchemaFrom[credential]()
)

// def implements definition.Assembler for the SCIM Directory Sync integration
type def struct{}

// Builder returns the SCIM definition builder
func Builder() definition.Builder {
	return definition.FromAssembler(&def{})
}

func (d *def) Spec() types.DefinitionSpec { return definitionSpec }

func (d *def) OperatorConfig() *types.OperatorConfigRegistration {
	return &types.OperatorConfigRegistration{Schema: operatorConfigSchema}
}

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
			Name:        "scim",
			Description: "SCIM API client",
			Build:       buildSCIMClient,
		},
	}
}

func (d *def) Operations() []types.OperationRegistration {
	return []types.OperationRegistration{
		{
			Name:        "directory.sync",
			Kind:        types.OperationKindSync,
			Description: "Synchronize directory state through SCIM",
			Topic:       gala.TopicName("integration.scim.directory.sync"),
			Client:      "scim",
			Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
			Handle:      runDirectorySyncOperation,
		},
	}
}

func (d *def) Mappings() []types.MappingRegistration {
	return []types.MappingRegistration{
		{
			Schema:  "directory_account",
			Variant: "default",
			Spec:    types.MappingOverride{Version: "v1"},
		},
		{
			Schema:  "directory_group",
			Variant: "default",
			Spec:    types.MappingOverride{Version: "v1"},
		},
	}
}

func (d *def) Webhooks() []types.WebhookRegistration {
	return []types.WebhookRegistration{
		{
			Name:    "directory.push",
			Verify:  verifyWebhook,
			Resolve: resolveWebhook,
			Handle:  handleWebhook,
		},
	}
}
