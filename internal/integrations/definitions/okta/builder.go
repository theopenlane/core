package okta

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Okta definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "okta",
				DisplayName: "Okta",
				Description: "Collect Okta tenant and sign-on policy metadata for identity posture and access governance.",
				Category:    "identity",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/okta/overview",
				Tags:        []string{"directory-sync"},
				Active:      false,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         oktaCredential.ID(),
					Name:        "Okta Credential",
					Description: "API token used to access Okta organization data.",
					Schema:      oktaCredentialSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       oktaCredential.ID(),
					Name:                "Okta API Token",
					Description:         "Configure Okta access using an API token from your organization.",
					CredentialRefs:      []types.CredentialSlotID{oktaCredential.ID()},
					ClientRefs:          []types.ClientID{oktaClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         integration.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: oktaCredential.ID(),
						Description:   "Removes the stored API token from Openlane. If the token is no longer needed, revoke it in your Okta admin console under Security > API.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            oktaClient.ID(),
					CredentialRefs: []types.CredentialSlotID{oktaCredential.ID()},
					Description:    "Okta API client",
					Build:          Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "Call Okta user API to verify API token",
					Topic:        definitionID.OperationTopic(healthCheckOperation.Name()),
					ClientRef:    oktaClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: healthCheckSchema,
					Handle:       HealthCheck{}.Handle(),
				},
				{
					Name:         directorySyncOperation.Name(),
					Description:  "Collect Okta directory users, groups, and memberships as directory accounts",
					Topic:        definitionID.OperationTopic(directorySyncOperation.Name()),
					ClientRef:    oktaClient.ID(),
					ConfigSchema: directorySyncSchema,
					Policy:       types.ExecutionPolicy{Reconcile: true},
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
			Mappings: oktaMappings(),
		}, nil
	})
}
