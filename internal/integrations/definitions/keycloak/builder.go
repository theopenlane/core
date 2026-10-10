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
			ID:           definitionID.ID(),
			Family:       "Keycloak",
			DisplayName:  "Keycloak",
			Description:  "Collect Keycloak realm users, groups, and memberships for identity posture and access governance.",
			Category:     "identity",
			DocsURL:      "https://docs.theopenlane.io/docs/platform/integrations/keycloak",
			Tags:         []string{"directory"},
			Active:       false,
			Visible:      true,
			UserInput:    userInput.Registration(),
			Installation: installation.Registration(),
			Connections: []types.Connector{
				clientCredentials.
					Name("Keycloak Client Credentials").
					Description("Configure Keycloak access using client credentials from your realm.").
					Provides(buildClient).
					Verified(verify).
					Disconnects("Removes the stored client credentials from Openlane. If the client is no longer needed, disable or delete it in your Keycloak admin console under Clients.", nil),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[providerkit.DirectorySync]().
					Ingests(runDirectorySync).
					Policy(types.ExecutionPolicy{Reconcile: true, Snapshot: true}).
					SkipDefaultLookback().
					Ingest(providerkit.DirectoryIngestContracts()...).
					Permissions("view-realm", "view-users", "query-groups", "view-events").
					Description("Collect Keycloak realm users, groups, and memberships as directory accounts").
					Registration(),
			},
			Mappings: providerkit.DirectoryMappings(mapExprDirectoryAccount, mapExprDirectoryGroup, mapExprDirectoryMembership),
		}, nil
	})
}
