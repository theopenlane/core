package okta

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Okta definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
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
					Schema:      providerkit.SchemaFrom[CredentialSchema](),
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       oktaCredential.ID(),
					Name:                "Okta API Token",
					Description:         "Configure Okta tenant access using an API token issued for the target organization.",
					CredentialRefs:      []types.CredentialSlotID{oktaCredential.ID()},
					ClientRefs:          []types.ClientID{OktaClient.ID()},
					ValidationOperation: HealthDefaultOperation.Name(),
					Installation:        Installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: oktaCredential.ID(),
						Name:          "Disconnect Okta API Token",
						Description:   "Remove the persisted Okta API token and disconnect this installation from Openlane.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            OktaClient.ID(),
					CredentialRefs: []types.CredentialSlotID{oktaCredential.ID()},
					Description:    "Okta API client",
					Build:          Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call Okta user API to verify API token",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   OktaClient.ID(),
					Handle:      HealthCheck{}.Handle(),
				},
				{
					Name:        PoliciesCollectOperation.Name(),
					Description: "Collect sign-on policy metadata for posture analysis",
					Topic:       PoliciesCollectOperation.Topic(Slug),
					ClientRef:   OktaClient.ID(),
					Handle:      PoliciesCollect{}.Handle(),
				},
				{
					Name:        DirectorySyncOperation.Name(),
					Description: "Collect Okta directory users, groups, and memberships as directory accounts",
					Topic:       DirectorySyncOperation.Topic(Slug),
					ClientRef:   OktaClient.ID(),
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
