package awssecurityhub

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
	{key: integrationgenerated.IntegrationMappingVulnerabilityExternalID, expr: `'Id' in payload ? payload.Id : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityExternalOwnerID, expr: `size(payload.Resources) > 0 && payload.Resources[0] != null && 'Id' in payload.Resources[0] && payload.Resources[0].Id != "" ? payload.Resources[0].Id : ('AwsAccountId' in payload ? payload.AwsAccountId : resource)`},
	{key: integrationgenerated.IntegrationMappingVulnerabilitySource, expr: `dyn("aws_security_hub")`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityCategory, expr: `'Types' in payload && size(payload.Types) > 0 ? payload.Types[0] : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityStatus, expr: `'Workflow' in payload && payload.Workflow != null && 'Status' in payload.Workflow && payload.Workflow.Status != "" ? payload.Workflow.Status : ('RecordState' in payload ? payload.RecordState : "")`},
	{key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, expr: `'Severity' in payload && payload.Severity != null && 'Label' in payload.Severity ? payload.Severity.Label : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilitySummary, expr: `'Title' in payload ? payload.Title : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityDescription, expr: `'Description' in payload ? payload.Description : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityDisplayName, expr: `'Title' in payload ? payload.Title : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityCveID, expr: `'Vulnerabilities' in payload && size(payload.Vulnerabilities) > 0 && payload.Vulnerabilities[0] != null && 'Id' in payload.Vulnerabilities[0] ? payload.Vulnerabilities[0].Id : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityExternalURI, expr: `'Vulnerabilities' in payload && size(payload.Vulnerabilities) > 0 && payload.Vulnerabilities[0] != null && 'ReferenceUrls' in payload.Vulnerabilities[0] && size(payload.Vulnerabilities[0].ReferenceUrls) > 0 ? payload.Vulnerabilities[0].ReferenceUrls[0] : ""`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt, expr: `'CreatedAt' in payload ? payload.CreatedAt : null`},
	{key: integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt, expr: `'UpdatedAt' in payload ? payload.UpdatedAt : null`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityOpen, expr: `dyn('RecordState' in payload ? payload.RecordState == "ACTIVE" : false)`},
	{key: integrationgenerated.IntegrationMappingVulnerabilityRawPayload, expr: "payload"},
})

// awsSecurityHubMappings returns the built-in Security Hub ingest mappings
func awsSecurityHubMappings() []types.MappingRegistration {
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
