package awssecurityhub

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprFinding is the CEL mapping expression for AWS Security Hub finding payloads
var mapExprFinding = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyFindingExternalID, Expr: `'Id' in payload ? payload.Id : ""`},
	{Key: entityops.InputKeyFindingDisplayName, Expr: `'Title' in payload ? payload.Title : ""`},
	{Key: entityops.InputKeyFindingOpen, Expr: `'Workflow' in payload && 'Status' in payload.Workflow ? (payload.Workflow.Status == "NEW" || payload.Workflow.Status == "NOTIFIED") : false`},
	{Key: entityops.InputKeyFindingState, Expr: `'RecordState' in payload ? payload.RecordState : ""`},
	{Key: entityops.InputKeyFindingPriority, Expr: `'Criticality' in payload ? payload.Criticality : ""`},
	{Key: entityops.InputKeyFindingFindingStatusName, Expr: `'Workflow' in payload && 'Status' in payload.Workflow ? payload.Workflow.Status : ""`},
	{Key: entityops.InputKeyFindingSeverity, Expr: `'Severity' in payload && payload.Severity != null && 'Label' in payload.Severity ? payload.Severity.Label : ""`},
	{Key: entityops.InputKeyFindingScore, Expr: `'Severity' in payload && payload.Severity != null && 'Product' in payload.Severity ? payload.Severity.Product : 0.0`},
	{Key: entityops.InputKeyFindingImpact, Expr: `'Severity' in payload && payload.Severity != null && 'Normalized' in payload.Severity ? payload.Severity.Normalized : 0.0`},
	{Key: entityops.InputKeyFindingDescription, Expr: `'Description' in payload ? payload.Description : ""`},
	{Key: entityops.InputKeyFindingReportedAt, Expr: `'FirstObservedAt' in payload ? payload.FirstObservedAt : ""`},
	{Key: entityops.InputKeyFindingEventTime, Expr: `'LastObservedAt' in payload ? payload.LastObservedAt : ""`},
	{Key: entityops.InputKeyFindingSourceUpdatedAt, Expr: `'ProcessedAt' in payload ? payload.ProcessedAt : ""`},
	{Key: entityops.InputKeyFindingCategories, Expr: `'Types' in payload ? payload.Types : []`},
	{Key: entityops.InputKeyFindingRecommendedActions, Expr: `'Remediation' in payload && payload.Remediation != null && 'Recommendation' in payload.Remediation && payload.Remediation.Recommendation != null && 'Text' in payload.Remediation.Recommendation ? payload.Remediation.Recommendation.Text : ""`},
	{Key: entityops.InputKeyFindingReferences, Expr: `'Remediation' in payload && payload.Remediation != null && 'Recommendation' in payload.Remediation && payload.Remediation.Recommendation != null && 'Url' in payload.Remediation.Recommendation ? [payload.Remediation.Recommendation.Url] : []`},
	{Key: entityops.InputKeyFindingCategory, Expr: `'Types' in payload && payload.Types != null && size(payload.Types) > 0 ? payload.Types[0] : ""`},
	{Key: entityops.InputKeyFindingResourceName, Expr: `'Resources' in payload && payload.Resources != null && size(payload.Resources) > 0 && payload.Resources[0] != null ? ('ApplicationName' in payload.Resources[0] && payload.Resources[0].ApplicationName != null ? payload.Resources[0].ApplicationName : ('Id' in payload.Resources[0] ? payload.Resources[0].Id : "")) : ""`},
	{Key: entityops.InputKeyFindingTargets, Expr: `'Resources' in payload && payload.Resources != null && size(payload.Resources) > 0 ? payload.Resources.filter(r, r != null && 'Id' in r).map(r, r.Id) : []`},
	{Key: entityops.InputKeyFindingTargetDetails, Expr: `'Resources' in payload && payload.Resources != null && size(payload.Resources) > 0 ? indexBy(payload.Resources.filter(r, r != null && 'Id' in r), "Id") : {}`},
	{Key: entityops.InputKeyFindingExternalURI, Expr: `'SourceUrl' in payload ? payload.SourceUrl : ""`},
	{Key: entityops.InputKeyFindingExternalOwnerID, Expr: `'Resources' in payload && payload.Resources != null && size(payload.Resources) > 0 && payload.Resources[0] != null && 'Id' in payload.Resources[0] && payload.Resources[0].Id != "" ? payload.Resources[0].Id : 'AwsAccountId' in payload ? payload.AwsAccountId : ""`},
	{Key: entityops.InputKeyFindingRawPayload, Expr: "payload"},
})

// mapExprVulnerability is the CEL mapping expression for AWS Security Hub vulnerability payloads
var mapExprVulnerability = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyVulnerabilityExternalID, Expr: `'Id' in payload ? payload.Id : ""`},
	// CVE ID from Vulnerabilities[0].Id
	{Key: entityops.InputKeyVulnerabilityDisplayName, Expr: `'Vulnerabilities' in payload && payload.Vulnerabilities != null && size(payload.Vulnerabilities) > 0 && payload.Vulnerabilities[0] != null && 'Id' in payload.Vulnerabilities[0] && payload.Vulnerabilities[0].Id != "" ? payload.Vulnerabilities[0].Id : ('Title' in payload && payload.Title != "" ? payload.Title : "")`},
	{Key: entityops.InputKeyVulnerabilityCveID, Expr: `'Vulnerabilities' in payload && payload.Vulnerabilities != null && size(payload.Vulnerabilities) > 0 && payload.Vulnerabilities[0] != null && 'Id' in payload.Vulnerabilities[0] ? payload.Vulnerabilities[0].Id : ""`},
	{Key: entityops.InputKeyVulnerabilityExternalOwnerID, Expr: `'Resources' in payload && payload.Resources != null && size(payload.Resources) > 0 && payload.Resources[0] != null && 'Id' in payload.Resources[0] && payload.Resources[0].Id != "" ? payload.Resources[0].Id : 'AwsAccountId' in payload ? payload.AwsAccountId : ""`},
	{Key: entityops.InputKeyVulnerabilityCategory, Expr: `'Types' in payload && payload.Types != null && size(payload.Types) > 0 ? payload.Types[0] : ""`},
	// Open if Workflow.Status is NEW or NOTIFIED
	{Key: entityops.InputKeyVulnerabilityOpen, Expr: `'Workflow' in payload && 'Status' in payload.Workflow ? (payload.Workflow.Status == "NEW" || payload.Workflow.Status == "NOTIFIED") : false`},
	{Key: entityops.InputKeyVulnerabilityVulnerabilityStatusName, Expr: `'Workflow' in payload && 'Status' in payload.Workflow ? payload.Workflow.Status : ""`},
	{Key: entityops.InputKeyVulnerabilitySeverity, Expr: `'Severity' in payload && 'Label' in payload.Severity ? payload.Severity.Label : ""`},
	{Key: entityops.InputKeyVulnerabilitySummary, Expr: `'Title' in payload ? payload.Title : ""`},
	{Key: entityops.InputKeyVulnerabilityDescription, Expr: `'Description' in payload ? payload.Description : ""`},
	{Key: entityops.InputKeyVulnerabilityExternalURI, Expr: `'Remediation' in payload && 'Recommendation' in payload.Remediation && 'Url' in payload.Remediation.Recommendation ? payload.Remediation.Recommendation.Url : ""`},
	{Key: entityops.InputKeyVulnerabilityDiscoveredAt, Expr: `'FirstObservedAt' in payload ? payload.FirstObservedAt : null`},
	{Key: entityops.InputKeyVulnerabilitySourceUpdatedAt, Expr: `'UpdatedAt' in payload ? payload.UpdatedAt : null`},
	// fix info from Vulnerabilities[0]
	{Key: entityops.InputKeyVulnerabilityFixAvailable, Expr: `'Vulnerabilities' in payload && payload.Vulnerabilities != null && size(payload.Vulnerabilities) > 0 && payload.Vulnerabilities[0] != null && 'FixAvailable' in payload.Vulnerabilities[0] ? payload.Vulnerabilities[0].FixAvailable == "YES" : false`},
	{Key: entityops.InputKeyVulnerabilityFirstPatchedVersion, Expr: `'Vulnerabilities' in payload && payload.Vulnerabilities != null && size(payload.Vulnerabilities) > 0 && payload.Vulnerabilities[0] != null && 'VulnerablePackages' in payload.Vulnerabilities[0] && payload.Vulnerabilities[0].VulnerablePackages != null && size(payload.Vulnerabilities[0].VulnerablePackages) > 0 && payload.Vulnerabilities[0].VulnerablePackages[0] != null && 'FixedInVersion' in payload.Vulnerabilities[0].VulnerablePackages[0] ? payload.Vulnerabilities[0].VulnerablePackages[0].FixedInVersion : ""`},
	{Key: entityops.InputKeyVulnerabilityReferences, Expr: `'Vulnerabilities' in payload && payload.Vulnerabilities != null && size(payload.Vulnerabilities) > 0 && payload.Vulnerabilities[0] != null && 'ReferenceUrls' in payload.Vulnerabilities[0] ? payload.Vulnerabilities[0].ReferenceUrls : []`},
	// CVSS score from Vulnerabilities[0].Cvss
	{Key: entityops.InputKeyVulnerabilityScore, Expr: `'Vulnerabilities' in payload && payload.Vulnerabilities != null && size(payload.Vulnerabilities) > 0 && payload.Vulnerabilities[0] != null && 'Cvss' in payload.Vulnerabilities[0] && payload.Vulnerabilities[0].Cvss != null && size(payload.Vulnerabilities[0].Cvss) > 0 ? (payload.Vulnerabilities[0].Cvss.filter(c, 'Version' in c && c.Version.startsWith("3")).size() > 0 ? payload.Vulnerabilities[0].Cvss.filter(c, 'Version' in c && c.Version.startsWith("3"))[0].BaseScore : payload.Vulnerabilities[0].Cvss[0].BaseScore) : 0.0`},
	{Key: entityops.InputKeyVulnerabilityRawPayload, Expr: "payload"},
})

// mapExprDirectoryAccount maps AWS IAM user payloads to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryAccountExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: entityops.InputKeyDirectoryAccountDisplayName, Expr: `'userName' in payload && payload.userName != "" ? payload.userName : ('id' in payload ? payload.id : "")`},
	{Key: entityops.InputKeyDirectoryAccountCanonicalEmail, Expr: `'userName' in payload ? payload.userName : ""`},
	{Key: entityops.InputKeyDirectoryAccountOrganizationUnit, Expr: `'path' in payload ? payload.path : ""`},
	{Key: entityops.InputKeyDirectoryAccountProfile, Expr: "payload"},
})

// mapExprDirectoryGroup maps AWS IAM group payloads to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryGroupExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: entityops.InputKeyDirectoryGroupDisplayName, Expr: `'name' in payload && payload.name != "" ? payload.name : ('id' in payload ? payload.id : "")`},
	{Key: entityops.InputKeyDirectoryGroupProfile, Expr: "payload"},
})

// mapExprDirectoryMembership maps AWS IAM membership payloads to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryMembershipDirectoryAccountID, Expr: `'member' in payload && payload.member != null && 'id' in payload.member ? payload.member.id : ""`},
	{Key: entityops.InputKeyDirectoryMembershipDirectoryGroupID, Expr: `'group' in payload && payload.group != null && 'id' in payload.group ? payload.group.id : ""`},
	{Key: entityops.InputKeyDirectoryMembershipRole, Expr: `dyn("MEMBER")`},
	{Key: entityops.InputKeyDirectoryMembershipMetadata, Expr: "payload"},
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
