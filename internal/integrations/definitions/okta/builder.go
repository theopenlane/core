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
			HealthCheck:  oktaClient.HealthCheck(checkHealth),
			Installation: installation.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				oktaCredential.Registration(types.CredentialRegistration{
					Name:        "Okta Credential",
					Description: "API token used to access Okta organization data.",
				}),
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef: oktaCredential.ID(),
					Name:          "Okta API Token",
					Description:   "Configure Okta access using an API token from your organization.",
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: oktaCredential.ID(),
						Description:   "Removes the stored API token from Openlane. If the token is no longer needed, revoke it in your Okta admin console under Security > API.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				oktaClient.Registration(Client{}.Build, types.ClientRegistration{
					Description: "Okta API client",
				}),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[DirectorySync]().
					Ingests(oktaClient, runDirectorySync).
					Policy(types.ExecutionPolicy{Reconcile: true, Snapshot: true}).
					SkipDefaultLookback().
					Ingest(providerkit.DirectoryIngestContracts()...).
					Registration(definitionID, types.OperationRegistration{
						Description: "Collect Okta directory users, groups, and memberships as directory accounts",
					}),
			},
			Mappings: providerkit.DirectoryMappings(mapExprDirectoryAccount, mapExprDirectoryGroup, mapExprDirectoryMembership),
		}, nil
	})
}
