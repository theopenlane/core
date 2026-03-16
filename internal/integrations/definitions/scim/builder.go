package scim

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the SCIM definition builder
func Builder() definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
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
					Name:        HealthDefaultOperation.Name(),
					Description: "Validate SCIM configuration",
					Topic:       HealthDefaultOperation.Topic(Slug),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      HealthCheck{}.Handle(),
				},
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
					Handle: DirectorySync{}.Handle(),
				},
			},
			Mappings: scimMappings(),
		}, nil
	})
}
