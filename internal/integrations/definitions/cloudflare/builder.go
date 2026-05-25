package cloudflare

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

// Builder returns the Cloudflare definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "Cloudflare",
				DisplayName: "Cloudflare",
				Description: "Perform directory sync and asset collection from Cloudflare.",
				Category:    "security-posture",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/cloudflare/overview",
				Tags:        []string{"directory", "assets"},
				Active:      true,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: jsonx.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         cloudflareCredential.ID(),
					Name:        "Cloudflare API Credential",
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
					SkipFirstReconcile:  true,
					Integration:         installation.Registration(),
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
					Topic:        definitionID.OperationTopic(healthCheckOperation.Name()),
					ClientRef:    cloudflareClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					Handle:       HealthCheck{}.Handle(),
					ConfigSchema: healthCheckSchema,
				},
				{
					Name:           directorySyncOperation.Name(),
					Description:    "Collect account members as directory accounts",
					Topic:          definitionID.OperationTopic(directorySyncOperation.Name()),
					ClientRef:      cloudflareClient.ID(),
					ConfigSchema:   directorySyncSchema,
					Policy:         types.ExecutionPolicy{Reconcile: true},
					Disabled:       providerkit.DisabledWhen(func(u UserInput) bool { return u.DirectorySync.Disable }),
					ConfigResolver: providerkit.ConfigFrom(func(u UserInput) DirectorySync { return u.DirectorySync }),
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
					IngestHandle:        DirectorySync{}.IngestHandle(),
					SkipDefaultLookback: true,
					RequiredPermissions: []string{"Account Settings Read", "Access: Users Read", "Access: Groups Read", "Access: Organizations, Identity Providers, and Groups Read"},
					ReconcileSchedule:   gala.NewFullFetchSchedule(),
				},
				{
					Name:           findingsSyncOperation.Name(),
					Description:    "Collect Cloudflare Security Center insights as findings",
					Topic:          definitionID.OperationTopic(findingsSyncOperation.Name()),
					ClientRef:      cloudflareClient.ID(),
					ConfigSchema:   findingsSyncSchema,
					Policy:         types.ExecutionPolicy{Reconcile: true},
					Disabled:       providerkit.DisabledWhen(func(u UserInput) bool { return u.FindingsSync.Disable }),
					ConfigResolver: providerkit.ConfigFrom(func(u UserInput) FindingsSync { return u.FindingsSync }),
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaFinding,
						},
					},
					IngestHandle:        FindingsCollect{}.IngestHandle(),
					RequiredPermissions: []string{"Account Security Center Insights Read"},
				},
			},
			Mappings: cloudflareMappings(),
		}, nil
	})
}
