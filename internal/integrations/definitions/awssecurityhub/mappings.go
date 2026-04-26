package awssecurityhub

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprFinding is the CEL mapping expression for AWS Security Hub finding payloads
var mapExprFinding = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingFindingExternalID, Expr: `'Id' in payload ? payload.Id : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingDisplayName, Expr: `'Title' in payload ? payload.Title : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingOpen, Expr: `'Workflow' in payload && 'Status' in payload.Workflow ? (payload.Workflow.Status == "NEW" || payload.Workflow.Status == "NOTIFIED") : false`},
	{Key: integrationgenerated.IntegrationMappingFindingState, Expr: `'RecordState' in payload ? payload.RecordState : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingPriority, Expr: `'Criticality' in payload ? payload.Criticality : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingFindingStatusName, Expr: `'Workflow' in payload && 'Status' in payload.Workflow ? payload.Workflow.Status : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingSeverity, Expr: `'Severity' in payload && payload.Severity != null && 'Label' in payload.Severity ? payload.Severity.Label : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingScore, Expr: `'Severity' in payload && payload.Severity != null && 'Product' in payload.Severity ? payload.Severity.Product : 0.0`},
	{Key: integrationgenerated.IntegrationMappingFindingImpact, Expr: `'Severity' in payload && payload.Severity != null && 'Normalized' in payload.Severity ? payload.Severity.Normalized : 0.0`},
	{Key: integrationgenerated.IntegrationMappingFindingDescription, Expr: `'Description' in payload ? payload.Description : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingReportedAt, Expr: `'FirstObservedAt' in payload ? payload.FirstObservedAt : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingEventTime, Expr: `'LastObservedAt' in payload ? payload.LastObservedAt : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingSourceUpdatedAt, Expr: `'ProcessedAt' in payload ? payload.ProcessedAt : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingCategories, Expr: `'Types' in payload ? payload.Types : []`},
	{Key: integrationgenerated.IntegrationMappingFindingRecommendedActions, Expr: `'Remediation' in payload && payload.Remediation != null && 'Recommendation' in payload.Remediation && payload.Remediation.Recommendation != null && 'Text' in payload.Remediation.Recommendation ? payload.Remediation.Recommendation.Text : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingReferences, Expr: `'Remediation' in payload && payload.Remediation != null && 'Recommendation' in payload.Remediation && payload.Remediation.Recommendation != null && 'Url' in payload.Remediation.Recommendation ? [payload.Remediation.Recommendation.Url] : []`},
	{Key: integrationgenerated.IntegrationMappingFindingCategory, Expr: `'Types' in payload && size(payload.Types) > 0 ? payload.Types[0] : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingResourceName, Expr: `'Resources' in payload && size(payload.Resources) > 0 && payload.Resources[0] != null && 'ApplicationName' in payload.Resources[0] && payload.Resources[0].ApplicationName != null ? payload.Resources[0].ApplicationName : 'Id' in payload.Resources[0] ? payload.Resources[0].Id :  ""`},
	{Key: integrationgenerated.IntegrationMappingFindingTargets, Expr: `'Resources' in payload && size(payload.Resources) > 0 ? payload.Resources.filter(r, r != null && 'Id' in r).map(r, r.Id) : []`},
	{Key: integrationgenerated.IntegrationMappingFindingTargetDetails, Expr: `'Resources' in payload && size(payload.Resources) > 0 ? indexBy(payload.Resources.filter(r, r != null && 'Id' in r), "Id") : {}`},
	{Key: integrationgenerated.IntegrationMappingFindingExternalURI, Expr: `'SourceUrl' in payload ? payload.SourceUrl : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingExternalOwnerID, Expr: `size(payload.Resources) > 0 && 'Id' in payload.Resources[0] && payload.Resources[0].Id != "" ? payload.Resources[0].Id : 'AwsAccountId' in payload ? payload.AwsAccountId : ""`},
	{Key: integrationgenerated.IntegrationMappingFindingRawPayload, Expr: "payload"},
})

// mapExprVulnerability is the CEL mapping expression for AWS Security Hub vulnerability payloads
var mapExprVulnerability = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalID, Expr: `'Id' in payload ? payload.Id : ""`},
	// CVE ID from Vulnerabilities[0].Id
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDisplayName, Expr: `'Vulnerabilities' in payload && size(payload.Vulnerabilities) > 0 && 'Id' in payload.Vulnerabilities[0] && payload.Vulnerabilities[0].Id != "" ? payload.Vulnerabilities[0].Id : ('Title' in payload && payload.Title != "" ? payload.Title : "")`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityCveID, Expr: `'Vulnerabilities' in payload && size(payload.Vulnerabilities) > 0 && 'Id' in payload.Vulnerabilities[0] ? payload.Vulnerabilities[0].Id : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalOwnerID, Expr: `size(payload.Resources) > 0 && 'Id' in payload.Resources[0] && payload.Resources[0].Id != "" ? payload.Resources[0].Id : 'AwsAccountId' in payload ? payload.AwsAccountId : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityCategory, Expr: `'Types' in payload && size(payload.Types) > 0 ? payload.Types[0] : ""`},
	// Open if Workflow.Status is NEW or NOTIFIED
	{Key: integrationgenerated.IntegrationMappingVulnerabilityOpen, Expr: `'Workflow' in payload && 'Status' in payload.Workflow ? (payload.Workflow.Status == "NEW" || payload.Workflow.Status == "NOTIFIED") : false`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityVulnerabilityStatusName, Expr: `'Workflow' in payload && 'Status' in payload.Workflow ? payload.Workflow.Status : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, Expr: `'Severity' in payload && 'Label' in payload.Severity ? payload.Severity.Label : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySummary, Expr: `'Title' in payload ? payload.Title : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDescription, Expr: `'Description' in payload ? payload.Description : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalURI, Expr: `'Remediation' in payload && 'Recommendation' in payload.Remediation && 'Url' in payload.Remediation.Recommendation ? payload.Remediation.Recommendation.Url : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt, Expr: `'FirstObservedAt' in payload ? payload.FirstObservedAt : null`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt, Expr: `'UpdatedAt' in payload ? payload.UpdatedAt : null`},
	// fix info from Vulnerabilities[0]
	{Key: integrationgenerated.IntegrationMappingVulnerabilityFixAvailable, Expr: `'Vulnerabilities' in payload && size(payload.Vulnerabilities) > 0 && 'FixAvailable' in payload.Vulnerabilities[0] ? payload.Vulnerabilities[0].FixAvailable == "YES" : false`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityFirstPatchedVersion, Expr: `'Vulnerabilities' in payload && size(payload.Vulnerabilities) > 0 && 'VulnerablePackages' in payload.Vulnerabilities[0] && size(payload.Vulnerabilities[0].VulnerablePackages) > 0 && 'FixedInVersion' in payload.Vulnerabilities[0].VulnerablePackages[0] ? payload.Vulnerabilities[0].VulnerablePackages[0].FixedInVersion : ""`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityReferences, Expr: `'Vulnerabilities' in payload && size(payload.Vulnerabilities) > 0 && 'ReferenceUrls' in payload.Vulnerabilities[0] ? payload.Vulnerabilities[0].ReferenceUrls : []`},
	// CVSS score from Vulnerabilities[0].Cvss
	{Key: integrationgenerated.IntegrationMappingVulnerabilityScore, Expr: `'Vulnerabilities' in payload && size(payload.Vulnerabilities) > 0 && 'Cvss' in payload.Vulnerabilities[0] && size(payload.Vulnerabilities[0].Cvss) > 0 ? (payload.Vulnerabilities[0].Cvss.filter(c, 'Version' in c && c.Version.startsWith("3")).size() > 0 ? payload.Vulnerabilities[0].Cvss.filter(c, 'Version' in c && c.Version.startsWith("3"))[0].BaseScore : payload.Vulnerabilities[0].Cvss[0].BaseScore) : 0.0`},
	{Key: integrationgenerated.IntegrationMappingVulnerabilityRawPayload, Expr: "payload"},
})

// mapExprDirectoryAccount maps AWS IAM user payloads to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDisplayName, Expr: `'userName' in payload && payload.userName != "" ? payload.userName : ('id' in payload ? payload.id : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail, Expr: `'userName' in payload ? payload.userName : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountOrganizationUnit, Expr: `'path' in payload ? payload.path : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountProfile, Expr: "payload"},
})

// mapExprDirectoryGroup maps AWS IAM group payloads to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupDisplayName, Expr: `'name' in payload && payload.name != "" ? payload.name : ('id' in payload ? payload.id : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupProfile, Expr: "payload"},
})

// mapExprDirectoryMembership maps AWS IAM membership payloads to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryAccountID, Expr: `'member' in payload && payload.member != null && 'id' in payload.member ? payload.member.id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryGroupID, Expr: `'group' in payload && payload.group != null && 'id' in payload.group ? payload.group.id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipRole, Expr: `dyn("MEMBER")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipSource, Expr: `dyn("aws-iam")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipMetadata, Expr: "payload"},
})

// awsIamMappings returns the built-in AWS IAM directory sync ingest mappings
func awsIamMappings() []types.MappingRegistration {
	return []types.MappingRegistration{
		{
			Schema: integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryAccount,
			},
		},
		{
			Schema: integrationgenerated.IntegrationMappingSchemaDirectoryGroup,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryGroup,
			},
		},
		{
			Schema: integrationgenerated.IntegrationMappingSchemaDirectoryMembership,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryMembership,
			},
		},
	}
}

// awsSecurityHubMappings returns the built-in Security Hub ingest mappings
func awsSecurityHubMappings() []types.MappingRegistration {
	return []types.MappingRegistration{
		{
			Schema: integrationgenerated.IntegrationMappingSchemaFinding,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprFinding,
			},
		},
		{
			Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprVulnerability,
			},
		},
	}
}
