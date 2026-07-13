package zitadel

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
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
			UserInput: &types.UserInputRegistration{
				Schema: jsonx.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         zitadelPATCredential.ID(),
					Name:        "Zitadel Personal Access Token",
					Description: "Personal Access Token used to access Zitadel instance data.",
					Schema:      zitadelPATCredentialSchema,
					Recommended: true,
				},
				{
					Ref:         zitadelOAuthCredential.ID(),
					Name:        "Zitadel OAuth (Client Credentials)",
					Description: "Service user Client ID and Client Secret used to access Zitadel instance data via the OAuth2 client-credentials grant.",
					Schema:      zitadelOAuthCredentialSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       zitadelPATCredential.ID(),
					Name:                "Zitadel Personal Access Token",
					Description:         "Configure Zitadel access using a Personal Access Token from your instance.",
					CredentialRefs:      []types.CredentialSlotID{zitadelPATCredential.ID()},
					ClientRefs:          []types.ClientID{zitadelClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         integration.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: zitadelPATCredential.ID(),
						Description:   "Removes the stored Personal Access Token from Openlane. If the token is no longer needed, revoke it in your Zitadel admin console under Personal Access Tokens.",
					},
				},
				{
					CredentialRef:       zitadelOAuthCredential.ID(),
					Name:                "Zitadel OAuth (Client Credentials)",
					Description:         "Configure Zitadel access using a service user Client ID and Client Secret.",
					CredentialRefs:      []types.CredentialSlotID{zitadelOAuthCredential.ID()},
					ClientRefs:          []types.ClientID{zitadelClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         integration.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: zitadelOAuthCredential.ID(),
						Description:   "Removes the stored Client ID and Client Secret from Openlane. If the service user is no longer needed, delete it in your Zitadel admin console.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            zitadelClient.ID(),
					CredentialRefs: []types.CredentialSlotID{zitadelPATCredential.ID(), zitadelOAuthCredential.ID()},
					Description:    "Zitadel user service API client",
					Build:          Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "Call Zitadel API to verify Personal Access Token and instance connectivity",
					Topic:        definitionID.OperationTopic(healthCheckOperation.Name()),
					ClientRef:    zitadelClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: healthCheckSchema,
					Handle:       HealthCheck{}.Handle(),
				},
				{
					Name:                directorySyncOperation.Name(),
					Description:         "Collect Zitadel directory users as directory accounts",
					Topic:               definitionID.OperationTopic(directorySyncOperation.Name()),
					ClientRef:           zitadelClient.ID(),
					ConfigSchema:        directorySyncSchema,
					Policy:              types.ExecutionPolicy{Reconcile: true},
					SkipDefaultLookback: true,
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
						},
					},
					IngestHandle: DirectorySync{}.IngestHandle(),
				},
			},
			Mappings: zitadelMappings(),
		}, nil
	})
}