package gcpscc

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprFinding is the CEL mapping expression for GCP Security Command Center finding payloads
var mapExprFinding = providerkit.CelMapExpr(
	entityops.FindingFields.ExternalID.Expr(`'finding_id' in payload ? payload.finding_id : 'name' in payload ? payload.name : ""`),
	entityops.FindingFields.ExternalOwnerID.Expr("resource"),
	entityops.FindingFields.Category.Expr(`'category' in payload ? payload.category : ""`),
	entityops.FindingFields.FindingClass.Expr(`'finding_class' in payload ? payload.finding_class : ""`),
	entityops.FindingFields.FindingStatusName.Expr(`'state' in payload && payload.state == "ACTIVE" ? "Open" : "Closed"`),
	entityops.FindingFields.Open.Expr(`'state' in payload && payload.state == "ACTIVE" ? true : false`),
	entityops.FindingFields.Severity.Expr(`'severity' in payload ? payload.severity : ""`),
	entityops.FindingFields.Description.Expr(`paragraphs('description' in payload ? payload.description : "")`),
	entityops.FindingFields.DisplayName.Expr(`'category' in payload && payload.category != "" ? payload.category : ('name' in payload ? payload.name : "")`),
	entityops.FindingFields.ExternalURI.Expr(`'external_uri' in payload ? payload.external_uri : ""`),
	entityops.FindingFields.ReportedAt.Expr(`'create_time' in payload ? payload.create_time : null`),
	entityops.FindingFields.SourceUpdatedAt.Expr(`'event_time' in payload ? payload.event_time : ('create_time' in payload ? payload.create_time : null)`),
	entityops.FindingFields.RawPayload.Expr("payload"),
	entityops.FindingFields.ResourceName.Expr(`'name' in payload ? payload.name : ""`),
	entityops.FindingFields.State.Expr(`'state' in payload ? payload.state : ""`),
	entityops.FindingFields.RecommendedActions.Expr(`'source_properties' in payload && 'recommendation' in payload.source_properties ? payload.source_properties.recommendation : ""`),
	entityops.FindingFields.StepsToReproduce.Expr(`'source_properties' in payload && 'explanation' in payload.source_properties ? [payload.source_properties.explanation] : []`),

	// Determine the resource the finding is related to
	entityops.FindingFields.Targets.Expr(`'resource_name' in payload ? [payload.resource_name] : []`),

	// determine what asset the finding is related to
	entityops.FindingFields.TargetDetails.Expr(`
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
`),
	entityops.FindingFields.References.Expr(`
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
`),
	entityops.FindingFields.Exploitability.Expr(`'attack_exposure_score' in payload ? payload.attack_exposure_score : null`),
	entityops.FindingFields.RecommendedActions.Expr(`'next_steps' in payload ? payload.next_steps : ""`),
)

// mapExprVuln is the CEL mapping expression for GCP Security Command Center vuln payloads
var mapExprVuln = providerkit.CelMapExpr(
	// use CVE ID -> fall back to category -> name
	entityops.VulnerabilityFields.DisplayName.Expr(`'vulnerability' in payload && payload.vulnerability != null && 'cve' in payload.vulnerability && payload.vulnerability.cve != null && 'id' in payload.vulnerability.cve && payload.vulnerability.cve.id != "" ? payload.vulnerability.cve.id : ('category' in payload && payload.category != "" ? payload.category : ('name' in payload ? payload.name : ""))`),
	entityops.VulnerabilityFields.CveID.Expr(`'vulnerability' in payload && payload.vulnerability != null && 'cve' in payload.vulnerability && payload.vulnerability.cve != null && 'id' in payload.vulnerability.cve && payload.vulnerability.cve.id != "" ? payload.vulnerability.cve.id : ""`),

	entityops.VulnerabilityFields.ExternalID.Expr(`'finding_id' in payload ? payload.finding_id : 'name' in payload ? payload.name : ""`),
	entityops.VulnerabilityFields.ExternalOwnerID.Expr("resource"),
	entityops.VulnerabilityFields.Category.Expr(`'category' in payload ? payload.category : ""`),
	entityops.VulnerabilityFields.VulnerabilityStatusName.Expr(`'state' in payload && payload.state == "ACTIVE" ? "Open" : "Closed"`),
	entityops.VulnerabilityFields.Open.Expr(`'state' in payload && payload.state == "ACTIVE" ? true : false`),
	entityops.VulnerabilityFields.Severity.Expr(`'severity' in payload ? payload.severity : ""`),
	entityops.VulnerabilityFields.Summary.Expr(`'description' in payload && payload.description != "" ? payload.description : ""`),
	entityops.VulnerabilityFields.Description.Expr(`paragraphs('description' in payload ? payload.description : "")`),
	entityops.VulnerabilityFields.ExternalURI.Expr(`'external_uri' in payload ? payload.external_uri : ""`),
	entityops.VulnerabilityFields.DiscoveredAt.Expr(`'create_time' in payload ? payload.create_time : null`),
	entityops.VulnerabilityFields.SourceUpdatedAt.Expr(`'event_time' in payload ? payload.event_time : ('create_time' in payload ? payload.create_time : null)`),
	entityops.VulnerabilityFields.RawPayload.Expr("payload"),
	entityops.VulnerabilityFields.FixAvailable.Expr(`'vulnerability' in payload && payload.vulnerability != null && 'fixed_package' in payload.vulnerability && 'package_version' in payload.vulnerability.fixed_package && payload.vulnerability.fixed_package.package_version != "" ? true : false`),
	entityops.VulnerabilityFields.FirstPatchedVersion.Expr(`'vulnerability' in payload && payload.vulnerability != null && 'fixed_package' in payload.vulnerability  && 'package_version' in payload.vulnerability.fixed_package ? payload.vulnerability.fixed_package.package_version : ""`),
	entityops.VulnerabilityFields.VulnerableVersionRange.Expr(`'vulnerability' in payload && payload.vulnerability != null && 'offending_package' in payload.vulnerability && 'package_version' in payload.vulnerability.offending_package ? payload.vulnerability.offending_package.package_version : ""`),
	entityops.VulnerabilityFields.PackageName.Expr(`'vulnerability' in payload && 'offending_package' in payload.vulnerability  && 'package_name' in payload.vulnerability.offending_package ? payload.vulnerability.offending_package.package_name : ""`),
	entityops.VulnerabilityFields.Score.Expr(`'vulnerability' in payload && payload.vulnerability != null && 'cve' in payload.vulnerability && payload.vulnerability.cve != null && 'cvssv3' in payload.vulnerability.cve && payload.vulnerability.cve.cvssv3 != null && 'base_score' in payload.vulnerability.cve.cvssv3 ? payload.vulnerability.cve.cvssv3.base_score : null`),
	entityops.VulnerabilityFields.Vector.Expr(`'vulnerability' in payload && payload.vulnerability != null && 'cve' in payload.vulnerability && payload.vulnerability.cve != null && 'cvssv3' in payload.vulnerability.cve && payload.vulnerability.cve.cvssv3 != null && 'attack_vector' in payload.vulnerability.cve.cvssv3 ? payload.vulnerability.cve.cvssv3.attack_vector : null`),
	entityops.VulnerabilityFields.DependencyScope.Expr(`'category' in payload && payload.category  == 'GKE_RUNTIME_OS_VULNERABILITY' ? "RUNTIME" : ""`),
	entityops.VulnerabilityFields.Source.Expr(providerkit.ExprInstallationName),
)

// mapExprRisk is the CEL mapping expression for GCP Security Command Center risk payloads
var mapExprRisk = providerkit.CelMapExpr(
	entityops.RiskFields.ExternalID.Expr(`'finding_id' in payload ? payload.finding_id : 'name' in payload ? payload.name : ""`),
	entityops.RiskFields.Name.Expr(`'category' in payload ? payload.category : ""`),
	entityops.RiskFields.Status.Expr(`'state' in payload && payload.state == "ACTIVE" ? "OPEN" : "CLOSED"`),
	entityops.RiskFields.Impact.Expr(`'severity' in payload ? (payload.severity == "MEDIUM" ? "MODERATE" : payload.severity ) : ""`),
	entityops.RiskFields.ObservedAt.Expr(`'create_time' in payload ? payload.create_time : ""`),
	entityops.RiskFields.Details.Expr(`'description' in payload ? payload.description : ""`),
	entityops.RiskFields.RiskCategoryName.Expr(`'finding_class' in payload ? payload.finding_class : ""`),
	entityops.RiskFields.Mitigation.Expr(`'next_steps' in payload ? payload.next_steps : ""`),
)
