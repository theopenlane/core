package gcpscc

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprFinding is the CEL mapping expression for GCP Security Command Center finding payloads
var mapExprFinding = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalID, Expr: `'name' in payload ? payload.name : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalOwnerID, Expr: "resource"},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityCategory, Expr: `'category' in payload ? payload.category : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityStatus, Expr: `'state' in payload ? payload.state : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, Expr: `'severity' in payload && payload.severity != "SEVERITY_UNSPECIFIED" ? payload.severity : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySummary, Expr: `'category' in payload && payload.category != "" ? payload.category : ('canonical_name' in payload && payload.canonical_name != "" ? payload.canonical_name : ('name' in payload ? payload.name : ""))`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDescription, Expr: `'description' in payload ? payload.description : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDisplayName, Expr: `'canonical_name' in payload && payload.canonical_name != "" ? payload.canonical_name : ('name' in payload ? payload.name : "")`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityCveID, Expr: `'vulnerability' in payload && payload.vulnerability != null && 'cve' in payload.vulnerability && payload.vulnerability.cve != null && 'id' in payload.vulnerability.cve ? payload.vulnerability.cve.id : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalURI, Expr: `'external_uri' in payload ? payload.external_uri : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt, Expr: `'create_time' in payload ? payload.create_time : null`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt, Expr: `'event_time' in payload ? payload.event_time : ('create_time' in payload ? payload.create_time : null)`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityRawPayload, Expr: "payload"},
})

// gcpsccMappings returns the default SCC ingest mappings
func gcpsccMappings() []types.MappingRegistration {
	return []types.MappingRegistration{
		{
			Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprFinding,
			},
		},
	}
}
