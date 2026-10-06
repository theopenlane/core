package gcpscc

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// Builder returns the GCP SCC definition builder advertising the supplied federation issuer
func Builder(federationIssuer string) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			ID:           definitionID.ID(),
			Family:       "Google Cloud",
			DisplayName:  "GCP Security Command Center",
			Description:  "Collect Google Cloud Security Command Center findings for security posture reporting.",
			Category:     "security-posture",
			DocsURL:      "https://docs.theopenlane.io/docs/platform/integrations/gcp-scc",
			Tags:         []string{"vulnerabilities", "assets", "findings", "risks"},
			Active:       true,
			Visible:      true,
			HealthCheck:  sccClient.HealthCheck(checkHealth),
			Installation: installation.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				workloadIdentityCredential.Registration(types.CredentialRegistration{
					Name:        "GCP Workload Identity Federation",
					Description: "Federated access to Security Command Center with no stored keys.",
					Recommended: true,
				}),
				sccCredential.Registration(types.CredentialRegistration{
					Name:        "GCP SCC Credential",
					Description: "GCP service account key used to access Security Command Center.",
				}),
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef: workloadIdentityCredential.ID(),
					Name:          "GCP Workload Identity Federation",
					Description:   "Configure Security Command Center access by trusting Openlane as an OIDC identity provider, so no service account key is ever stored.",
					Meta: map[string]types.MetaInfo{
						"Openlane Issuer URI": {
							Value:     federationIssuer,
							AllowCopy: true,
						},
					},
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: workloadIdentityCredential.ID(),
						Description:   "Removes the stored workload identity configuration from Openlane. If the workload identity pool is no longer needed, delete the pool and its provider from your Google Cloud project.",
					},
				},
				{
					CredentialRef: sccCredential.ID(),
					Name:          "GCP Service Account",
					Description:   "Configure Security Command Center access using a GCP service account.",
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: sccCredential.ID(),
						Description:   "Removes the stored service account credentials from Openlane. If the GCP service account is no longer needed, delete it from your Google Cloud project.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				sccClient.Registration(Client{}.Build, types.ClientRegistration{
					Description: "Google Cloud Security Command Center v2 client",
				}),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[FindingsSync]().
					Ingests(sccClient, runFindingsCollect).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(
						types.IngestContract{Schema: entityops.SchemaVulnerability.Name},
						types.IngestContract{Schema: entityops.SchemaFinding.Name},
						types.IngestContract{Schema: entityops.SchemaRisk.Name},
					).
					Permissions("https://www.googleapis.com/auth/cloud-platform").
					Description("Collect GCP Security Command Center findings for vulnerabilities, findings, and risk ingestion").
					Registration(definitionID),
			},
			Mappings: []types.MappingRegistration{
				{
					Schema: entityops.SchemaRisk.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprRisk,
					},
				},
				{
					Schema: entityops.SchemaVulnerability.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprVuln,
					},
				},
				providerkit.FindingMapping(mapExprFinding),
			},
		}, nil
	})
}
