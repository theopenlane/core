package zitadel

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// Builder returns the Zitadel definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "Zitadel",
				DisplayName: "Zitadel",
				Description: "Collect Zitadel directory users for identity posture and access governance.",
				Category:    "identity",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/zitadel/overview",
				Tags:        []string{"directory"},
				Active:      false,
				Visible:     true,
			},
			UserInput:    userInput.Registration(),
			HealthCheck:  zitadelClient.HealthCheck(checkHealth),
			Installation: integration.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				zitadelPATCredential.Registration(types.CredentialRegistration{
					Name:        "Zitadel Personal Access Token",
					Description: "Personal Access Token used to access Zitadel instance data.",
				}),
				zitadelOAuthCredential.Registration(types.CredentialRegistration{
					Name:        "Zitadel OAuth (Client Credentials)",
					Description: "Service user Client ID and Client Secret used to access Zitadel instance data via the OAuth2 client-credentials grant.",
					Recommended: true,
				}),
			},
			Connections: []types.ConnectionRegistration{
				zitadelPATConnection.Registration(types.ConnectionRegistration{
					Name:        "Zitadel Personal Access Token",
					Description: "Configure Zitadel access using a Personal Access Token from your instance.",
					Disconnect: &types.DisconnectRegistration{
						Description: "Removes the stored Personal Access Token from Openlane. If the token is no longer needed, revoke it in your Zitadel admin console under Personal Access Tokens.",
					},
				}),
				zitadelOAuthConnection.Registration(types.ConnectionRegistration{
					Name:        "Zitadel OAuth (Client Credentials)",
					Description: "Configure Zitadel access using a service user Client ID and Client Secret.",
					Disconnect: &types.DisconnectRegistration{
						Description: "Removes the stored Client ID and Client Secret from Openlane. If the service user is no longer needed, delete it in your Zitadel admin console.",
					},
				}),
			},
			Clients: []types.ClientRegistration{
				zitadelClient.Registration(Client{}.Build, types.ClientRegistration{
					Description: "Zitadel user service API client",
				}),
			},
			Operations: []types.OperationRegistration{
				directorySyncOperation.Registration(definitionID, types.OperationRegistration{
					Description:         "Collect Zitadel directory users as directory accounts",
					Policy:              types.ExecutionPolicy{Reconcile: true, Snapshot: true},
					SkipDefaultLookback: true,
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaDirectoryAccount.Name,
						},
					},
				}),
			},
			Mappings: zitadelMappings(),
		}, nil
	})
}
