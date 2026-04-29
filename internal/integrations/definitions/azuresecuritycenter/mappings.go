package azuresecuritycenter

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprAssessment maps AssessmentPayload fields to the Vulnerability schema.
//
// Assessments are security posture policy checks (misconfigurations), not CVE vulnerabilities.
// The ARM assessment resource ID is unique per (resource, policy) and serves as the upsert key.
// Timestamps come from AssessmentStatusResponse: first_evaluated_at → discovered_at,
// status_changed_at → source_updated_at.
var mapExprAssessment = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalOwnerID, Expr: `'resource_id' in payload && payload.resource_id != "" ? payload.resource_id : resource`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDisplayName, Expr: `'display_name' in payload ? payload.display_name : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySummary, Expr: `'display_name' in payload ? payload.display_name : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDescription, Expr: `'description' in payload ? payload.description : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, Expr: `'severity' in payload ? payload.severity : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityCategory, Expr: `'category' in payload ? payload.category : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityVulnerabilityStatusName, Expr: `'status_code' in payload ? payload.status_code : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityOpen, Expr: `dyn('status_code' in payload ? payload.status_code == "Unhealthy" : false)`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalURI, Expr: `'external_uri' in payload ? payload.external_uri : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt, Expr: `'first_evaluated_at' in payload ? payload.first_evaluated_at : null`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt, Expr: `'status_changed_at' in payload ? payload.status_changed_at : null`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityRawPayload, Expr: "payload"},
})

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
var mapExprSubAssessment = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalOwnerID, Expr: `'resource_id' in payload && payload.resource_id != "" ? payload.resource_id : resource`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDisplayName, Expr: `'display_name' in payload ? payload.display_name : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySummary, Expr: `'display_name' in payload ? payload.display_name : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDescription, Expr: `'description' in payload ? payload.description : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, Expr: `'severity' in payload ? payload.severity : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityCategory, Expr: `'category' in payload ? payload.category : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityVulnerabilityStatusName, Expr: `'status_code' in payload ? payload.status_code : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityOpen, Expr: `dyn('status_code' in payload ? payload.status_code == "Unhealthy" : false)`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityScore, Expr: `'cvss_score' in payload && payload.cvss_score != null ? payload.cvss_score : null`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt, Expr: `'published_at' in payload ? payload.published_at : null`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt, Expr: `'time_generated' in payload ? payload.time_generated : null`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityRawPayload, Expr: "payload"},
})

// azureSecurityCenterMappings returns the built-in Azure Security Center ingest mappings
func azureSecurityCenterMappings() []types.MappingRegistration {
	return []types.MappingRegistration{
		{
			Schema:  integrationgenerated.IntegrationMappingSchemaVulnerability,
			Variant: variantAssessment,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprAssessment,
			},
		},
		{
			Schema:  integrationgenerated.IntegrationMappingSchemaVulnerability,
			Variant: variantSubAssessment,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprSubAssessment,
			},
		},
	}
}
