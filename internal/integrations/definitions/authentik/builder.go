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
			ID:           definitionID.ID(),
			Family:       "Authentik",
			DisplayName:  "Authentik",
			Description:  "Collect Authentik directory users, groups, and memberships for identity posture and access governance.",
			Category:     "identity",
			DocsURL:      "https://docs.theopenlane.io/docs/platform/integrations/authentik",
			Tags:         []string{"directory"},
			Active:       true,
			Visible:      true,
			UserInput:    userInput.Registration(),
			Installation: installation.Registration(),
			Connections: []types.Connector{
				authentikConnection.
					Name("Authentik API Token").
					Description("Configure Authentik access using an API token from your instance.").
					Provides(buildClient).
					Verified(verify).
					Disconnects("Removes the stored API token from Openlane. If the token is no longer needed, revoke it in your Authentik admin panel under Directory > Tokens.", nil),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[providerkit.DirectorySync]().
					Ingests(runDirectorySync).
					Policy(types.ExecutionPolicy{Reconcile: true, Snapshot: true}).
					Ingest(providerkit.DirectoryIngestContracts()...).
					Description("Collect Authentik directory users, groups, and memberships as directory accounts").
					Registration(),
			},
			Mappings: providerkit.DirectoryMappings(mapExprDirectoryAccount, mapExprDirectoryGroup, mapExprDirectoryMembership),
		}, nil
	})
}
