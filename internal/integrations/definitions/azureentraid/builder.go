package azureentraid

import (
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// Builder returns the Azure EntraID definition builder with the supplied operator config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			ID:          definitionID.ID(),
			Family:      "Azure",
			DisplayName: "Azure EntraID",
			Description: "Connect to Microsoft Graph to validate tenant access and inspect Azure Entra ID organization metadata.",
			Category:    "identity",
			DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/azure_entra_id/overview",
			Tags:        []string{"directory"},
			Active:      true,
			Visible:     true,
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[Config](),
			},
			UserInput:    userInput.Registration(),
			Installation: installation.Registration(),
			Connections: []types.Connector{
				adminConsent.
					Name("Azure Entra ID Admin Consent").
					Description("Connect your Azure Entra ID tenant using admin consent.").
					Authenticates(adminConsentFlow(cfg)).
					Provides(tokenCredentialClient(cfg)).
					Provides(graphClient(cfg)).
					Verified(verify).
					Disconnects("Removes the stored credential from Openlane. To fully revoke access, remove the Openlane app from your Azure Entra ID enterprise applications.", nil),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[DirectorySync]().
					Ingests(runDirectorySync).
					Policy(types.ExecutionPolicy{Reconcile: true, Snapshot: true}).
					Schedule(gala.NewFullFetchSchedule()).
					SkipDefaultLookback().
					Ingest(providerkit.DirectoryIngestContracts()...).
					Permissions("User.Read.All", "Group.Read.All", "GroupMember.Read.All", "Directory.Read.All").
					HealthCheck(probeDirectory).
					Description("Collect Azure Entra ID users, groups, and memberships as directory accounts").
					Registration(),
			},
			Mappings: providerkit.DirectoryMappings(mapExprDirectoryAccount, mapExprDirectoryGroup, mapExprDirectoryMembership),
		}, nil
	})
}
