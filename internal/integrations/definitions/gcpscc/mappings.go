package gcpscc

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprFinding is the CEL mapping expression for GCP Security Command Center finding payloads
var mapExprFinding = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingFindingExternalID, Expr: `'finding_id' in payload ? payload.finding_id : 'name' in payload ? payload.name : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingExternalOwnerID, Expr: "resource"},
	{Key: integrationgenerated.IntegrationMappingFindingCategory, Expr: `'category' in payload ? payload.category : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingFindingClass, Expr: `'finding_class' in payload ? payload.finding_class : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingFindingStatusName, Expr: `'state' in payload && payload.state == "ACTIVE" ? "Open" : "Closed"`},
	{Key: integrationgenerated.IntegrationMappingFindingOpen, Expr: `'state' in payload && payload.state == "ACTIVE" ? true : false`},
	{Key: integrationgenerated.IntegrationMappingFindingSeverity, Expr: `'severity' in payload ? payload.severity : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingDescription, Expr: `'description' in payload ? payload.description : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingDisplayName, Expr: `'category' in payload && payload.category != "" ? payload.category : ('name' in payload ? payload.name : "")`},
	{Key: integrationgenerated.IntegrationMappingFindingExternalURI, Expr: `'external_uri' in payload ? payload.external_uri : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingReportedAt, Expr: `'create_time' in payload ? payload.create_time : null`},
	{Key: integrationgenerated.IntegrationMappingFindingSourceUpdatedAt, Expr: `'event_time' in payload ? payload.event_time : ('create_time' in payload ? payload.create_time : null)`},
	{Key: integrationgenerated.IntegrationMappingFindingRawPayload, Expr: "payload"},
	{Key: integrationgenerated.IntegrationMappingFindingResourceName, Expr: `'name' in payload ? payload.name : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingState, Expr: `'state' in payload ? payload.state : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingRecommendedActions, Expr: `'source_properties' in payload && 'recommendation' in payload.source_properties ? payload.source_properties.recommendation : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingStepsToReproduce, Expr: `'source_properties' in payload && 'explanation' in payload.source_properties ? [payload.source_properties.explanation] : []`},

	// Determine the resource the finding is related to
	{Key: integrationgenerated.IntegrationMappingFindingTargets, Expr: `'resource_name' in payload ? [payload.resource_name] : []`},

	// determine what asset the finding is related to
	{Key: integrationgenerated.IntegrationMappingFindingTargetDetails, Expr: `
  'kubernetes' in payload && size(payload.kubernetes) > 0
    ? payload.kubernetes
  : 'containers' in payload && size(payload.containers) > 0
    ? {"containers": payload.containers}
  : 'database' in payload && payload.database != null
    ? payload.database
  : 'processes' in payload && size(payload.processes) > 0
    ? {"processes": payload.processes}
  : 'files' in payload && size(payload.files) > 0
    ? {"files": payload.files}
  : 'application' in payload && payload.application != null
    ? payload.application
  : null
`},
	{Key: integrationgenerated.IntegrationMappingFindingReferences, Expr: `
  'contextUris' in payload ?
    [
      ('mitreUri' in payload.contextUris && 'url' in payload.contextUris.mitreUri)
        ? payload.contextUris.mitreUri.url : null,
      ('relatedFindingUri' in payload.contextUris && 'url' in payload.contextUris.relatedFindingUri)
        ? payload.contextUris.relatedFindingUri.url : null,
      ('virustotalIndicatorQueryUri' in payload.contextUris && size(payload.contextUris.virustotalIndicatorQueryUri) > 0)
        ? payload.contextUris.virustotalIndicatorQueryUri[0].url : null
    ].filter(u, u != null)
  : []
`},
	{Key: integrationgenerated.IntegrationMappingFindingExploitability, Expr: `'attack_exposure_score' in payload ? payload.attack_exposure_score : null`},
	{Key: integrationgenerated.IntegrationMappingFindingRecommendedActions, Expr: `'next_steps' in payload ? payload.next_steps : ""`},
})

// mapExprVuln is the CEL mapping expression for GCP Security Command Center vuln payloads
var mapExprVuln = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	// use CVE ID -> fall back to category -> name
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDisplayName, Expr: `'vulnerability' in payload && payload.vulnerability != null && 'cve' in payload.vulnerability && payload.vulnerability.cve != null && 'id' in payload.vulnerability.cve && payload.vulnerability.cve.id != "" ? payload.vulnerability.cve.id : ('category' in payload && payload.category != "" ? payload.category : ('name' in payload ? payload.name : ""))`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityCveID, Expr: `'vulnerability' in payload && payload.vulnerability != null && 'cve' in payload.vulnerability && payload.vulnerability.cve != null && 'id' in payload.vulnerability.cve && payload.vulnerability.cve.id != "" ? payload.vulnerability.cve.id : ""`},

	{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalID, Expr: `'finding_id' in payload ? payload.finding_id : 'name' in payload ? payload.name : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalOwnerID, Expr: "resource"},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityCategory, Expr: `'category' in payload ? payload.category : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityVulnerabilityStatusName, Expr: `'state' in payload && payload.state == "ACTIVE" ? "Open" : "Closed"`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityOpen, Expr: `'state' in payload && payload.state == "ACTIVE" ? true : false`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, Expr: `'severity' in payload ? payload.severity : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySummary, Expr: `'description' in payload && payload.description != "" ? payload.description : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDescription, Expr: `'description' in payload ? payload.description : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalURI, Expr: `'external_uri' in payload ? payload.external_uri : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt, Expr: `'create_time' in payload ? payload.create_time : null`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt, Expr: `'event_time' in payload ? payload.event_time : ('create_time' in payload ? payload.create_time : null)`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityRawPayload, Expr: "payload"},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityFixAvailable, Expr: `'vulnerability' in payload && payload.vulnerability != null && 'fixed_package' in payload.vulnerability && 'package_version' in payload.vulnerability.fixed_package && payload.vulnerability.fixed_package.package_version != "" ? true : false`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityFirstPatchedVersion, Expr: `'vulnerability' in payload && payload.vulnerability != null && 'fixed_package' in payload.vulnerability  && 'package_version' in payload.vulnerability.fixed_package ? payload.vulnerability.fixed_package.package_version : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityVulnerableVersionRange, Expr: `'vulnerability' in payload && payload.vulnerability != null && 'offending_package' in payload.vulnerability && 'package_version' in payload.vulnerability.offending_package ? payload.vulnerability.offending_package.package_version : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityPackageName, Expr: `'vulnerability' in payload && 'offending_package' in payload.vulnerability  && 'package_name' in payload.vulnerability.offending_package ? payload.vulnerability.offending_package.package_name : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityScore, Expr: `'vulnerability' in payload && payload.vulnerability != null && 'cve' in payload.vulnerability && payload.vulnerability.cve != null && 'cvssv3' in payload.vulnerability.cve && payload.vulnerability.cve.cvssv3 != null && 'base_score' in payload.vulnerability.cve.cvssv3 ? payload.vulnerability.cve.cvssv3.base_score : null`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityVector, Expr: `'vulnerability' in payload && payload.vulnerability != null && 'cve' in payload.vulnerability && payload.vulnerability.cve != null && 'cvssv3' in payload.vulnerability.cve && payload.vulnerability.cve.cvssv3 != null && 'attack_vector' in payload.vulnerability.cve.cvssv3 ? payload.vulnerability.cve.cvssv3.attack_vector : null`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDependencyScope, Expr: `'category' in payload && payload.category  == 'GKE_RUNTIME_OS_VULNERABILITY' ? "RUNTIME" : ""`},
})

// mapExprRisk is the CEL mapping expression for GCP Security Command Center risk payloads
var mapExprRisk = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingRiskExternalID, Expr: `'finding_id' in payload ? payload.finding_id : 'name' in payload ? payload.name : ""`},
	{Key: integrationgenerated.IntegrationMappingRiskName, Expr: `'category' in payload ? payload.category : ""`},
	{Key: integrationgenerated.IntegrationMappingRiskStatus, Expr: `'state' in payload && payload.state == "ACTIVE" ? "OPEN" : "CLOSED"`},
	{Key: integrationgenerated.IntegrationMappingRiskImpact, Expr: `'severity' in payload ? (payload.severity == "MEDIUM" ? "MODERATE" : payload.severity ) : ""`},
	{Key: integrationgenerated.IntegrationMappingRiskObservedAt, Expr: `'create_time' in payload ? payload.create_time : ""`},
	{Key: integrationgenerated.IntegrationMappingRiskDetails, Expr: `'description' in payload ? payload.description : ""`},
	{Key: integrationgenerated.IntegrationMappingRiskRiskCategoryName, Expr: `'finding_class' in payload ? payload.finding_class : ""`},
	{Key: integrationgenerated.IntegrationMappingRiskMitigation, Expr: `'next_steps' in payload ? payload.next_steps : ""`},
})

// gcpsccMappings returns the default SCC ingest mappings
func gcpsccMappings() []types.MappingRegistration {
	return []types.MappingRegistration{
		{
			Schema: integrationgenerated.IntegrationMappingSchemaRisk,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprRisk,
			},
		},
		{
			Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprVuln,
			},
		},
		{
			Schema: integrationgenerated.IntegrationMappingSchemaFinding,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprFinding,
			},
		},
	}
}
