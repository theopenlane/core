package scim

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// OperatorConfig holds operator-owned defaults that apply across all SCIM installations
type OperatorConfig struct {
	// DefaultProvisioningMode is the fallback provisioning mode for new installations
	DefaultProvisioningMode string `json:"defaultProvisioningMode,omitempty" jsonschema:"title=Default Provisioning Mode"`
	// DefaultBasePath is the fallback SCIM base path for new installations
	DefaultBasePath string `json:"defaultBasePath,omitempty" jsonschema:"title=Default Base Path"`
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// Label is the user-defined display label for the installation
	Label string `json:"label,omitempty" jsonschema:"title=Tenant Label"`
	// TenantKey is the SCIM tenant key
	TenantKey string `json:"tenantKey,omitempty" jsonschema:"title=Tenant Key"`
	// MappingMode controls how directory objects are mapped
	MappingMode string `json:"mappingMode,omitempty" jsonschema:"title=Mapping Mode"`
	// ProvisioningMode controls how directory objects are provisioned
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
	return definition.Builder(func() (types.Definition, error) {
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
				Schema: providerkit.SchemaFrom[OperatorConfig](),
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema: providerkit.SchemaFrom[credential](),
			},
			Operations: []types.OperationRegistration{
				{
					Name:        DirectorySyncOperation.Name(),
					Description: "Synchronize directory state through SCIM",
					Topic:       DirectorySyncOperation.Topic(Slug),
					Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Ingest: []types.IngestContract{
						{
							Schema:         integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
							EnsurePayloads: true,
						},
						{
							Schema:         integrationgenerated.IntegrationMappingSchemaDirectoryGroup,
							EnsurePayloads: true,
						},
						{
							Schema:         integrationgenerated.IntegrationMappingSchemaDirectoryMembership,
							EnsurePayloads: true,
						},
					},
					Handle: func(_ context.Context, _ types.OperationRequest) (json.RawMessage, error) {
						return jsonx.ToRawMessage(DirectorySync{
							Message: directorySyncAckMessage,
						})
					},
				},
			},
			Mappings: scimMappings(),
		}, nil
	})
}
