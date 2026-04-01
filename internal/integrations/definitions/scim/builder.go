package scim

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns a registry builder that constructs the SCIM directory sync definition
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Family:      "scim",
				DisplayName: "SCIM 2.0",
				Description: "Synchronize directory objects through SCIM",
				Category:    "identity",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/scim/overview",
				Tags:        []string{"directory-sync"},
				Active:      true,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			Webhooks: []types.WebhookRegistration{
				{
					Name:                SCIMAuthWebhook.Name(),
					EndpointURLTemplate: "/v1/integrations/scim/{endpointID}/v2",
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "Report push-based SCIM health status",
					Topic:        DefinitionID.OperationTopic(healthCheckOperation.Name()),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: healthCheckSchema,
					Handle:       providerkit.StaticHandler(HealthCheck{}.Run),
				},
				{
					Name:         directorySyncOperation.Name(),
					Description:  "Synchronize directory state through SCIM",
					Topic:        DefinitionID.OperationTopic(directorySyncOperation.Name()),
					ConfigSchema: directorySyncSchema,
					Policy:       types.ExecutionPolicy{Inline: true},
					Ingest: []types.IngestContract{
						{Schema: integrationgenerated.IntegrationMappingSchemaDirectoryAccount},
						{Schema: integrationgenerated.IntegrationMappingSchemaDirectoryGroup},
						{Schema: integrationgenerated.IntegrationMappingSchemaDirectoryMembership},
					},
					Handle: providerkit.StaticHandler(DirectorySync{}.Run),
				},
			},
			Mappings: scimMappings(),
		}, nil
	})
}
