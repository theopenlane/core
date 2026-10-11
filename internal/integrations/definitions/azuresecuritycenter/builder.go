package azuresecuritycenter

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// Builder returns the Azure Security Center definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			ID:           definitionID.ID(),
			Family:       "Azure",
			DisplayName:  "Microsoft Defender for Cloud",
			Description:  "Collect security assessment findings and vulnerability data from Microsoft Defender for Cloud across an Azure subscription.",
			Category:     "security-posture",
			DocsURL:      "https://docs.theopenlane.io/docs/platform/integrations/azure_security_center",
			Tags:         []string{"vulnerabilities", "assets"},
			Active:       false,
			Visible:      true,
			Installation: installation.Registration(),
			Connections: []types.Connector{
				securityCenterConnection.
					Name("Azure Service Principal").
					Description("Configure Defender for Cloud access using an Azure service principal.").
					Provides(buildClient).
					Verified(verify).
					Disconnects("Removes the stored service principal credentials from Openlane. If the Azure app registration is no longer needed, delete it from your Azure tenant.", nil),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[AssessmentsCollect]().
					Ingests(runAssessmentsCollect).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaVulnerability.Name}).
					Description("Collect unhealthy security posture assessment findings for vulnerability ingestion").
					Registration(),
				types.OperationRefOf[SubAssessmentsCollect]().
					Ingests(runSubAssessmentsCollect).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaVulnerability.Name}).
					Description("Collect granular sub-assessment vulnerability findings (CVEs from container images, servers, and SQL checks)").
					Registration(),
			},
			Mappings: []types.MappingRegistration{
				{
					Schema:  entityops.SchemaVulnerability.Name,
					Variant: variantAssessment,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprAssessment,
					},
				},
				{
					Schema:  entityops.SchemaVulnerability.Name,
					Variant: variantSubAssessment,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprSubAssessment,
					},
				},
			},
		}, nil
	})
}
