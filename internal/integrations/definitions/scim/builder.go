package scim

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// DirectorySync identifies the directory synchronization operation
type DirectorySync struct{}

var (
	DefinitionID           = types.NewDefinitionRef("def_01K0SCIM000000000000000001")
	DirectorySyncOperation = types.NewOperationRef[DirectorySync]("directory.sync")
)

const Slug = "scim_directory_sync"

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

// Builder returns the SCIM definition builder
func Builder() definition.Builder {
	return definition.Builder(func(_ context.Context) (types.Definition, error) {
		clientRef := types.NewClientRef[any]()

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Version:     "v1",
				Family:      "scim",
				DisplayName: "SCIM Directory Sync",
				Description: "Synchronize directory objects through SCIM",
				Category:    "directory",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/scim/overview",
				Labels:      map[string]string{"protocol": "scim"},
				Active:      true,
				Visible:     true,
			},
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: providerkit.SchemaFrom[operatorConfig](),
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[userInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema: providerkit.SchemaFrom[credential](),
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         clientRef.ID(),
					Description: "SCIM API client",
					Build:       buildSCIMClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        DirectorySyncOperation.Name(),
					Description: "Synchronize directory state through SCIM",
					Topic:       DirectorySyncOperation.Topic(Slug),
					ClientRef:   clientRef.ID(),
					Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Handle:      runDirectorySyncOperation,
				},
			},
			Mappings: []types.MappingRegistration{
				{Schema: "directory_account", Variant: "default", Spec: types.MappingOverride{Version: "v1"}},
				{Schema: "directory_group", Variant: "default", Spec: types.MappingOverride{Version: "v1"}},
			},
			Webhooks: []types.WebhookRegistration{
				{
					Name:   "directory.push",
					Verify: verifyWebhook,
				},
			},
		}, nil
	})
}
