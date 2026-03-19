package gcpscc

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the GCP SCC definition builder
func Builder() definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Family:      "gcp",
				DisplayName: "GCP Security Command Center",
				Description: "Collect Google Cloud Security Command Center findings for security posture reporting.",
				Category:    "cloud",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/gcp_scc/overview",
				Labels:      map[string]string{"vendor": "google", "product": "Security Command Center"},
				Active:      false,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema: providerkit.SchemaFrom[CredentialSchema](),
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         SCCClient.ID(),
					Description: "Google Cloud Security Command Center v2 client",
					Build:       Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Verify GCP SCC access",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   SCCClient.ID(),
					Handle:      HealthCheck{}.Handle(),
				},
				{
					Name:         FindingsCollectOperation.Name(),
					Description:  "Collect GCP Security Command Center findings for vulnerability ingestion",
					Topic:        FindingsCollectOperation.Topic(Slug),
					ClientRef:    SCCClient.ID(),
					ConfigSchema: providerkit.SchemaFrom[FindingsConfig](),
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
						},
					},
					IngestHandle: FindingsCollect{}.IngestHandle(),
				},
				{
					Name:        SettingsScanOperation.Name(),
					Description: "Scan GCP Security Command Center source and notification settings",
					Topic:       SettingsScanOperation.Topic(Slug),
					ClientRef:   SCCClient.ID(),
					Handle:      SettingsScan{}.Handle(),
				},
			},
			Mappings: gcpsccMappings(),
		}, nil
	})
}
