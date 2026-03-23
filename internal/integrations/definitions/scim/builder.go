package scim

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the SCIM definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Family:      "scim",
				DisplayName: "SCIM Directory Sync",
				Description: "Synchronize directory objects through SCIM",
				Category:    "directory",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/scim/overview",
				Labels:      map[string]string{"protocol": "scim"},
				Active:      true,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Report push-based SCIM health status",
					Topic:       types.OperationTopic(DefinitionID.ID(), HealthDefaultOperation.Name()),
					Policy:      types.ExecutionPolicy{Inline: true},
					Handle:      HealthCheck{}.Handle(),
				},
				{
					Name:        DirectorySyncOperation.Name(),
					Description: "Synchronize directory state through SCIM",
					Topic:       types.OperationTopic(DefinitionID.ID(), DirectorySyncOperation.Name()),
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
						},
						{
							Schema: integrationgenerated.IntegrationMappingSchemaDirectoryGroup,
						},
						{
							Schema: integrationgenerated.IntegrationMappingSchemaDirectoryMembership,
						},
					},
					Handle: DirectorySync{}.Handle(),
				},
			},
			Mappings: scimMappings(),
		}, nil
	})
}
