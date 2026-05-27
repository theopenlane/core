package keycloak

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// Builder returns the Keycloak definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "Keycloak",
				DisplayName: "Keycloak",
				Description: "Collect Keycloak realm users, groups, and memberships for identity posture and access governance.",
				Category:    "identity",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/keycloak/overview",
				Tags:        []string{"directory"},
				Active:      false,
				Visible:     false,
			},
			UserInput: &types.UserInputRegistration{
				Schema: jsonx.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         keycloakCredential.ID(),
					Name:        "Keycloak Credential",
					Description: "Client credentials used to access Keycloak realm data.",
					Schema:      keycloakCredentialSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       keycloakCredential.ID(),
					Name:                "Keycloak Client Credentials",
					Description:         "Configure Keycloak access using client credentials from your realm.",
					CredentialRefs:      []types.CredentialSlotID{keycloakCredential.ID()},
					ClientRefs:          []types.ClientID{keycloakClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         integration.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: keycloakCredential.ID(),
						Description:   "Removes the stored client credentials from Openlane. If the client is no longer needed, disable or delete it in your Keycloak admin console under Clients.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            keycloakClient.ID(),
					CredentialRefs: []types.CredentialSlotID{keycloakCredential.ID()},
					Description:    "Keycloak API client",
					Build:          Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "Call Keycloak realm API to verify client credentials and realm connectivity",
					Topic:        definitionID.OperationTopic(healthCheckOperation.Name()),
					ClientRef:    keycloakClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: healthCheckSchema,
					Handle:       HealthCheck{}.Handle(),
				},
				{
					Name:                directorySyncOperation.Name(),
					Description:         "Collect Keycloak realm users, groups, and memberships as directory accounts",
					Topic:               definitionID.OperationTopic(directorySyncOperation.Name()),
					ClientRef:           keycloakClient.ID(),
					ConfigSchema:        directorySyncSchema,
					Policy:              types.ExecutionPolicy{Reconcile: true},
					SkipDefaultLookback: true,
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
			Mappings: keycloakMappings(),
		}, nil
	})
}