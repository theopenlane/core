package awssecurityhub

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprFinding is the CEL mapping expression for AWS Security Hub finding payloads
var mapExprFinding = providerkit.CelMapExpr(
	entityops.FindingFields.ExternalID.Expr(`'Id' in payload ? payload.Id : ""`),
	entityops.FindingFields.DisplayName.Expr(`'Title' in payload ? payload.Title : ""`),
	entityops.FindingFields.Open.Expr(`'Workflow' in payload && 'Status' in payload.Workflow ? (payload.Workflow.Status == "NEW" || payload.Workflow.Status == "NOTIFIED") : false`),
	entityops.FindingFields.State.Expr(`'RecordState' in payload ? payload.RecordState : ""`),
	entityops.FindingFields.Priority.Expr(`'Criticality' in payload ? payload.Criticality : ""`),
	entityops.FindingFields.FindingStatusName.Expr(`'Workflow' in payload && 'Status' in payload.Workflow ? payload.Workflow.Status : ""`),
	entityops.FindingFields.Severity.Expr(`'Severity' in payload && payload.Severity != null && 'Label' in payload.Severity ? payload.Severity.Label : ""`),
	entityops.FindingFields.Score.Expr(`'Severity' in payload && payload.Severity != null && 'Product' in payload.Severity ? payload.Severity.Product : 0.0`),
	entityops.FindingFields.Impact.Expr(`'Severity' in payload && payload.Severity != null && 'Normalized' in payload.Severity ? payload.Severity.Normalized : 0.0`),
	entityops.FindingFields.Description.Expr(`paragraphs('Description' in payload ? payload.Description : "")`),
	entityops.FindingFields.ReportedAt.Expr(`'FirstObservedAt' in payload ? payload.FirstObservedAt : ""`),
	entityops.FindingFields.EventTime.Expr(`'LastObservedAt' in payload ? payload.LastObservedAt : ""`),
	entityops.FindingFields.SourceUpdatedAt.Expr(`'ProcessedAt' in payload ? payload.ProcessedAt : ""`),
	entityops.FindingFields.Categories.Expr(`'Types' in payload ? payload.Types : []`),
	entityops.FindingFields.RecommendedActions.Expr(`'Remediation' in payload && payload.Remediation != null && 'Recommendation' in payload.Remediation && payload.Remediation.Recommendation != null && 'Text' in payload.Remediation.Recommendation ? payload.Remediation.Recommendation.Text : ""`),
	entityops.FindingFields.References.Expr(`'Remediation' in payload && payload.Remediation != null && 'Recommendation' in payload.Remediation && payload.Remediation.Recommendation != null && 'Url' in payload.Remediation.Recommendation ? [payload.Remediation.Recommendation.Url] : []`),
	entityops.FindingFields.Category.Expr(`'Types' in payload && payload.Types != null && size(payload.Types) > 0 ? payload.Types[0] : ""`),
	entityops.FindingFields.ResourceName.Expr(`'Resources' in payload && payload.Resources != null && size(payload.Resources) > 0 && payload.Resources[0] != null ? ('ApplicationName' in payload.Resources[0] && payload.Resources[0].ApplicationName != null ? payload.Resources[0].ApplicationName : ('Id' in payload.Resources[0] ? payload.Resources[0].Id : "")) : ""`),
	entityops.FindingFields.Targets.Expr(`'Resources' in payload && payload.Resources != null && size(payload.Resources) > 0 ? payload.Resources.filter(r, r != null && 'Id' in r).map(r, r.Id) : []`),
	entityops.FindingFields.TargetDetails.Expr(`'Resources' in payload && payload.Resources != null && size(payload.Resources) > 0 ? indexBy(payload.Resources.filter(r, r != null && 'Id' in r), "Id") : {}`),
	entityops.FindingFields.ExternalURI.Expr(`'SourceUrl' in payload ? payload.SourceUrl : ""`),
	entityops.FindingFields.ExternalOwnerID.Expr(`'Resources' in payload && payload.Resources != null && size(payload.Resources) > 0 && payload.Resources[0] != null && 'Id' in payload.Resources[0] && payload.Resources[0].Id != "" ? payload.Resources[0].Id : 'AwsAccountId' in payload ? payload.AwsAccountId : ""`),
	entityops.FindingFields.RawPayload.Expr("payload"),
)

// mapExprVulnerability is the CEL mapping expression for AWS Security Hub vulnerability payloads
var mapExprVulnerability = providerkit.CelMapExpr(
	entityops.VulnerabilityFields.ExternalID.Expr(`'Id' in payload ? payload.Id : ""`),
	// CVE ID from Vulnerabilities[0].Id
	entityops.VulnerabilityFields.DisplayName.Expr(`'Vulnerabilities' in payload && payload.Vulnerabilities != null && size(payload.Vulnerabilities) > 0 && payload.Vulnerabilities[0] != null && 'Id' in payload.Vulnerabilities[0] && payload.Vulnerabilities[0].Id != "" ? payload.Vulnerabilities[0].Id : ('Title' in payload && payload.Title != "" ? payload.Title : "")`),
	entityops.VulnerabilityFields.CveID.Expr(`'Vulnerabilities' in payload && payload.Vulnerabilities != null && size(payload.Vulnerabilities) > 0 && payload.Vulnerabilities[0] != null && 'Id' in payload.Vulnerabilities[0] ? payload.Vulnerabilities[0].Id : ""`),
	entityops.VulnerabilityFields.ExternalOwnerID.Expr(`'Resources' in payload && payload.Resources != null && size(payload.Resources) > 0 && payload.Resources[0] != null && 'Id' in payload.Resources[0] && payload.Resources[0].Id != "" ? payload.Resources[0].Id : 'AwsAccountId' in payload ? payload.AwsAccountId : ""`),
	entityops.VulnerabilityFields.Category.Expr(`'Types' in payload && payload.Types != null && size(payload.Types) > 0 ? payload.Types[0] : ""`),
	// Open if Workflow.Status is NEW or NOTIFIED
	entityops.VulnerabilityFields.Open.Expr(`'Workflow' in payload && 'Status' in payload.Workflow ? (payload.Workflow.Status == "NEW" || payload.Workflow.Status == "NOTIFIED") : false`),
	entityops.VulnerabilityFields.VulnerabilityStatusName.Expr(`'Workflow' in payload && 'Status' in payload.Workflow ? payload.Workflow.Status : ""`),
	entityops.VulnerabilityFields.Severity.Expr(`'Severity' in payload && 'Label' in payload.Severity ? payload.Severity.Label : ""`),
	entityops.VulnerabilityFields.Summary.Expr(`'Title' in payload ? payload.Title : ""`),
	entityops.VulnerabilityFields.Description.Expr(`paragraphs('Description' in payload ? payload.Description : "")`),
	entityops.VulnerabilityFields.ExternalURI.Expr(`'Remediation' in payload && 'Recommendation' in payload.Remediation && 'Url' in payload.Remediation.Recommendation ? payload.Remediation.Recommendation.Url : ""`),
	entityops.VulnerabilityFields.DiscoveredAt.Expr(`'FirstObservedAt' in payload ? payload.FirstObservedAt : null`),
	entityops.VulnerabilityFields.SourceUpdatedAt.Expr(`'UpdatedAt' in payload ? payload.UpdatedAt : null`),
	// fix info from Vulnerabilities[0]
	entityops.VulnerabilityFields.FixAvailable.Expr(`'Vulnerabilities' in payload && payload.Vulnerabilities != null && size(payload.Vulnerabilities) > 0 && payload.Vulnerabilities[0] != null && 'FixAvailable' in payload.Vulnerabilities[0] ? payload.Vulnerabilities[0].FixAvailable == "YES" : false`),
	entityops.VulnerabilityFields.FirstPatchedVersion.Expr(`'Vulnerabilities' in payload && payload.Vulnerabilities != null && size(payload.Vulnerabilities) > 0 && payload.Vulnerabilities[0] != null && 'VulnerablePackages' in payload.Vulnerabilities[0] && payload.Vulnerabilities[0].VulnerablePackages != null && size(payload.Vulnerabilities[0].VulnerablePackages) > 0 && payload.Vulnerabilities[0].VulnerablePackages[0] != null && 'FixedInVersion' in payload.Vulnerabilities[0].VulnerablePackages[0] ? payload.Vulnerabilities[0].VulnerablePackages[0].FixedInVersion : ""`),
	entityops.VulnerabilityFields.References.Expr(`'Vulnerabilities' in payload && payload.Vulnerabilities != null && size(payload.Vulnerabilities) > 0 && payload.Vulnerabilities[0] != null && 'ReferenceUrls' in payload.Vulnerabilities[0] ? payload.Vulnerabilities[0].ReferenceUrls : []`),
	// CVSS score from Vulnerabilities[0].Cvss
	entityops.VulnerabilityFields.Score.Expr(`'Vulnerabilities' in payload && payload.Vulnerabilities != null && size(payload.Vulnerabilities) > 0 && payload.Vulnerabilities[0] != null && 'Cvss' in payload.Vulnerabilities[0] && payload.Vulnerabilities[0].Cvss != null && size(payload.Vulnerabilities[0].Cvss) > 0 ? (payload.Vulnerabilities[0].Cvss.filter(c, 'Version' in c && c.Version.startsWith("3")).size() > 0 ? payload.Vulnerabilities[0].Cvss.filter(c, 'Version' in c && c.Version.startsWith("3"))[0].BaseScore : payload.Vulnerabilities[0].Cvss[0].BaseScore) : 0.0`),
	entityops.VulnerabilityFields.RawPayload.Expr("payload"),
	entityops.VulnerabilityFields.Source.Expr("installation.name"),
)

// mapExprDirectoryAccount maps AWS IAM user payloads to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr(
	entityops.DirectoryAccountFields.ExternalID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.DirectoryAccountFields.DisplayName.Expr(`'userName' in payload && payload.userName != "" ? payload.userName : ('id' in payload ? payload.id : "")`),
	entityops.DirectoryAccountFields.CanonicalEmail.Expr(`'userName' in payload ? payload.userName : ""`),
	entityops.DirectoryAccountFields.OrganizationUnit.Expr(`'path' in payload ? payload.path : ""`),
	entityops.DirectoryAccountFields.Profile.Expr("payload"),
	entityops.DirectoryAccountFields.DirectoryName.Expr("installation.name"),
	entityops.DirectoryAccountFields.PrimarySource.Expr("installation.primary_directory"),
)

// mapExprDirectoryGroup maps AWS IAM group payloads to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr(
	entityops.DirectoryGroupFields.ExternalID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.DirectoryGroupFields.DisplayName.Expr(`'name' in payload && payload.name != "" ? payload.name : ('id' in payload ? payload.id : "")`),
	entityops.DirectoryGroupFields.Profile.Expr("payload"),
	entityops.DirectoryGroupFields.DirectoryName.Expr("installation.name"),
)

// mapExprDirectoryMembership maps AWS IAM membership payloads to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr(
	entityops.DirectoryMembershipFields.DirectoryAccountID.Expr(`'member' in payload && payload.member != null && 'id' in payload.member ? payload.member.id : ""`),
	entityops.DirectoryMembershipFields.DirectoryGroupID.Expr(`'group' in payload && payload.group != null && 'id' in payload.group ? payload.group.id : ""`),
	entityops.DirectoryMembershipFields.Role.Expr(`dyn("MEMBER")`),
	entityops.DirectoryMembershipFields.Metadata.Expr("payload"),
	entityops.DirectoryMembershipFields.DirectoryName.Expr("installation.name"),
)
