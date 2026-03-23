package awssecurityhub

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

// mapExprFinding is the CEL mapping expression for AWS Security Hub finding payloads
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

// mapExprAssessment is the CEL mapping expression for AWS Audit Manager assessment payloads mapped to Finding
var mapExprAssessment = celMapExpr([]celMapEntry{
	{key: integrationgenerated.IntegrationMappingFindingExternalID, expr: `'id' in payload ? payload.id : ""`},
	{key: integrationgenerated.IntegrationMappingFindingExternalOwnerID, expr: `'accountId' in payload ? payload.accountId : resource`},
	{key: integrationgenerated.IntegrationMappingFindingSource, expr: `dyn("aws_audit_manager")`},
	{key: integrationgenerated.IntegrationMappingFindingCategory, expr: `'complianceType' in payload ? payload.complianceType : ""`},
	{key: integrationgenerated.IntegrationMappingFindingDisplayName, expr: `'name' in payload ? payload.name : ""`},
	{key: integrationgenerated.IntegrationMappingFindingDescription, expr: `'name' in payload && 'complianceType' in payload ? payload.name + " (" + payload.complianceType + ")" : ('name' in payload ? payload.name : "")`},
	{key: integrationgenerated.IntegrationMappingFindingStatus, expr: `'status' in payload ? payload.status : ""`},
	{key: integrationgenerated.IntegrationMappingFindingFindingClass, expr: `dyn("compliance_assessment")`},
	{key: integrationgenerated.IntegrationMappingFindingReportedAt, expr: `'creationTime' in payload ? payload.creationTime : null`},
	{key: integrationgenerated.IntegrationMappingFindingSourceUpdatedAt, expr: `'lastUpdated' in payload ? payload.lastUpdated : null`},
	{key: integrationgenerated.IntegrationMappingFindingRawPayload, expr: "payload"},
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
		{
			Schema:  integrationgenerated.IntegrationMappingSchemaFinding,
			Variant: assessmentVariant,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprAssessment,
			},
		},
	}
}
