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
				Category:    "security-posture",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/cloudflare/overview",
				Tags:        []string{"directory-sync", "assets"},
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
					Description: "API token used to access Cloudflare account and zone data.",
					Schema:      cloudflareSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       cloudflareCredential.ID(),
					Name:                "Cloudflare API Token",
					Description:         "Configure Cloudflare access using an API token scoped to your account and zones.",
					CredentialRefs:      []types.CredentialSlotID{cloudflareCredential.ID()},
					ClientRefs:          []types.ClientID{cloudflareClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Installation:        installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: cloudflareCredential.ID(),
						Description:   "Removes the stored API token from Openlane. If the token is no longer needed, revoke it in your Cloudflare dashboard.",
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
