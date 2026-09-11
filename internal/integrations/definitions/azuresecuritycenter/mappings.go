package azuresecuritycenter

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprAssessment maps AssessmentPayload fields to the Vulnerability schema.
//
// Assessments are security posture policy checks (misconfigurations), not CVE vulnerabilities.
// The ARM assessment resource ID is unique per (resource, policy) and serves as the upsert key.
// Timestamps come from AssessmentStatusResponse: first_evaluated_at → discovered_at,
// status_changed_at → source_updated_at.
var mapExprAssessment = providerkit.CelMapExpr(
	entityops.VulnerabilityFields.ExternalID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.VulnerabilityFields.ExternalOwnerID.Expr(`'resource_id' in payload && payload.resource_id != "" ? payload.resource_id : resource`),
	entityops.VulnerabilityFields.DisplayName.Expr(`'display_name' in payload ? payload.display_name : ""`),
	entityops.VulnerabilityFields.Summary.Expr(`'display_name' in payload ? payload.display_name : ""`),
	entityops.VulnerabilityFields.Description.Expr(`paragraphs('description' in payload ? payload.description : "")`),
	entityops.VulnerabilityFields.Severity.Expr(`'severity' in payload ? payload.severity : ""`),
	entityops.VulnerabilityFields.Category.Expr(`'category' in payload ? payload.category : ""`),
	entityops.VulnerabilityFields.VulnerabilityStatusName.Expr(`'status_code' in payload ? payload.status_code : ""`),
	entityops.VulnerabilityFields.Open.Expr(`dyn('status_code' in payload ? payload.status_code == "Unhealthy" : false)`),
	entityops.VulnerabilityFields.ExternalURI.Expr(`'external_uri' in payload ? payload.external_uri : ""`),
	entityops.VulnerabilityFields.DiscoveredAt.Expr(`'first_evaluated_at' in payload ? payload.first_evaluated_at : null`),
	entityops.VulnerabilityFields.SourceUpdatedAt.Expr(`'status_changed_at' in payload ? payload.status_changed_at : null`),
	entityops.VulnerabilityFields.RawPayload.Expr("payload"),
)

// mapExprSubAssessment maps SubAssessmentPayload fields to the Vulnerability schema.
//
// Sub-assessments are granular findings: container/server types carry real CVE identifiers
// and CVSS scores; SQL types carry configuration check results. The ARM sub-assessment
// resource ID is unique per (resource, parent assessment, sub-assessment) and serves as
// the upsert key.
//
// CVE identifiers are included in raw_payload but not mapped to the cve_id field because
// the Vulnerability schema enforces a (cve_id, owner_id) unique constraint that assumes
// one record per CVE per organization, whereas Azure sub-assessments are scoped per
// resource (the same CVE can appear on multiple container images or VMs).
var mapExprSubAssessment = providerkit.CelMapExpr(
	entityops.VulnerabilityFields.ExternalID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.VulnerabilityFields.ExternalOwnerID.Expr(`'resource_id' in payload && payload.resource_id != "" ? payload.resource_id : resource`),
	entityops.VulnerabilityFields.DisplayName.Expr(`'display_name' in payload ? payload.display_name : ""`),
	entityops.VulnerabilityFields.Summary.Expr(`'display_name' in payload ? payload.display_name : ""`),
	entityops.VulnerabilityFields.Description.Expr(`paragraphs('description' in payload ? payload.description : "")`),
	entityops.VulnerabilityFields.Severity.Expr(`'severity' in payload ? payload.severity : ""`),
	entityops.VulnerabilityFields.Category.Expr(`'category' in payload ? payload.category : ""`),
	entityops.VulnerabilityFields.VulnerabilityStatusName.Expr(`'status_code' in payload ? payload.status_code : ""`),
	entityops.VulnerabilityFields.Open.Expr(`dyn('status_code' in payload ? payload.status_code == "Unhealthy" : false)`),
	entityops.VulnerabilityFields.Score.Expr(`'cvss_score' in payload && payload.cvss_score != null ? payload.cvss_score : null`),
	entityops.VulnerabilityFields.DiscoveredAt.Expr(`'published_at' in payload ? payload.published_at : null`),
	entityops.VulnerabilityFields.SourceUpdatedAt.Expr(`'time_generated' in payload ? payload.time_generated : null`),
	entityops.VulnerabilityFields.RawPayload.Expr("payload"),
)
