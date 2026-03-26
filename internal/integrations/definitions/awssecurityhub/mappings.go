package awssecurityhub

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprFinding is the CEL mapping expression for AWS Security Hub finding payloads
var mapExprFinding = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalID, Expr: `'Id' in payload ? payload.Id : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalOwnerID, Expr: `size(payload.Resources) > 0 && payload.Resources[0] != null && 'Id' in payload.Resources[0] && payload.Resources[0].Id != "" ? payload.Resources[0].Id : ('AwsAccountId' in payload ? payload.AwsAccountId : resource)`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySource, Expr: `dyn("aws_security_hub")`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityCategory, Expr: `'Types' in payload && size(payload.Types) > 0 ? payload.Types[0] : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityStatus, Expr: `'Workflow' in payload && payload.Workflow != null && 'Status' in payload.Workflow && payload.Workflow.Status != "" ? payload.Workflow.Status : ('RecordState' in payload ? payload.RecordState : "")`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, Expr: `'Severity' in payload && payload.Severity != null && 'Label' in payload.Severity ? payload.Severity.Label : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySummary, Expr: `'Title' in payload ? payload.Title : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDescription, Expr: `'Description' in payload ? payload.Description : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDisplayName, Expr: `'Title' in payload ? payload.Title : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityCveID, Expr: `'Vulnerabilities' in payload && size(payload.Vulnerabilities) > 0 && payload.Vulnerabilities[0] != null && 'Id' in payload.Vulnerabilities[0] ? payload.Vulnerabilities[0].Id : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalURI, Expr: `'Vulnerabilities' in payload && size(payload.Vulnerabilities) > 0 && payload.Vulnerabilities[0] != null && 'ReferenceUrls' in payload.Vulnerabilities[0] && size(payload.Vulnerabilities[0].ReferenceUrls) > 0 ? payload.Vulnerabilities[0].ReferenceUrls[0] : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt, Expr: `'CreatedAt' in payload ? payload.CreatedAt : null`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt, Expr: `'UpdatedAt' in payload ? payload.UpdatedAt : null`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityOpen, Expr: `dyn('RecordState' in payload ? payload.RecordState == "ACTIVE" : false)`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityRawPayload, Expr: "payload"},
})

// mapExprAssessment is the CEL mapping expression for AWS Audit Manager assessment payloads mapped to Finding
var mapExprAssessment = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingFindingExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingExternalOwnerID, Expr: `'accountId' in payload ? payload.accountId : resource`},
	{Key: integrationgenerated.IntegrationMappingFindingSource, Expr: `dyn("aws_audit_manager")`},
	{Key: integrationgenerated.IntegrationMappingFindingCategory, Expr: `'complianceType' in payload ? payload.complianceType : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingDisplayName, Expr: `'name' in payload ? payload.name : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingDescription, Expr: `'name' in payload && 'complianceType' in payload ? payload.name + " (" + payload.complianceType + ")" : ('name' in payload ? payload.name : "")`},
	{Key: integrationgenerated.IntegrationMappingFindingStatus, Expr: `'status' in payload ? payload.status : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingFindingClass, Expr: `dyn("compliance_assessment")`},
	{Key: integrationgenerated.IntegrationMappingFindingReportedAt, Expr: `'creationTime' in payload ? payload.creationTime : null`},
	{Key: integrationgenerated.IntegrationMappingFindingSourceUpdatedAt, Expr: `'lastUpdated' in payload ? payload.lastUpdated : null`},
	{Key: integrationgenerated.IntegrationMappingFindingRawPayload, Expr: "payload"},
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
