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
				Category:    "security-posture",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/gcp-scc",
				Tags:        []string{"vulnerabilities", "assets"},
				Active:      true,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         sccCredential.ID(),
					Name:        "GCP SCC Credential",
					Description: "GCP service account key used to access Security Command Center.",
					Schema:      sccSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       sccCredential.ID(),
					Name:                "GCP Service Account",
					Description:         "Configure Security Command Center access using a GCP service account.",
					CredentialRefs:      []types.CredentialSlotID{sccCredential.ID()},
					ClientRefs:          []types.ClientID{sccClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: sccCredential.ID(),
						Description:   "Removes the stored service account credentials from Openlane. If the GCP service account is no longer needed, delete it from your Google Cloud project.",
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
					Topic:        definitionID.OperationTopic(healthCheckOperation.Name()),
					ClientRef:    sccClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					Handle:       HealthCheck{}.Handle(),
					ConfigSchema: healthCheckSchema,
				},
				{
					Name:         findingsCollectOperation.Name(),
					Description:  "Collect GCP Security Command Center findings for vulnerabilities, findings, and risk ingestion",
					Topic:        definitionID.OperationTopic(findingsCollectOperation.Name()),
					ClientRef:    sccClient.ID(),
					ConfigSchema: findingsCollectSchema,
					Policy:       types.ExecutionPolicy{Reconcile: true},
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
						},
						{
							Schema: integrationgenerated.IntegrationMappingSchemaFinding,
						},
						{
							Schema: integrationgenerated.IntegrationMappingSchemaRisk,
						},
					},
					IngestHandle:        FindingsCollect{}.IngestHandle(),
					RequiredPermissions: []string{"https://www.googleapis.com/auth/cloud-platform"},
				},
			},
			Mappings: gcpsccMappings(),
		}, nil
	})
}
