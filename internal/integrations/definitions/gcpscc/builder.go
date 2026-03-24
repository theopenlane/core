package gcpscc

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the GCP SCC definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
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
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         sccCredential.ID(),
					Name:        "GCP SCC Credential",
					Description: "Credential slot used by the GCP Security Command Center client in this definition.",
					Schema:      sccSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       sccCredential.ID(),
					Name:                "GCP Service Account",
					Description:         "Configure Security Command Center access using a service account credential payload.",
					CredentialRefs:      []types.CredentialSlotID{sccCredential.ID()},
					ClientRefs:          []types.ClientID{sccClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Installation:        installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: sccCredential.ID(),
						Name:          "Disconnect GCP Service Account",
						Description:   "Remove the persisted GCP Security Command Center credential and disconnect this installation from Openlane.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            sccClient.ID(),
					CredentialRefs: []types.CredentialSlotID{sccCredential.ID()},
					Description:    "Google Cloud Security Command Center v2 client",
					Build:          Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "Verify GCP SCC access",
					Topic:        types.OperationTopic(definitionID.ID(), healthCheckOperation.Name()),
					ClientRef:    sccClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					Handle:       HealthCheck{}.Handle(),
					ConfigSchema: healthCheckSchema,
				},
				{
					Name:         findingsCollectOperation.Name(),
					Description:  "Collect GCP Security Command Center findings for vulnerability ingestion",
					Topic:        types.OperationTopic(definitionID.ID(), findingsCollectOperation.Name()),
					ClientRef:    sccClient.ID(),
					ConfigSchema: findingsCollectSchema,
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
						},
					},
					IngestHandle: FindingsCollect{}.IngestHandle(),
				},
			},
			Mappings: gcpsccMappings(),
		}, nil
	})
}
