package gcpscc

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/finding"
	"github.com/theopenlane/core/v2/internal/ent/generated/risk"
	"github.com/theopenlane/core/v2/internal/ent/generated/vulnerability"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// exprActive reports whether SCC considers the finding active
const exprActive = `payload.?state.orValue("") == "ACTIVE"`

// exprMuted reports whether the finding is muted, statically or by a dynamic mute rule currently matching it
const exprMuted = `(payload.?mute.orValue("") == "MUTED" || size(payload.?mute_info.?dynamic_mute_records.orValue([])) > 0)`

// exprOpen reports whether the finding is active and not muted
const exprOpen = exprActive + " && !" + exprMuted

// exprStatusName names the finding status, Dismissed when muted
const exprStatusName = exprMuted + ` ? "Dismissed" : (` + exprActive + ` ? "Open" : "Closed")`

// exprResourcePaths lists the full resource path the record was raised against
const exprResourcePaths = `[payload.?resource_name.orValue("")].filter(n, n != "")`

// exprRiskImpact maps the SCC severity to a risk impact, null when the severity has no impact equivalent
const exprRiskImpact = `{"LOW": "LOW", "MEDIUM": "MODERATE", "HIGH": "HIGH", "CRITICAL": "CRITICAL"}[?payload.?severity.orValue("")].orValue(null)`

// mapExprFinding is the CEL mapping expression for GCP Security Command Center finding payloads
var mapExprFinding = providerkit.CelMapExpr(
	entityops.FindingFields.ExternalID.Expr(`payload.?name.orValue("")`),
	entityops.FindingFields.ExternalOwnerID.Expr("resource"),
	entityops.FindingFields.Category.Expr(`payload.?category.orValue("")`),
	entityops.FindingFields.FindingClass.Expr(`payload.?finding_class.orValue("")`),
	entityops.FindingFields.FindingStatusName.Expr(exprStatusName),
	entityops.FindingFields.Open.Expr(exprOpen),
	entityops.FindingFields.Severity.Expr(`payload.?severity.orValue("")`),
	entityops.FindingFields.Description.Expr(`paragraphs(payload.?description.orValue(""))`),
	entityops.FindingFields.DisplayName.Expr(`payload.?category_display_name.or(payload.?category).orValue(payload.?name.orValue(""))`),
	entityops.FindingFields.ExternalURI.Expr(`payload.?external_uri.orValue("")`),
	entityops.FindingFields.ReportedAt.Expr(`payload.?create_time.orValue(null)`),
	entityops.FindingFields.SourceUpdatedAt.Expr(`payload.?event_time.orValue(payload.?create_time.orValue(null))`),
	entityops.FindingFields.RawPayload.Expr("payload"),
	entityops.FindingFields.ResourceName.Expr(`payload.?resource_name.orValue("")`),
	entityops.FindingFields.State.Expr(`payload.?state.orValue("")`),
	entityops.FindingFields.Recommendation.Expr(`payload.?source_properties.?Recommendation.orValue("")`),
	entityops.FindingFields.StepsToReproduce.Expr(`payload.?source_properties.?Explanation.optMap(e, [e]).orValue([])`),

	// Determine the resource the finding is related to
	entityops.FindingFields.Targets.Expr(exprResourcePaths),

	// determine what asset the finding is related to
	entityops.FindingFields.TargetDetails.Expr(`
  size(payload.?kubernetes.orValue({})) > 0
    ? payload.kubernetes
  : size(payload.?containers.orValue([])) > 0
    ? {"containers": payload.containers}
  : payload.?database.hasValue()
    ? payload.database
  : size(payload.?processes.orValue([])) > 0
    ? {"processes": payload.processes}
  : size(payload.?files.orValue([])) > 0
    ? {"files": payload.files}
  : payload.?application.hasValue()
    ? payload.application
  : payload.?resource_metadata.orValue(null)
`),
	entityops.FindingFields.References.Expr(`
  [
    payload.?source_properties.?contextUris.?mitreUri.?url,
    payload.?source_properties.?contextUris.?relatedFindingUri.?url,
    payload.?source_properties.?contextUris.?virustotalIndicatorQueryUri[?0].?url
  ].filter(u, u.hasValue()).map(u, u.value())
`),
	entityops.FindingFields.Exploitability.Expr(`payload.?attack_exposure.?score.orValue(null)`),
	entityops.FindingFields.RecommendedActions.Expr(`payload.?next_steps.orValue("")`),
	providerkit.EdgeExternalIDEntry(entityops.SchemaFinding, finding.EdgeAssets, exprResourcePaths),
)

// mapExprVuln is the CEL mapping expression for GCP Security Command Center vuln payloads
var mapExprVuln = providerkit.CelMapExpr(
	// use CVE ID -> fall back to category -> name
	entityops.VulnerabilityFields.DisplayName.Expr(`payload.?vulnerability.?cve.?id.or(payload.?category_display_name).or(payload.?category).orValue(payload.?name.orValue(""))`),
	entityops.VulnerabilityFields.CveID.Expr(`payload.?vulnerability.?cve.?id.orValue("")`),

	entityops.VulnerabilityFields.ExternalID.Expr(`payload.?name.orValue("")`),
	entityops.VulnerabilityFields.ExternalOwnerID.Expr("resource"),
	entityops.VulnerabilityFields.Category.Expr(`payload.?category.orValue("")`),
	entityops.VulnerabilityFields.VulnerabilityStatusName.Expr(exprStatusName),
	entityops.VulnerabilityFields.Open.Expr(exprOpen),
	entityops.VulnerabilityFields.DismissedAt.Expr(exprMuted+` ? payload.?mute_update_time.orValue(null) : null`),
	entityops.VulnerabilityFields.DismissedReason.Expr(exprMuted+` ? payload.?mute_initiator.orValue(null) : null`),
	entityops.VulnerabilityFields.Severity.Expr(`payload.?severity.orValue("")`),
	entityops.VulnerabilityFields.Summary.Expr(`payload.?vulnerability_summary.orValue("")`),
	entityops.VulnerabilityFields.Description.Expr(`paragraphs(payload.?description.orValue(""))`),
	entityops.VulnerabilityFields.ExternalURI.Expr(`payload.?external_uri.orValue("")`),
	entityops.VulnerabilityFields.DiscoveredAt.Expr(`payload.?create_time.orValue(null)`),
	entityops.VulnerabilityFields.SourceUpdatedAt.Expr(`payload.?event_time.orValue(payload.?create_time.orValue(null))`),
	entityops.VulnerabilityFields.RawPayload.Expr("payload"),
	entityops.VulnerabilityFields.FixAvailable.Expr(`payload.?vulnerability.?fixed_package.?package_version.orValue("") != ""`),
	entityops.VulnerabilityFields.FirstPatchedVersion.Expr(`payload.?vulnerability.?fixed_package.?package_version.orValue("")`),
	entityops.VulnerabilityFields.VulnerableVersionRange.Expr(`payload.?vulnerability.?offending_package.?package_version.orValue("")`),
	entityops.VulnerabilityFields.PackageName.Expr(`payload.?vulnerability.?offending_package.?package_name.orValue("")`),
	entityops.VulnerabilityFields.Score.Expr(`payload.?vulnerability.?cve.?cvssv3.?base_score.orValue(null)`),
	entityops.VulnerabilityFields.Vector.Expr(`payload.?cvss_vector.orValue("")`),
	entityops.VulnerabilityFields.DependencyScope.Expr(`payload.?category.orValue("") == "GKE_RUNTIME_OS_VULNERABILITY" ? "RUNTIME" : ""`),
	entityops.VulnerabilityFields.Source.Expr(providerkit.ExprInstallationName),
	providerkit.EdgeExternalIDEntry(entityops.SchemaVulnerability, vulnerability.EdgeAssets, exprResourcePaths),
)

// mapExprRisk is the CEL mapping expression for GCP Security Command Center risk payloads
var mapExprRisk = providerkit.CelMapExpr(
	entityops.RiskFields.ExternalID.Expr(`payload.?name.orValue("")`),
	entityops.RiskFields.Name.Expr(`payload.?category_display_name.orValue(payload.?category.orValue(""))`),
	entityops.RiskFields.Status.Expr(exprMuted+` ? "ACCEPTED" : (`+exprActive+` ? "OPEN" : "CLOSED")`),
	entityops.RiskFields.Impact.Expr(exprRiskImpact),
	entityops.RiskFields.ObservedAt.Expr(`payload.?create_time.orValue(null)`),
	entityops.RiskFields.Details.Expr(`payload.?description.orValue("")`),
	entityops.RiskFields.RiskCategoryName.Expr(`payload.?finding_class.orValue("")`),
	entityops.RiskFields.Mitigation.Expr(`payload.?next_steps.orValue("")`),
	providerkit.EdgeExternalIDEntry(entityops.SchemaRisk, risk.EdgeAssets, exprResourcePaths),
)
