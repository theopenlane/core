package gcpscc

import (
	"strconv"
	"strings"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/types"
)

type celMapEntry struct {
	key  string
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

var mapExprFinding = celMapExpr([]celMapEntry{
	{key: integrationgenerated.IntegrationMappingVulnerabilityExternalID, expr: `'name' in payload ? payload.name : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityExternalOwnerID, expr: "resource"},
	{key: integrationgenerated.IntegrationMappingVulnerabilityCategory, expr: `'category' in payload ? payload.category : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityStatus, expr: `'state' in payload ? payload.state : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, expr: `'severity' in payload && payload.severity != "SEVERITY_UNSPECIFIED" ? payload.severity : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilitySummary, expr: `'category' in payload && payload.category != "" ? payload.category : ('canonical_name' in payload && payload.canonical_name != "" ? payload.canonical_name : ('name' in payload ? payload.name : ""))`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityDescription, expr: `'description' in payload ? payload.description : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityDisplayName, expr: `'canonical_name' in payload && payload.canonical_name != "" ? payload.canonical_name : ('name' in payload ? payload.name : "")`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityCveID, expr: `'vulnerability' in payload && payload.vulnerability != null && 'cve' in payload.vulnerability && payload.vulnerability.cve != null && 'id' in payload.vulnerability.cve ? payload.vulnerability.cve.id : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityExternalURI, expr: `'external_uri' in payload ? payload.external_uri : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt, expr: `'create_time' in payload ? payload.create_time : null`},
	{key: integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt, expr: `'event_time' in payload ? payload.event_time : ('create_time' in payload ? payload.create_time : null)`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityRawPayload, expr: "payload"},
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
