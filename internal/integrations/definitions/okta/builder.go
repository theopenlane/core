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
				Category:    "sso",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/okta/overview",
				Labels:      map[string]string{"vendor": "okta", "product": "identity"},
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
					Description: "Credential slot used by the Okta client in this definition.",
					Schema:      oktaCredentialSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       oktaCredential.ID(),
					Name:                "Okta API Token",
					Description:         "Configure Okta tenant access using an API token issued for the target organization.",
					CredentialRefs:      []types.CredentialSlotID{oktaCredential.ID()},
					ClientRefs:          []types.ClientID{oktaClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Installation:        installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: oktaCredential.ID(),
						Name:          "Disconnect Okta API Token",
						Description:   "Remove the persisted Okta API token and disconnect this installation from Openlane.",
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
					Topic:        types.OperationTopic(definitionID.ID(), healthCheckOperation.Name()),
					ClientRef:    oktaClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: healthCheckSchema,
					Handle:       HealthCheck{}.Handle(),
				},
				{
					Name:         directorySyncOperation.Name(),
					Description:  "Collect Okta directory users, groups, and memberships as directory accounts",
					Topic:        types.OperationTopic(definitionID.ID(), directorySyncOperation.Name()),
					ClientRef:    oktaClient.ID(),
					ConfigSchema: directorySyncSchema,
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
