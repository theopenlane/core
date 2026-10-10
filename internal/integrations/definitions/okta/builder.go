package okta

import (
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// Builder returns the Okta definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			ID:           definitionID.ID(),
			Family:       "Okta",
			DisplayName:  "Okta",
			Description:  "Collect Okta tenant and sign-on policy metadata for identity posture and access governance.",
			Category:     "identity",
			DocsURL:      "https://docs.theopenlane.io/docs/platform/integrations/okta",
			Tags:         []string{"directory"},
			Active:       false,
			Visible:      true,
			UserInput:    userInput.Registration(),
			Installation: installation.Registration(),
			Connections: []types.Connector{
				oktaConnection.
					Name("Okta API Token").
					Description("Configure Okta access using an API token from your organization.").
					Provides(buildClient).
					Verified(verify).
					Disconnects("Removes the stored API token from Openlane. If the token is no longer needed, revoke it in your Okta admin console under Security > API.", nil),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[DirectorySync]().
					Ingests(runDirectorySync).
					HealthCheck(probeUser).
					Policy(types.ExecutionPolicy{Reconcile: true, Snapshot: true}).
					SkipDefaultLookback().
					Ingest(providerkit.DirectoryIngestContracts()...).
					Description("Collect Okta directory users, groups, and memberships as directory accounts").
					Registration(),
			},
			Mappings: providerkit.DirectoryMappings(mapExprDirectoryAccount, mapExprDirectoryGroup, mapExprDirectoryMembership),
		}, nil
	})
}
