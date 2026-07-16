package gcpscc

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprFinding is the CEL mapping expression for GCP Security Command Center finding payloads
var mapExprFinding = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyFindingExternalID, Expr: `'finding_id' in payload ? payload.finding_id : 'name' in payload ? payload.name : ""`},
	{Key: entityops.InputKeyFindingExternalOwnerID, Expr: "resource"},
	{Key: entityops.InputKeyFindingCategory, Expr: `'category' in payload ? payload.category : ""`},
	{Key: entityops.InputKeyFindingFindingClass, Expr: `'finding_class' in payload ? payload.finding_class : ""`},
	{Key: entityops.InputKeyFindingFindingStatusName, Expr: `'state' in payload && payload.state == "ACTIVE" ? "Open" : "Closed"`},
	{Key: entityops.InputKeyFindingOpen, Expr: `'state' in payload && payload.state == "ACTIVE" ? true : false`},
	{Key: entityops.InputKeyFindingSeverity, Expr: `'severity' in payload ? payload.severity : ""`},
	{Key: entityops.InputKeyFindingDescription, Expr: `'description' in payload ? payload.description : ""`},
	{Key: entityops.InputKeyFindingDisplayName, Expr: `'category' in payload && payload.category != "" ? payload.category : ('name' in payload ? payload.name : "")`},
	{Key: entityops.InputKeyFindingExternalURI, Expr: `'external_uri' in payload ? payload.external_uri : ""`},
	{Key: entityops.InputKeyFindingReportedAt, Expr: `'create_time' in payload ? payload.create_time : null`},
	{Key: entityops.InputKeyFindingSourceUpdatedAt, Expr: `'event_time' in payload ? payload.event_time : ('create_time' in payload ? payload.create_time : null)`},
	{Key: entityops.InputKeyFindingRawPayload, Expr: "payload"},
	{Key: entityops.InputKeyFindingResourceName, Expr: `'name' in payload ? payload.name : ""`},
	{Key: entityops.InputKeyFindingState, Expr: `'state' in payload ? payload.state : ""`},
	{Key: entityops.InputKeyFindingRecommendedActions, Expr: `'source_properties' in payload && 'recommendation' in payload.source_properties ? payload.source_properties.recommendation : ""`},
	{Key: entityops.InputKeyFindingStepsToReproduce, Expr: `'source_properties' in payload && 'explanation' in payload.source_properties ? [payload.source_properties.explanation] : []`},

	// Determine the resource the finding is related to
	{Key: entityops.InputKeyFindingTargets, Expr: `'resource_name' in payload ? [payload.resource_name] : []`},

	// determine what asset the finding is related to
	{Key: entityops.InputKeyFindingTargetDetails, Expr: `
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
	{Key: entityops.InputKeyFindingReferences, Expr: `
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
	{Key: entityops.InputKeyFindingExploitability, Expr: `'attack_exposure_score' in payload ? payload.attack_exposure_score : null`},
	{Key: entityops.InputKeyFindingRecommendedActions, Expr: `'next_steps' in payload ? payload.next_steps : ""`},
})

// mapExprVuln is the CEL mapping expression for GCP Security Command Center vuln payloads
var mapExprVuln = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	// use CVE ID -> fall back to category -> name
	{Key: entityops.InputKeyVulnerabilityDisplayName, Expr: `'vulnerability' in payload && payload.vulnerability != null && 'cve' in payload.vulnerability && payload.vulnerability.cve != null && 'id' in payload.vulnerability.cve && payload.vulnerability.cve.id != "" ? payload.vulnerability.cve.id : ('category' in payload && payload.category != "" ? payload.category : ('name' in payload ? payload.name : ""))`},
	{Key: entityops.InputKeyVulnerabilityCveID, Expr: `'vulnerability' in payload && payload.vulnerability != null && 'cve' in payload.vulnerability && payload.vulnerability.cve != null && 'id' in payload.vulnerability.cve && payload.vulnerability.cve.id != "" ? payload.vulnerability.cve.id : ""`},

	{Key: entityops.InputKeyVulnerabilityExternalID, Expr: `'finding_id' in payload ? payload.finding_id : 'name' in payload ? payload.name : ""`},
	{Key: entityops.InputKeyVulnerabilityExternalOwnerID, Expr: "resource"},
	{Key: entityops.InputKeyVulnerabilityCategory, Expr: `'category' in payload ? payload.category : ""`},
	{Key: entityops.InputKeyVulnerabilityVulnerabilityStatusName, Expr: `'state' in payload && payload.state == "ACTIVE" ? "Open" : "Closed"`},
	{Key: entityops.InputKeyVulnerabilityOpen, Expr: `'state' in payload && payload.state == "ACTIVE" ? true : false`},
	{Key: entityops.InputKeyVulnerabilitySeverity, Expr: `'severity' in payload ? payload.severity : ""`},
	{Key: entityops.InputKeyVulnerabilitySummary, Expr: `'description' in payload && payload.description != "" ? payload.description : ""`},
	{Key: entityops.InputKeyVulnerabilityDescription, Expr: `'description' in payload ? payload.description : ""`},
	{Key: entityops.InputKeyVulnerabilityExternalURI, Expr: `'external_uri' in payload ? payload.external_uri : ""`},
	{Key: entityops.InputKeyVulnerabilityDiscoveredAt, Expr: `'create_time' in payload ? payload.create_time : null`},
	{Key: entityops.InputKeyVulnerabilitySourceUpdatedAt, Expr: `'event_time' in payload ? payload.event_time : ('create_time' in payload ? payload.create_time : null)`},
	{Key: entityops.InputKeyVulnerabilityRawPayload, Expr: "payload"},
	{Key: entityops.InputKeyVulnerabilityFixAvailable, Expr: `'vulnerability' in payload && payload.vulnerability != null && 'fixed_package' in payload.vulnerability && 'package_version' in payload.vulnerability.fixed_package && payload.vulnerability.fixed_package.package_version != "" ? true : false`},
	{Key: entityops.InputKeyVulnerabilityFirstPatchedVersion, Expr: `'vulnerability' in payload && payload.vulnerability != null && 'fixed_package' in payload.vulnerability  && 'package_version' in payload.vulnerability.fixed_package ? payload.vulnerability.fixed_package.package_version : ""`},
	{Key: entityops.InputKeyVulnerabilityVulnerableVersionRange, Expr: `'vulnerability' in payload && payload.vulnerability != null && 'offending_package' in payload.vulnerability && 'package_version' in payload.vulnerability.offending_package ? payload.vulnerability.offending_package.package_version : ""`},
	{Key: entityops.InputKeyVulnerabilityPackageName, Expr: `'vulnerability' in payload && 'offending_package' in payload.vulnerability  && 'package_name' in payload.vulnerability.offending_package ? payload.vulnerability.offending_package.package_name : ""`},
	{Key: entityops.InputKeyVulnerabilityScore, Expr: `'vulnerability' in payload && payload.vulnerability != null && 'cve' in payload.vulnerability && payload.vulnerability.cve != null && 'cvssv3' in payload.vulnerability.cve && payload.vulnerability.cve.cvssv3 != null && 'base_score' in payload.vulnerability.cve.cvssv3 ? payload.vulnerability.cve.cvssv3.base_score : null`},
	{Key: entityops.InputKeyVulnerabilityVector, Expr: `'vulnerability' in payload && payload.vulnerability != null && 'cve' in payload.vulnerability && payload.vulnerability.cve != null && 'cvssv3' in payload.vulnerability.cve && payload.vulnerability.cve.cvssv3 != null && 'attack_vector' in payload.vulnerability.cve.cvssv3 ? payload.vulnerability.cve.cvssv3.attack_vector : null`},
	{Key: entityops.InputKeyVulnerabilityDependencyScope, Expr: `'category' in payload && payload.category  == 'GKE_RUNTIME_OS_VULNERABILITY' ? "RUNTIME" : ""`},
})

// mapExprRisk is the CEL mapping expression for GCP Security Command Center risk payloads
var mapExprRisk = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyRiskExternalID, Expr: `'finding_id' in payload ? payload.finding_id : 'name' in payload ? payload.name : ""`},
	{Key: entityops.InputKeyRiskName, Expr: `'category' in payload ? payload.category : ""`},
	{Key: entityops.InputKeyRiskStatus, Expr: `'state' in payload && payload.state == "ACTIVE" ? "OPEN" : "CLOSED"`},
	{Key: entityops.InputKeyRiskImpact, Expr: `'severity' in payload ? (payload.severity == "MEDIUM" ? "MODERATE" : payload.severity ) : ""`},
	{Key: entityops.InputKeyRiskObservedAt, Expr: `'create_time' in payload ? payload.create_time : ""`},
	{Key: entityops.InputKeyRiskDetails, Expr: `'description' in payload ? payload.description : ""`},
	{Key: entityops.InputKeyRiskRiskCategoryName, Expr: `'finding_class' in payload ? payload.finding_class : ""`},
	{Key: entityops.InputKeyRiskMitigation, Expr: `'next_steps' in payload ? payload.next_steps : ""`},
})

// gcpsccMappings returns the default SCC ingest mappings
func gcpsccMappings() []types.MappingRegistration {
	return []types.MappingRegistration{
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
			},
		},
	}
}
