package gcpscc

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// Builder returns the GCP SCC definition builder advertising the supplied federation issuer
func Builder(federationIssuer string) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "Google Cloud",
				DisplayName: "GCP Security Command Center",
				Description: "Collect Google Cloud Security Command Center findings for security posture reporting.",
				Category:    "security-posture",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/gcp-scc",
				Tags:        []string{"vulnerabilities", "assets", "findings", "risks"},
				Active:      true,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: jsonx.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         workloadIdentityCredential.ID(),
					Name:        "GCP Workload Identity Federation",
					Description: "Federated access to Security Command Center with no stored keys.",
					Schema:      workloadIdentitySchema,
					Recommended: true,
				},
				{
					Ref:         sccCredential.ID(),
					Name:        "GCP SCC Credential",
					Description: "GCP service account key used to access Security Command Center.",
					Schema:      sccSchema,
				},
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
					CredentialRefs:      []types.CredentialSlotID{workloadIdentityCredential.ID()},
					ClientRefs:          []types.ClientID{sccClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: workloadIdentityCredential.ID(),
						Description:   "Removes the stored workload identity configuration from Openlane. If the workload identity pool is no longer needed, delete the pool and its provider from your Google Cloud project.",
					},
				},
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
					CredentialRefs: []types.CredentialSlotID{workloadIdentityCredential.ID(), sccCredential.ID()},
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
							Schema: entityops.SchemaVulnerability.Name,
						},
						{
							Schema: entityops.SchemaFinding.Name,
						},
						{
							Schema: entityops.SchemaRisk.Name,
						},
					},
					IngestHandle:        FindingsCollect{}.IngestHandle(),
					RequiredPermissions: []string{"https://www.googleapis.com/auth/cloud-platform"},
					ConfigResolver:      providerkit.ConfigFrom(func(u UserInput) FindingsSyncConfig { return u.FindingsSync }),
				},
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
				{
					Schema: entityops.SchemaFinding.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprFinding,
						Links: []types.LinkRule{
							{
								TargetSchema: entityops.SchemaControl.Name,
								TargetField:  control.FieldRefCode,
								SourceField:  entityops.InputKeyFindingCategory,
								SourceList:   entityops.InputKeyFindingCategories,
							},
						},
					},
				},
			},
		}, nil
	})
}
