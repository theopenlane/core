package cloudflare

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Cloudflare definition builder
func Builder() definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
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
					Ref:         cloudflareCredential,
					Name:        "Cloudflare Credential",
					Description: "Credential slot used by the Cloudflare client in this definition.",
					Schema:      providerkit.SchemaFrom[CredentialSchema](),
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       cloudflareCredential,
					Name:                "Cloudflare API Token",
					Description:         "Configure Cloudflare access using an API token scoped to the account and zones you want Openlane to inspect.",
					CredentialRefs:      []types.CredentialRef{cloudflareCredential},
					ClientRefs:          []types.ClientID{CloudflareClient.ID()},
					ValidationOperation: HealthDefaultOperation.Name(),
					Installation:        Installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: cloudflareCredential,
						Name:          "Disconnect Cloudflare API Token",
						Description:   "Remove the persisted Cloudflare API token and disconnect this installation from Openlane.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            CloudflareClient.ID(),
					CredentialRefs: []types.CredentialRef{cloudflareCredential},
					Description:    "Cloudflare REST API client",
					Build:          Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Verify Cloudflare API token via /user/tokens/verify",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   CloudflareClient.ID(),
					Handle:      HealthCheck{}.Handle(),
				},
				{
					Name:        DirectorySyncOperation.Name(),
					Description: "Collect account members as directory accounts",
					Topic:       DirectorySyncOperation.Topic(Slug),
					ClientRef:   CloudflareClient.ID(),
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
