package authentik

import (
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
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
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/authentik",
				Tags:        []string{"directory"},
				Active:      true,
				Visible:     true,
			},
			UserInput:    userInput.Registration(),
			HealthCheck:  authentikClient.HealthCheck(checkHealth),
			Installation: installation.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				authentikCredential.Registration(types.CredentialRegistration{
					Name:        "Authentik Credential",
					Description: "API token used to access Authentik instance data.",
				}),
			},
			Connections: []types.ConnectionRegistration{
				authentikConnection.Registration(types.ConnectionRegistration{
					Name:        "Authentik API Token",
					Description: "Configure Authentik access using an API token from your instance.",
					Disconnect: &types.DisconnectRegistration{
						Description: "Removes the stored API token from Openlane. If the token is no longer needed, revoke it in your Authentik admin panel under Directory > Tokens.",
					},
				}),
			},
			Clients: []types.ClientRegistration{
				authentikClient.Registration(Client{}.Build, types.ClientRegistration{
					Description: "Authentik API client",
				}),
			},
			Operations: []types.OperationRegistration{
				directorySyncOperation.Registration(definitionID, types.OperationRegistration{
					Description: "Collect Authentik directory users, groups, and memberships as directory accounts",
					Policy:      types.ExecutionPolicy{Reconcile: true, Snapshot: true},
					Ingest:      providerkit.DirectoryIngestContracts(),
				}),
			},
			Mappings: providerkit.DirectoryMappings(mapExprDirectoryAccount, mapExprDirectoryGroup, mapExprDirectoryMembership),
		}, nil
	})
}
