package keycloak

import (
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
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
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/keycloak",
				Tags:        []string{"directory"},
				Active:      false,
				Visible:     true,
			},
			UserInput:    userInput.Registration(),
			HealthCheck:  keycloakClient.HealthCheck(checkHealth),
			Installation: installation.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				keycloakCredential.Registration(types.CredentialRegistration{
					Name:        "Keycloak Credential",
					Description: "Client credentials used to access Keycloak realm data.",
				}),
			},
			Connections: []types.ConnectionRegistration{
				keycloakConnection.Registration(types.ConnectionRegistration{
					Name:        "Keycloak Client Credentials",
					Description: "Configure Keycloak access using client credentials from your realm.",
					Disconnect: &types.DisconnectRegistration{
						Description: "Removes the stored client credentials from Openlane. If the client is no longer needed, disable or delete it in your Keycloak admin console under Clients.",
					},
				}),
			},
			Clients: []types.ClientRegistration{
				keycloakClient.Registration(Client{}.Build, types.ClientRegistration{
					Description: "Keycloak API client",
				}),
			},
			Operations: []types.OperationRegistration{
				directorySyncOperation.Registration(definitionID, types.OperationRegistration{
					Description:         "Collect Keycloak realm users, groups, and memberships as directory accounts",
					Policy:              types.ExecutionPolicy{Reconcile: true, Snapshot: true},
					SkipDefaultLookback: true,
					RequiredPermissions: []string{"view-realm", "view-users", "query-groups", "view-events"},
					Ingest:              providerkit.DirectoryIngestContracts(),
				}),
			},
			Mappings: providerkit.DirectoryMappings(mapExprDirectoryAccount, mapExprDirectoryGroup, mapExprDirectoryMembership),
		}, nil
	})
}
