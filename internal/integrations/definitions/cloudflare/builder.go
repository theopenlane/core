package cloudflare

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Cloudflare definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "cloudflare",
				DisplayName: "Cloudflare",
				Description: "Validate Cloudflare account access and collect security-relevant account and zone context for posture workflows.",
				Category:    "security",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/cloudflare/overview",
				Labels:      map[string]string{"vendor": "cloudflare", "product": "zero-trust"},
				Active:      true,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         cloudflareCredential.ID(),
					Name:        "Cloudflare Credential",
					Description: "Credential slot used by the Cloudflare client in this definition.",
					Schema:      cloudflareSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       cloudflareCredential.ID(),
					Name:                "Cloudflare API Token",
					Description:         "Configure Cloudflare access using an API token scoped to the account and zones you want Openlane to inspect.",
					CredentialRefs:      []types.CredentialSlotID{cloudflareCredential.ID()},
					ClientRefs:          []types.ClientID{cloudflareClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Installation:        installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: cloudflareCredential.ID(),
						Name:          "Disconnect Cloudflare API Token",
						Description:   "Remove the persisted Cloudflare API token and disconnect this installation from Openlane.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            cloudflareClient.ID(),
					CredentialRefs: []types.CredentialSlotID{cloudflareCredential.ID()},
					Description:    "Cloudflare REST API client",
					Build:          Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "Verify Cloudflare API token via /user/tokens/verify",
					Topic:        types.OperationTopic(definitionID.ID(), healthCheckOperation.Name()),
					ClientRef:    cloudflareClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					Handle:       HealthCheck{}.Handle(),
					ConfigSchema: healthCheckSchema,
				},
				{
					Name:         directorySyncOperation.Name(),
					Description:  "Collect account members as directory accounts",
					Topic:        types.OperationTopic(definitionID.ID(), directorySyncOperation.Name()),
					ClientRef:    cloudflareClient.ID(),
					ConfigSchema: directorySyncSchema,
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
						},
					},
					IngestHandle: DirectorySync{}.IngestHandle(),
				},
			},
			Mappings: cloudflareMappings(),
		}, nil
	})
}
