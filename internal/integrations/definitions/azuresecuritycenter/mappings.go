package azuresecuritycenter

import (
	"strconv"
	"strings"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/types"
)

// celMapEntry holds one key-expression pair for building CEL object literal mapping expressions
type celMapEntry struct {
	// key is the target field name in the mapped output document
	key string
	// expr is the CEL expression that produces the value for key
	expr string
}

// celMapExpr renders CEL map entries into a CEL object literal string
func celMapExpr(entries []celMapEntry) string {
	if len(entries) == 0 {
		return "{}"
	}

	var b strings.Builder

	b.WriteString("{\n")

	for i, entry := range entries {
		b.WriteString("  ")
		b.WriteString(strconv.Quote(entry.key))
		b.WriteString(": ")
		b.WriteString(entry.expr)

		if i < len(entries)-1 {
			b.WriteString(",")
		}

		b.WriteString("\n")
	}

	b.WriteString("}")

	return b.String()
}

// mapExprAssessment maps AssessmentPayload fields to the Vulnerability schema.
//
// Assessments are security posture policy checks (misconfigurations), not CVE vulnerabilities.
// The ARM assessment resource ID is unique per (resource, policy) and serves as the upsert key.
// Timestamps come from AssessmentStatusResponse: first_evaluated_at → discovered_at,
// status_changed_at → source_updated_at.
var mapExprAssessment = celMapExpr([]celMapEntry{
	{key: integrationgenerated.IntegrationMappingVulnerabilityExternalID, expr: `'id' in payload ? payload.id : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityExternalOwnerID, expr: `'resource_id' in payload && payload.resource_id != "" ? payload.resource_id : resource`},
	{key: integrationgenerated.IntegrationMappingVulnerabilitySource, expr: `dyn("azure_security_center")`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityDisplayName, expr: `'display_name' in payload ? payload.display_name : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilitySummary, expr: `'display_name' in payload ? payload.display_name : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityDescription, expr: `'description' in payload ? payload.description : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, expr: `'severity' in payload ? payload.severity : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityCategory, expr: `'category' in payload ? payload.category : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityStatus, expr: `'status_code' in payload ? payload.status_code : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityOpen, expr: `dyn('status_code' in payload ? payload.status_code == "Unhealthy" : false)`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityExternalURI, expr: `'external_uri' in payload ? payload.external_uri : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt, expr: `'first_evaluated_at' in payload ? payload.first_evaluated_at : null`},
	{key: integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt, expr: `'status_changed_at' in payload ? payload.status_changed_at : null`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityRawPayload, expr: "payload"},
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
var mapExprSubAssessment = celMapExpr([]celMapEntry{
	{key: integrationgenerated.IntegrationMappingVulnerabilityExternalID, expr: `'id' in payload ? payload.id : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityExternalOwnerID, expr: `'resource_id' in payload && payload.resource_id != "" ? payload.resource_id : resource`},
	{key: integrationgenerated.IntegrationMappingVulnerabilitySource, expr: `dyn("azure_security_center")`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityDisplayName, expr: `'display_name' in payload ? payload.display_name : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilitySummary, expr: `'display_name' in payload ? payload.display_name : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityDescription, expr: `'description' in payload ? payload.description : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, expr: `'severity' in payload ? payload.severity : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityCategory, expr: `'category' in payload ? payload.category : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityStatus, expr: `'status_code' in payload ? payload.status_code : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityOpen, expr: `dyn('status_code' in payload ? payload.status_code == "Unhealthy" : false)`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityScore, expr: `'cvss_score' in payload && payload.cvss_score != null ? payload.cvss_score : null`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt, expr: `'published_at' in payload ? payload.published_at : null`},
	{key: integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt, expr: `'time_generated' in payload ? payload.time_generated : null`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityRawPayload, expr: "payload"},
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
