package gcpscc

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the GCP SCC definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
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
					ClientRefs:          []types.ClientID{SCCClient.ID()},
					ValidationOperation: HealthDefaultOperation.Name(),
					Installation:        Installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: sccCredential.ID(),
						Name:          "Disconnect GCP Service Account",
						Description:   "Remove the persisted GCP Security Command Center credential and disconnect this installation from Openlane.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            SCCClient.ID(),
					CredentialRefs: []types.CredentialSlotID{sccCredential.ID()},
					Description:    "Google Cloud Security Command Center v2 client",
					Build:          Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Verify GCP SCC access",
					Topic:       types.OperationTopic(DefinitionID.ID(), HealthDefaultOperation.Name()),
					ClientRef:   SCCClient.ID(),
					Policy:      types.ExecutionPolicy{Inline: true},
					Handle:      HealthCheck{}.Handle(),
				},
				{
					Name:         FindingsCollectOperation.Name(),
					Description:  "Collect GCP Security Command Center findings for vulnerability ingestion",
					Topic:        types.OperationTopic(DefinitionID.ID(), FindingsCollectOperation.Name()),
					ClientRef:    SCCClient.ID(),
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
