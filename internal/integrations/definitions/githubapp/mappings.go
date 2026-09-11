package githubapp

import (
	"strconv"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// webhookBaseEntries returns mapping entries for webhook alert payloads (snake_case JSON keys)
func webhookBaseEntries(category, externalIDExpr string) []entityops.MappingEntry {
	return []entityops.MappingEntry{
		entityops.VulnerabilityFields.ExternalID.Expr(externalIDExpr),
		entityops.VulnerabilityFields.ExternalOwnerID.Expr("resource"),
		entityops.VulnerabilityFields.Category.Expr(strconv.Quote(category)),
		entityops.VulnerabilityFields.VulnerabilityStatusName.Expr(`'state' in payload ? payload.state : ""`),
		entityops.VulnerabilityFields.ExternalURI.Expr(`'html_url' in payload ? payload.html_url : ""`),
		entityops.VulnerabilityFields.SourceUpdatedAt.Expr(`'updated_at' in payload ? payload.updated_at : null`),
		entityops.VulnerabilityFields.DiscoveredAt.Expr(`'created_at' in payload ? payload.created_at : null`),
		entityops.VulnerabilityFields.PublishedAt.Expr(`'security_advisory' in payload && 'published_at' in payload.security_advisory ? payload.security_advisory.published_at : null`),
		entityops.VulnerabilityFields.DismissedAt.Expr(`'dismissed_at' in payload ? payload.dismissed_at : null`),
		entityops.VulnerabilityFields.DismissedReason.Expr(`'dismissed_reason' in payload ? payload.dismissed_reason : null`),
		entityops.VulnerabilityFields.DismissedComment.Expr(`'dismissed_comment' in payload ? payload.dismissed_comment : null`),
		entityops.VulnerabilityFields.FixedAt.Expr(`'fixed_at' in payload ? payload.fixed_at : null`),
		entityops.VulnerabilityFields.AutoDismissedAt.Expr(`'auto_dismissed_at' in payload ? payload.auto_dismissed_at : null`),
		entityops.VulnerabilityFields.Open.Expr(`'state' in payload ? payload.state == "open" : false`),
		entityops.VulnerabilityFields.RawPayload.Expr("payload"),
		entityops.VulnerabilityFields.Source.Expr("installation.name"),
	}
}

// pollBaseEntries returns mapping entries for GraphQL poll payloads (PascalCase JSON keys)
func pollBaseEntries(category, externalIDExpr string) []entityops.MappingEntry {
	return []entityops.MappingEntry{
		entityops.VulnerabilityFields.ExternalID.Expr(externalIDExpr),
		entityops.VulnerabilityFields.ExternalOwnerID.Expr("resource"),
		entityops.VulnerabilityFields.Category.Expr(strconv.Quote(category)),
		entityops.VulnerabilityFields.VulnerabilityStatusName.Expr(`'State' in payload ? payload.State : ""`),
		entityops.VulnerabilityFields.ExternalURI.Expr(`'Number' in payload && payload.Number != 0 ? "https://github.com/" + resource + "/security/dependabot/" + string(payload.Number) : ""`),
		entityops.VulnerabilityFields.SourceUpdatedAt.Expr(`'SecurityVulnerability' in payload && 'UpdatedAt' in payload.SecurityVulnerability ? payload.SecurityVulnerability.UpdatedAt : null`),
		entityops.VulnerabilityFields.DiscoveredAt.Expr(`'CreatedAt' in payload ? payload.CreatedAt : null`),
		entityops.VulnerabilityFields.PublishedAt.Expr(`'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'PublishedAt' in payload.SecurityVulnerability.Advisory ? payload.SecurityVulnerability.Advisory.PublishedAt : null`),
		entityops.VulnerabilityFields.DismissedAt.Expr(`'DismissedAt' in payload ? payload.DismissedAt : null`),
		entityops.VulnerabilityFields.DismissedReason.Expr(`'DismissReason' in payload ? payload.DismissReason : null`),
		entityops.VulnerabilityFields.DismissedComment.Expr(`'DismissComment' in payload ? payload.DismissComment : null`),
		entityops.VulnerabilityFields.FixedAt.Expr(`'FixedAt' in payload ? payload.FixedAt : null`),
		entityops.VulnerabilityFields.AutoDismissedAt.Expr(`'AutoDismissedAt' in payload ? payload.AutoDismissedAt : null`),
		entityops.VulnerabilityFields.Open.Expr(`'State' in payload ? payload.State == "OPEN" : false`),
		entityops.VulnerabilityFields.RawPayload.Expr("payload"),
		entityops.VulnerabilityFields.Source.Expr("installation.name"),
	}
}

// buildWebhookMappingExpr constructs a CEL mapping expression for webhook payloads
func buildWebhookMappingExpr(category, externalIDExpr string, extras []entityops.MappingEntry) string {
	entries := webhookBaseEntries(category, externalIDExpr)
	entries = append(entries, extras...)

	return providerkit.CelMapExpr(entries...)
}

// buildPollMappingExpr constructs a CEL mapping expression for GraphQL poll payloads
func buildPollMappingExpr(category, externalIDExpr string, extras []entityops.MappingEntry) string {
	entries := pollBaseEntries(category, externalIDExpr)
	entries = append(entries, extras...)

	return providerkit.CelMapExpr(entries...)
}

var (
	// mapExprDependabot is the CEL mapping expression for Dependabot webhook alert payloads
	mapExprDependabot = buildWebhookMappingExpr(githubAlertTypeDependabot, `"github:" + resource + ":dependabot:" + ('number' in payload && payload.number != 0 ? string(payload.number) : ('security_advisory' in payload && 'ghsa_id' in payload.security_advisory && payload.security_advisory.ghsa_id != "" ? payload.security_advisory.ghsa_id : "unknown"))`, []entityops.MappingEntry{
		entityops.VulnerabilityFields.Severity.Expr(`'security_advisory' in payload && 'severity' in payload.security_advisory ? payload.security_advisory.severity : ""`),
		entityops.VulnerabilityFields.Summary.Expr(`'security_advisory' in payload && 'summary' in payload.security_advisory ? payload.security_advisory.summary : ""`),
		entityops.VulnerabilityFields.Description.Expr(`paragraphs('security_advisory' in payload && 'description' in payload.security_advisory ? payload.security_advisory.description : "")`),
		entityops.VulnerabilityFields.CveID.Expr(`'security_advisory' in payload && 'cve_id' in payload.security_advisory ? payload.security_advisory.cve_id : ""`),
		entityops.VulnerabilityFields.DisplayName.Expr(`'security_advisory' in payload && 'cve_id' in payload.security_advisory ? payload.security_advisory.cve_id : ""`),
		entityops.VulnerabilityFields.Score.Expr(`'security_advisory' in payload && 'cvss' in payload.security_advisory && 'score' in payload.security_advisory.cvss ? payload.security_advisory.cvss.score : null`),
		entityops.VulnerabilityFields.Vector.Expr(`'security_advisory' in payload && 'cvss' in payload.security_advisory && 'vector_string' in payload.security_advisory.cvss ? payload.security_advisory.cvss.vector_string : ""`),
		entityops.VulnerabilityFields.CweIds.Expr(`'security_advisory' in payload && 'cwes' in payload.security_advisory ? payload.security_advisory.cwes.map(c, c.cwe_id) : []`),
		entityops.VulnerabilityFields.References.Expr(`'security_advisory' in payload && 'references' in payload.security_advisory ? payload.security_advisory.references.map(r, r.url) : []`),
		entityops.VulnerabilityFields.VulnerableVersionRange.Expr(`'security_vulnerability' in payload && 'vulnerable_version_range' in payload.security_vulnerability ? payload.security_vulnerability.vulnerable_version_range : ""`),
		entityops.VulnerabilityFields.FirstPatchedVersion.Expr(`'security_vulnerability' in payload && 'first_patched_version' in payload.security_vulnerability && 'identifier' in payload.security_vulnerability.first_patched_version ? payload.security_vulnerability.first_patched_version.identifier : ""`),
		entityops.VulnerabilityFields.PackageName.Expr(`'security_vulnerability' in payload && 'package' in payload.security_vulnerability && 'name' in payload.security_vulnerability.package ? payload.security_vulnerability.package.name : ""`),
		entityops.VulnerabilityFields.PackageEcosystem.Expr(`'security_vulnerability' in payload && 'package' in payload.security_vulnerability && 'ecosystem' in payload.security_vulnerability.package ? payload.security_vulnerability.package.ecosystem : ""`),
		entityops.VulnerabilityFields.ManifestPath.Expr(`'dependency' in payload && 'manifest_path' in payload.dependency ? payload.dependency.manifest_path : ""`),
		entityops.VulnerabilityFields.DependencyScope.Expr(`'dependency' in payload && 'scope' in payload.dependency ? payload.dependency.scope : ""`),
	},
	)
	// mapExprDependabotPoll is the CEL mapping expression for Dependabot alerts collected via GraphQL poll
	mapExprDependabotPoll = buildPollMappingExpr(githubAlertTypeDependabot, `"github:" + resource + ":dependabot:" + ('Number' in payload && payload.Number != 0 ? string(payload.Number) : ('SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'GHSAID' in payload.SecurityVulnerability.Advisory && payload.SecurityVulnerability.Advisory.GHSAID != "" ? payload.SecurityVulnerability.Advisory.GHSAID : "unknown"))`, []entityops.MappingEntry{
		entityops.VulnerabilityFields.DisplayName.Expr(`'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'Identifiers' in payload.SecurityVulnerability.Advisory && payload.SecurityVulnerability.Advisory.Identifiers.filter(i, i.Type == "CVE").size() > 0 ? payload.SecurityVulnerability.Advisory.Identifiers.filter(i, i.Type == "CVE")[0].Value : 'GHSAID' in payload.SecurityVulnerability.Advisory ? payload.SecurityVulnerability.Advisory.GHSAID : ""`),
		entityops.VulnerabilityFields.Severity.Expr(`'SecurityVulnerability' in payload && 'Severity' in payload.SecurityVulnerability ? payload.SecurityVulnerability.Severity : ""`),
		entityops.VulnerabilityFields.Summary.Expr(`'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'Summary' in payload.SecurityVulnerability.Advisory ? payload.SecurityVulnerability.Advisory.Summary : ""`),
		entityops.VulnerabilityFields.Description.Expr(`paragraphs('SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'Description' in payload.SecurityVulnerability.Advisory ? payload.SecurityVulnerability.Advisory.Description : "")`),
		entityops.VulnerabilityFields.CveID.Expr(`'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'Identifiers' in payload.SecurityVulnerability.Advisory && payload.SecurityVulnerability.Advisory.Identifiers.filter(i, i.Type == "CVE").size() > 0 ? payload.SecurityVulnerability.Advisory.Identifiers.filter(i, i.Type == "CVE")[0].Value : ""`),
		entityops.VulnerabilityFields.Score.Expr(`'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'CvssSeverities' in payload.SecurityVulnerability.Advisory && 'CvssV4' in payload.SecurityVulnerability.Advisory.CvssSeverities && payload.SecurityVulnerability.Advisory.CvssSeverities.CvssV4 != null ? payload.SecurityVulnerability.Advisory.CvssSeverities.CvssV4.Score : null`),
		entityops.VulnerabilityFields.Vector.Expr(`'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'CvssSeverities' in payload.SecurityVulnerability.Advisory && 'CvssV4' in payload.SecurityVulnerability.Advisory.CvssSeverities && payload.SecurityVulnerability.Advisory.CvssSeverities.CvssV4 != null ? payload.SecurityVulnerability.Advisory.CvssSeverities.CvssV4.VectorString : ""`),
		entityops.VulnerabilityFields.CweIds.Expr(`'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'Cwes' in payload.SecurityVulnerability.Advisory && 'Nodes' in payload.SecurityVulnerability.Advisory.Cwes ? payload.SecurityVulnerability.Advisory.Cwes.Nodes.map(c, c.CweID) : []`),
		entityops.VulnerabilityFields.References.Expr(`'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'References' in payload.SecurityVulnerability.Advisory ? payload.SecurityVulnerability.Advisory.References.map(r, r.URL) : []`),
		entityops.VulnerabilityFields.VulnerableVersionRange.Expr(`'SecurityVulnerability' in payload && 'VulnerableVersionRange' in payload.SecurityVulnerability ? payload.SecurityVulnerability.VulnerableVersionRange : ""`),
		entityops.VulnerabilityFields.FirstPatchedVersion.Expr(`'SecurityVulnerability' in payload && 'FirstPatchedVersion' in payload.SecurityVulnerability && 'Identifier' in payload.SecurityVulnerability.FirstPatchedVersion ? payload.SecurityVulnerability.FirstPatchedVersion.Identifier : ""`),
		entityops.VulnerabilityFields.PackageName.Expr(`'SecurityVulnerability' in payload && 'Package' in payload.SecurityVulnerability && 'Name' in payload.SecurityVulnerability.Package ? payload.SecurityVulnerability.Package.Name : ""`),
		entityops.VulnerabilityFields.PackageEcosystem.Expr(`'SecurityVulnerability' in payload && 'Package' in payload.SecurityVulnerability && 'Ecosystem' in payload.SecurityVulnerability.Package ? payload.SecurityVulnerability.Package.Ecosystem : ""`),
		entityops.VulnerabilityFields.ManifestPath.Expr(`'VulnerableManifestPath' in payload ? payload.VulnerableManifestPath : ""`),
		entityops.VulnerabilityFields.DependencyScope.Expr(`'DependencyScope' in payload ? payload.DependencyScope : ""`),
	},
	)
	// mapExprCodeScanning is the CEL mapping expression for code scanning webhook alert payloads
	mapExprCodeScanning = buildWebhookMappingExpr(githubAlertTypeCodeScanning, `"github:" + resource + ":code_scanning:" + ('number' in payload && payload.number != 0 ? string(payload.number) : "unknown")`, []entityops.MappingEntry{
		entityops.VulnerabilityFields.Severity.Expr(`'rule' in payload && 'security_severity_level' in payload.rule && payload.rule.security_severity_level != "" ? payload.rule.security_severity_level : ('rule' in payload && 'severity' in payload.rule ? payload.rule.severity : "")`),
		entityops.VulnerabilityFields.Summary.Expr(`'rule' in payload && 'description' in payload.rule && payload.rule.description != "" ? payload.rule.description : ('rule' in payload && 'name' in payload.rule ? payload.rule.name : "")`),
		entityops.VulnerabilityFields.Description.Expr(`paragraphs('most_recent_instance' in payload && 'message' in payload.most_recent_instance && 'text' in payload.most_recent_instance.message ? payload.most_recent_instance.message.text : "")`),
	},
	)
	// mapExprSecretScanning is the CEL mapping expression for secret scanning webhook alert payloads
	mapExprSecretScanning = buildWebhookMappingExpr(githubAlertTypeSecretScan, `"github:" + resource + ":secret_scanning:" + ('number' in payload && payload.number != 0 ? string(payload.number) : "unknown")`, []entityops.MappingEntry{
		entityops.VulnerabilityFields.Severity.Expr(`"high"`),
		entityops.VulnerabilityFields.Summary.Expr(`'secret_type_display_name' in payload && payload.secret_type_display_name != "" ? payload.secret_type_display_name : ('secret_type' in payload ? payload.secret_type : "")`),
		entityops.VulnerabilityFields.Description.Expr(`paragraphs('resolution' in payload ? payload.resolution : "")`),
	},
	)
)

// mapExprRepositoryAsset is the CEL mapping expression for GitHub repository payloads mapped to Asset
var mapExprRepositoryAsset = providerkit.CelMapExpr(
	entityops.AssetFields.SourceIdentifier.Expr("payload.NameWithOwner"),
	entityops.AssetFields.DisplayName.Expr("payload.NameWithOwner"),
	entityops.AssetFields.Name.Expr("payload.NameWithOwner"),
	entityops.AssetFields.AssetType.Expr(`"REPOSITORY"`),
	entityops.AssetFields.Website.Expr("payload.URL"),
	entityops.AssetFields.ObservedAt.Expr("payload.UpdatedAt"),
	entityops.AssetFields.Categories.Expr(`payload.IsPrivate ? ["private", "repository"] : ["public", "repository"]`),
)

// mapExprDirectoryAccount is the CEL mapping expression for GitHub organization member payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr(
	entityops.DirectoryAccountFields.ExternalID.Expr(`payload.DatabaseID != 0 ? string(payload.DatabaseID) : payload.Login`),
	entityops.DirectoryAccountFields.CanonicalEmail.Expr(`payload.CanonicalEmail != "" ? payload.CanonicalEmail : payload.Login`),
	entityops.DirectoryAccountFields.EmailAliases.Expr(`'EmailAliases' in payload && payload.EmailAliases != null ? payload.EmailAliases : []`),
	entityops.DirectoryAccountFields.DisplayName.Expr(`payload.Name != "" ? payload.Name : payload.Login`),
	entityops.DirectoryAccountFields.AvatarRemoteURL.Expr(`payload.AvatarURL`),
	entityops.DirectoryAccountFields.GivenName.Expr(`payload.GivenName`),
	entityops.DirectoryAccountFields.FamilyName.Expr(`payload.FamilyName`),
	entityops.DirectoryAccountFields.Profile.Expr("payload"),
	entityops.DirectoryAccountFields.DirectoryName.Expr("installation.name"),
)

// mapExprDirectoryGroup is the CEL mapping expression for GitHub teams mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr(
	entityops.DirectoryGroupFields.ExternalID.Expr(`payload.DatabaseID != 0 ? string(payload.DatabaseID) : payload.Slug`),
	entityops.DirectoryGroupFields.DisplayName.Expr(`payload.Name != "" ? payload.Name : payload.Slug`),
	entityops.DirectoryGroupFields.Classification.Expr(`dyn("TEAM")`),
	entityops.DirectoryGroupFields.Status.Expr(`dyn("ACTIVE")`),
	entityops.DirectoryGroupFields.Profile.Expr("payload"),
	entityops.DirectoryGroupFields.DirectoryName.Expr("installation.name"),
)

// mapExprDirectoryMembership is the CEL mapping expression for GitHub team memberships mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr(
	entityops.DirectoryMembershipFields.DirectoryAccountID.Expr(`payload.Member.DatabaseID != 0 ? string(payload.Member.DatabaseID) : payload.Member.Login`),
	entityops.DirectoryMembershipFields.DirectoryGroupID.Expr(`payload.Team.DatabaseID != 0 ? string(payload.Team.DatabaseID) : payload.Team.Slug`),
	entityops.DirectoryMembershipFields.Role.Expr(`dyn(payload.Role != "" ? payload.Role : "MEMBER")`),
	entityops.DirectoryMembershipFields.Metadata.Expr("payload"),
	entityops.DirectoryMembershipFields.DirectoryName.Expr("installation.name"),
)
