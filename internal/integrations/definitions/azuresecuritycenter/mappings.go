package azuresecuritycenter

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprFinding maps AssessmentPayload fields to the Finding schema.
//
// Assessments are security posture policy checks (misconfigurations), not CVE vulnerabilities.
// The ARM assessment resource ID is unique per (resource, policy) and serves as the upsert key.
// Failed compliance assessments are converted to the same AssessmentPayload, keyed by the compliance assessment ID.
var mapExprFinding = providerkit.CelMapExpr(
	entityops.FindingFields.ExternalID.Expr(`payload.?id.orValue("")`),
	entityops.FindingFields.ExternalOwnerID.Expr(`payload.?governance_owner.orValue("")`),
	entityops.FindingFields.InternalOwner.Expr(`payload.?governance_owner.orValue(null)`),
	entityops.FindingFields.DisplayName.Expr(`payload.?display_name.orValue("")`),
	entityops.FindingFields.Description.Expr(`paragraphs(payload.?description.orValue(""))`),
	entityops.FindingFields.Severity.Expr(`payload.?severity.orValue("")`),
	entityops.FindingFields.Category.Expr(`payload.?category.orValue("")`),
	entityops.FindingFields.Categories.Expr(`payload.?categories.orValue([])`),
	entityops.FindingFields.Tags.Expr(`payload.?standards.orValue([])`),
	entityops.FindingFields.FindingStatusName.Expr(`payload.?status_code.orValue("")`),
	entityops.FindingFields.Open.Expr(`payload.?status_code.orValue("") == "Unhealthy"`),
	entityops.FindingFields.ExternalURI.Expr(`payload.?external_uri.orValue("")`),
	entityops.FindingFields.RecommendedActions.Expr(`payload.?remediation.orValue("")`),
	entityops.FindingFields.Targets.Expr(`payload.?resources.orValue([])`),
	entityops.FindingFields.ResourceName.Expr(`payload.?resource_id.orValue("")`),
	entityops.FindingFields.TargetDetails.Expr(`payload.?affected_resources.optMap(r, dyn({'resources': r})).orValue(null)`),
	entityops.FindingFields.Metadata.Expr(`payload.?metadata.orValue(null)`),
	entityops.FindingFields.ReportedAt.Expr(`payload.?discovered_at.orValue(null)`),
	entityops.FindingFields.SourceUpdatedAt.Expr(`payload.?updated_at.orValue(null)`),
	entityops.FindingFields.RawPayload.Expr("payload"),
	entityops.FindingFields.Source.Expr(providerkit.ExprInstallationName),
)

// mapExprCheckResult maps CompliancePayload fields to the CheckResult schema; the control ID in tags links the check to its control
var mapExprCheckResult = providerkit.CelMapExpr(
	entityops.CheckResultFields.ParentExternalID.Expr(`payload.?id.orValue("")`),
	entityops.CheckResultFields.Status.Expr(`payload.?status.orValue("UNKNOWN")`),
	entityops.CheckResultFields.Details.Expr(`payload.?display_name.orValue("")`),
	entityops.CheckResultFields.ExternalURI.Expr(`payload.?external_uri.orValue("")`),
	entityops.CheckResultFields.LastObservedAt.Expr(`payload.?observed_at.orValue(null)`),
	entityops.CheckResultFields.Tags.Expr(`[payload.standard, payload.control].filter(t, t != "")`),
	entityops.CheckResultFields.Source.Expr(providerkit.ExprInstallationName),
)

// mapExprVulnerability maps AssessmentPayload fields to the Vulnerability schema.
//
// Each software update or service upgrade recommendation is split into one record per CVE.
// The recommendation ID suffixed with the CVE ID serves as the upsert key.
var mapExprVulnerability = providerkit.CelMapExpr(
	entityops.VulnerabilityFields.ExternalID.Expr(`payload.?id.orValue("")`),
	entityops.VulnerabilityFields.ExternalOwnerID.Expr(`payload.?governance_owner.orValue("")`),
	entityops.VulnerabilityFields.InternalOwner.Expr(`payload.?governance_owner.orValue(null)`),
	entityops.VulnerabilityFields.Metadata.Expr(`{'resource_details': payload.?resource_details.orValue(null), 'detected_versions': payload.?detected_versions.orValue([]), 'recommendation': payload.?metadata.orValue(null)}`),
	entityops.VulnerabilityFields.PackageName.Expr(`payload.?package_name.orValue("")`),
	entityops.VulnerabilityFields.PackageEcosystem.Expr(`payload.?package_ecosystem.orValue("")`),
	entityops.VulnerabilityFields.Score.Expr(`payload.?cvss_score.orValue(null)`),
	entityops.VulnerabilityFields.CveID.Expr(`payload.?cve_id.orValue("")`),
	entityops.VulnerabilityFields.DisplayName.Expr(`payload.?cve_id.orValue("") != "" ? payload.cve_id : payload.?display_name.orValue("")`),
	entityops.VulnerabilityFields.Summary.Expr(`payload.?display_name.orValue("")`),
	entityops.VulnerabilityFields.Description.Expr(`paragraphs(payload.?description.orValue(""))`),
	entityops.VulnerabilityFields.Severity.Expr(`payload.?severity.orValue("")`),
	entityops.VulnerabilityFields.Category.Expr(`payload.?category.orValue("")`),
	entityops.VulnerabilityFields.VulnerabilityStatusName.Expr(`payload.?status_code.orValue("")`),
	entityops.VulnerabilityFields.Open.Expr(`payload.?status_code.orValue("") == "Unhealthy"`),
	entityops.VulnerabilityFields.ExternalURI.Expr(`payload.?external_uri.orValue("")`),
	entityops.VulnerabilityFields.FirstPatchedVersion.Expr(`payload.?fixed_version.orValue("")`),
	entityops.VulnerabilityFields.FixAvailable.Expr(`payload.?fixed_version.orValue("") != ""`),
	entityops.VulnerabilityFields.DiscoveredAt.Expr(`payload.?discovered_at.orValue(null)`),
	entityops.VulnerabilityFields.SourceUpdatedAt.Expr(`payload.?updated_at.orValue(null)`),
	entityops.VulnerabilityFields.RawPayload.Expr("payload"),
	entityops.VulnerabilityFields.Source.Expr(providerkit.ExprInstallationName),
)
