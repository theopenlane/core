package authentik

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Authentik definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "Authentik",
				DisplayName: "Authentik",
				Description: "Collect Authentik directory users, groups, and memberships for identity posture and access governance.",
				Category:    "identity",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/authentik/overview",
				Tags:        []string{"directory"},
				Active:      false,
				Visible:     false,
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         authentikCredential.ID(),
					Name:        "Authentik Credential",
					Description: "API token used to access Authentik instance data.",
					Schema:      authentikCredentialSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       authentikCredential.ID(),
					Name:                "Authentik API Token",
					Description:         "Configure Authentik access using an API token from your instance.",
					CredentialRefs:      []types.CredentialSlotID{authentikCredential.ID()},
					ClientRefs:          []types.ClientID{authentikClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         integration.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: authentikCredential.ID(),
						Description:   "Removes the stored API token from Openlane. If the token is no longer needed, revoke it in your Authentik admin panel under Directory > Tokens.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            authentikClient.ID(),
					CredentialRefs: []types.CredentialSlotID{authentikCredential.ID()},
					Description:    "Authentik API client",
					Build:          Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "Call Authentik API to verify token and instance connectivity",
					Topic:        definitionID.OperationTopic(healthCheckOperation.Name()),
					ClientRef:    authentikClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: healthCheckSchema,
					Handle:       HealthCheck{}.Handle(),
				},
				{
					Name:         directorySyncOperation.Name(),
					Description:  "Collect Authentik directory users, groups, and memberships as directory accounts",
					Topic:        definitionID.OperationTopic(directorySyncOperation.Name()),
					ClientRef:    authentikClient.ID(),
					ConfigSchema: directorySyncSchema,
					Policy:       types.ExecutionPolicy{Reconcile: true},
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
					IngestHandle: DirectorySync{}.IngestHandle(),
				},
			},
			Mappings: authentikMappings(),
		}, nil
	})
}
