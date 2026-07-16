package githubapp

import (
	"strconv"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// webhookBaseEntries returns CEL map entries for webhook alert payloads (snake_case JSON keys)
func webhookBaseEntries(category, externalIDExpr string) []providerkit.CelMapEntry {
	return []providerkit.CelMapEntry{
		{Key: entityops.InputKeyVulnerabilityExternalID, Expr: externalIDExpr},
		{Key: entityops.InputKeyVulnerabilityExternalOwnerID, Expr: "resource"},
		{Key: entityops.InputKeyVulnerabilityCategory, Expr: strconv.Quote(category)},
		{Key: entityops.InputKeyVulnerabilityVulnerabilityStatusName, Expr: `'state' in payload ? payload.state : ""`},
		{Key: entityops.InputKeyVulnerabilityExternalURI, Expr: `'html_url' in payload ? payload.html_url : ""`},
		{Key: entityops.InputKeyVulnerabilitySourceUpdatedAt, Expr: `'updated_at' in payload ? payload.updated_at : null`},
		{Key: entityops.InputKeyVulnerabilityDiscoveredAt, Expr: `'created_at' in payload ? payload.created_at : null`},
		{Key: entityops.InputKeyVulnerabilityPublishedAt, Expr: `'security_advisory' in payload && 'published_at' in payload.security_advisory ? payload.security_advisory.published_at : null`},
		{Key: entityops.InputKeyVulnerabilityDismissedAt, Expr: `'dismissed_at' in payload ? payload.dismissed_at : null`},
		{Key: entityops.InputKeyVulnerabilityDismissedReason, Expr: `'dismissed_reason' in payload ? payload.dismissed_reason : null`},
		{Key: entityops.InputKeyVulnerabilityDismissedComment, Expr: `'dismissed_comment' in payload ? payload.dismissed_comment : null`},
		{Key: entityops.InputKeyVulnerabilityFixedAt, Expr: `'fixed_at' in payload ? payload.fixed_at : null`},
		{Key: entityops.InputKeyVulnerabilityAutoDismissedAt, Expr: `'auto_dismissed_at' in payload ? payload.auto_dismissed_at : null`},
		{Key: entityops.InputKeyVulnerabilityOpen, Expr: `'state' in payload ? payload.state == "open" : false`},
		{Key: entityops.InputKeyVulnerabilityRawPayload, Expr: "payload"},
	}
}

// pollBaseEntries returns CEL map entries for GraphQL poll payloads (PascalCase JSON keys)
func pollBaseEntries(category, externalIDExpr string) []providerkit.CelMapEntry {
	return []providerkit.CelMapEntry{
		{Key: entityops.InputKeyVulnerabilityExternalID, Expr: externalIDExpr},
		{Key: entityops.InputKeyVulnerabilityExternalOwnerID, Expr: "resource"},
		{Key: entityops.InputKeyVulnerabilityCategory, Expr: strconv.Quote(category)},
		{Key: entityops.InputKeyVulnerabilityVulnerabilityStatusName, Expr: `'State' in payload ? payload.State : ""`},
		{Key: entityops.InputKeyVulnerabilityExternalURI, Expr: `'Number' in payload && payload.Number != 0 ? "https://github.com/" + resource + "/security/dependabot/" + string(payload.Number) : ""`},
		{Key: entityops.InputKeyVulnerabilitySourceUpdatedAt, Expr: `'SecurityVulnerability' in payload && 'UpdatedAt' in payload.SecurityVulnerability ? payload.SecurityVulnerability.UpdatedAt : null`},
		{Key: entityops.InputKeyVulnerabilityDiscoveredAt, Expr: `'CreatedAt' in payload ? payload.CreatedAt : null`},
		{Key: entityops.InputKeyVulnerabilityPublishedAt, Expr: `'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'PublishedAt' in payload.SecurityVulnerability.Advisory ? payload.SecurityVulnerability.Advisory.PublishedAt : null`},
		{Key: entityops.InputKeyVulnerabilityDismissedAt, Expr: `'DismissedAt' in payload ? payload.DismissedAt : null`},
		{Key: entityops.InputKeyVulnerabilityDismissedReason, Expr: `'DismissReason' in payload ? payload.DismissReason : null`},
		{Key: entityops.InputKeyVulnerabilityDismissedComment, Expr: `'DismissComment' in payload ? payload.DismissComment : null`},
		{Key: entityops.InputKeyVulnerabilityFixedAt, Expr: `'FixedAt' in payload ? payload.FixedAt : null`},
		{Key: entityops.InputKeyVulnerabilityAutoDismissedAt, Expr: `'AutoDismissedAt' in payload ? payload.AutoDismissedAt : null`},
		{Key: entityops.InputKeyVulnerabilityOpen, Expr: `'State' in payload ? payload.State == "OPEN" : false`},
		{Key: entityops.InputKeyVulnerabilityRawPayload, Expr: "payload"},
	}
}

// buildWebhookMappingExpr constructs a CEL mapping expression for webhook payloads
func buildWebhookMappingExpr(category, externalIDExpr string, extras []providerkit.CelMapEntry) string {
	entries := webhookBaseEntries(category, externalIDExpr)
	entries = append(entries, extras...)

	return providerkit.CelMapExpr(entries)
}

// buildPollMappingExpr constructs a CEL mapping expression for GraphQL poll payloads
func buildPollMappingExpr(category, externalIDExpr string, extras []providerkit.CelMapEntry) string {
	entries := pollBaseEntries(category, externalIDExpr)
	entries = append(entries, extras...)

	return providerkit.CelMapExpr(entries)
}

var (
	// mapExprDependabot is the CEL mapping expression for Dependabot webhook alert payloads
	mapExprDependabot = buildWebhookMappingExpr(githubAlertTypeDependabot, `"github:" + resource + ":dependabot:" + ('number' in payload && payload.number != 0 ? string(payload.number) : ('security_advisory' in payload && 'ghsa_id' in payload.security_advisory && payload.security_advisory.ghsa_id != "" ? payload.security_advisory.ghsa_id : "unknown"))`, []providerkit.CelMapEntry{
		{Key: entityops.InputKeyVulnerabilitySeverity, Expr: `'security_advisory' in payload && 'severity' in payload.security_advisory ? payload.security_advisory.severity : ""`},
		{Key: entityops.InputKeyVulnerabilitySummary, Expr: `'security_advisory' in payload && 'summary' in payload.security_advisory ? payload.security_advisory.summary : ""`},
		{Key: entityops.InputKeyVulnerabilityDescription, Expr: `'security_advisory' in payload && 'description' in payload.security_advisory ? payload.security_advisory.description : ""`},
		{Key: entityops.InputKeyVulnerabilityCveID, Expr: `'security_advisory' in payload && 'cve_id' in payload.security_advisory ? payload.security_advisory.cve_id : ""`},
		{Key: entityops.InputKeyVulnerabilityDisplayName, Expr: `'security_advisory' in payload && 'cve_id' in payload.security_advisory ? payload.security_advisory.cve_id : ""`},
		{Key: entityops.InputKeyVulnerabilityScore, Expr: `'security_advisory' in payload && 'cvss' in payload.security_advisory && 'score' in payload.security_advisory.cvss ? payload.security_advisory.cvss.score : null`},
		{Key: entityops.InputKeyVulnerabilityVector, Expr: `'security_advisory' in payload && 'cvss' in payload.security_advisory && 'vector_string' in payload.security_advisory.cvss ? payload.security_advisory.cvss.vector_string : ""`},
		{Key: entityops.InputKeyVulnerabilityCweIds, Expr: `'security_advisory' in payload && 'cwes' in payload.security_advisory ? payload.security_advisory.cwes.map(c, c.cwe_id) : []`},
		{Key: entityops.InputKeyVulnerabilityReferences, Expr: `'security_advisory' in payload && 'references' in payload.security_advisory ? payload.security_advisory.references.map(r, r.url) : []`},
		{Key: entityops.InputKeyVulnerabilityVulnerableVersionRange, Expr: `'security_vulnerability' in payload && 'vulnerable_version_range' in payload.security_vulnerability ? payload.security_vulnerability.vulnerable_version_range : ""`},
		{Key: entityops.InputKeyVulnerabilityFirstPatchedVersion, Expr: `'security_vulnerability' in payload && 'first_patched_version' in payload.security_vulnerability && 'identifier' in payload.security_vulnerability.first_patched_version ? payload.security_vulnerability.first_patched_version.identifier : ""`},
		{Key: entityops.InputKeyVulnerabilityPackageName, Expr: `'security_vulnerability' in payload && 'package' in payload.security_vulnerability && 'name' in payload.security_vulnerability.package ? payload.security_vulnerability.package.name : ""`},
		{Key: entityops.InputKeyVulnerabilityPackageEcosystem, Expr: `'security_vulnerability' in payload && 'package' in payload.security_vulnerability && 'ecosystem' in payload.security_vulnerability.package ? payload.security_vulnerability.package.ecosystem : ""`},
		{Key: entityops.InputKeyVulnerabilityManifestPath, Expr: `'dependency' in payload && 'manifest_path' in payload.dependency ? payload.dependency.manifest_path : ""`},
		{Key: entityops.InputKeyVulnerabilityDependencyScope, Expr: `'dependency' in payload && 'scope' in payload.dependency ? payload.dependency.scope : ""`},
	},
	)
	// mapExprDependabotPoll is the CEL mapping expression for Dependabot alerts collected via GraphQL poll
	mapExprDependabotPoll = buildPollMappingExpr(githubAlertTypeDependabot, `"github:" + resource + ":dependabot:" + ('Number' in payload && payload.Number != 0 ? string(payload.Number) : ('SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'GHSAID' in payload.SecurityVulnerability.Advisory && payload.SecurityVulnerability.Advisory.GHSAID != "" ? payload.SecurityVulnerability.Advisory.GHSAID : "unknown"))`, []providerkit.CelMapEntry{
		{Key: entityops.InputKeyVulnerabilityDisplayName, Expr: `'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'Identifiers' in payload.SecurityVulnerability.Advisory && payload.SecurityVulnerability.Advisory.Identifiers.filter(i, i.Type == "CVE").size() > 0 ? payload.SecurityVulnerability.Advisory.Identifiers.filter(i, i.Type == "CVE")[0].Value : 'GHSAID' in payload.SecurityVulnerability.Advisory ? payload.SecurityVulnerability.Advisory.GHSAID : ""`},
		{Key: entityops.InputKeyVulnerabilitySeverity, Expr: `'SecurityVulnerability' in payload && 'Severity' in payload.SecurityVulnerability ? payload.SecurityVulnerability.Severity : ""`},
		{Key: entityops.InputKeyVulnerabilitySummary, Expr: `'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'Summary' in payload.SecurityVulnerability.Advisory ? payload.SecurityVulnerability.Advisory.Summary : ""`},
		{Key: entityops.InputKeyVulnerabilityDescription, Expr: `'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'Description' in payload.SecurityVulnerability.Advisory ? payload.SecurityVulnerability.Advisory.Description : ""`},
		{Key: entityops.InputKeyVulnerabilityCveID, Expr: `'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'Identifiers' in payload.SecurityVulnerability.Advisory && payload.SecurityVulnerability.Advisory.Identifiers.filter(i, i.Type == "CVE").size() > 0 ? payload.SecurityVulnerability.Advisory.Identifiers.filter(i, i.Type == "CVE")[0].Value : ""`},
		{Key: entityops.InputKeyVulnerabilityScore, Expr: `'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'CvssSeverities' in payload.SecurityVulnerability.Advisory && 'CvssV4' in payload.SecurityVulnerability.Advisory.CvssSeverities && payload.SecurityVulnerability.Advisory.CvssSeverities.CvssV4 != null ? payload.SecurityVulnerability.Advisory.CvssSeverities.CvssV4.Score : null`},
		{Key: entityops.InputKeyVulnerabilityVector, Expr: `'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'CvssSeverities' in payload.SecurityVulnerability.Advisory && 'CvssV4' in payload.SecurityVulnerability.Advisory.CvssSeverities && payload.SecurityVulnerability.Advisory.CvssSeverities.CvssV4 != null ? payload.SecurityVulnerability.Advisory.CvssSeverities.CvssV4.VectorString : ""`},
		{Key: entityops.InputKeyVulnerabilityCweIds, Expr: `'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'Cwes' in payload.SecurityVulnerability.Advisory && 'Nodes' in payload.SecurityVulnerability.Advisory.Cwes ? payload.SecurityVulnerability.Advisory.Cwes.Nodes.map(c, c.CweID) : []`},
		{Key: entityops.InputKeyVulnerabilityReferences, Expr: `'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'References' in payload.SecurityVulnerability.Advisory ? payload.SecurityVulnerability.Advisory.References.map(r, r.URL) : []`},
		{Key: entityops.InputKeyVulnerabilityVulnerableVersionRange, Expr: `'SecurityVulnerability' in payload && 'VulnerableVersionRange' in payload.SecurityVulnerability ? payload.SecurityVulnerability.VulnerableVersionRange : ""`},
		{Key: entityops.InputKeyVulnerabilityFirstPatchedVersion, Expr: `'SecurityVulnerability' in payload && 'FirstPatchedVersion' in payload.SecurityVulnerability && 'Identifier' in payload.SecurityVulnerability.FirstPatchedVersion ? payload.SecurityVulnerability.FirstPatchedVersion.Identifier : ""`},
		{Key: entityops.InputKeyVulnerabilityPackageName, Expr: `'SecurityVulnerability' in payload && 'Package' in payload.SecurityVulnerability && 'Name' in payload.SecurityVulnerability.Package ? payload.SecurityVulnerability.Package.Name : ""`},
		{Key: entityops.InputKeyVulnerabilityPackageEcosystem, Expr: `'SecurityVulnerability' in payload && 'Package' in payload.SecurityVulnerability && 'Ecosystem' in payload.SecurityVulnerability.Package ? payload.SecurityVulnerability.Package.Ecosystem : ""`},
		{Key: entityops.InputKeyVulnerabilityManifestPath, Expr: `'VulnerableManifestPath' in payload ? payload.VulnerableManifestPath : ""`},
		{Key: entityops.InputKeyVulnerabilityDependencyScope, Expr: `'DependencyScope' in payload ? payload.DependencyScope : ""`},
	},
	)
	// mapExprCodeScanning is the CEL mapping expression for code scanning webhook alert payloads
	mapExprCodeScanning = buildWebhookMappingExpr(githubAlertTypeCodeScanning, `"github:" + resource + ":code_scanning:" + ('number' in payload && payload.number != 0 ? string(payload.number) : "unknown")`, []providerkit.CelMapEntry{
		{Key: entityops.InputKeyVulnerabilitySeverity, Expr: `'rule' in payload && 'security_severity_level' in payload.rule && payload.rule.security_severity_level != "" ? payload.rule.security_severity_level : ('rule' in payload && 'severity' in payload.rule ? payload.rule.severity : "")`},
		{Key: entityops.InputKeyVulnerabilitySummary, Expr: `'rule' in payload && 'description' in payload.rule && payload.rule.description != "" ? payload.rule.description : ('rule' in payload && 'name' in payload.rule ? payload.rule.name : "")`},
		{Key: entityops.InputKeyVulnerabilityDescription, Expr: `'most_recent_instance' in payload && 'message' in payload.most_recent_instance && 'text' in payload.most_recent_instance.message ? payload.most_recent_instance.message.text : ""`},
	},
	)
	// mapExprSecretScanning is the CEL mapping expression for secret scanning webhook alert payloads
	mapExprSecretScanning = buildWebhookMappingExpr(githubAlertTypeSecretScan, `"github:" + resource + ":secret_scanning:" + ('number' in payload && payload.number != 0 ? string(payload.number) : "unknown")`, []providerkit.CelMapEntry{
		{Key: entityops.InputKeyVulnerabilitySeverity, Expr: `"high"`},
		{Key: entityops.InputKeyVulnerabilitySummary, Expr: `'secret_type_display_name' in payload && payload.secret_type_display_name != "" ? payload.secret_type_display_name : ('secret_type' in payload ? payload.secret_type : "")`},
		{Key: entityops.InputKeyVulnerabilityDescription, Expr: `'resolution' in payload ? payload.resolution : ""`},
	},
	)
)

// mapExprRepositoryAsset is the CEL mapping expression for GitHub repository payloads mapped to Asset
var mapExprRepositoryAsset = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyAssetSourceIdentifier, Expr: "payload.NameWithOwner"},
	{Key: entityops.InputKeyAssetDisplayName, Expr: "payload.NameWithOwner"},
	{Key: entityops.InputKeyAssetName, Expr: "payload.NameWithOwner"},
	{Key: entityops.InputKeyAssetAssetType, Expr: `"REPOSITORY"`},
	{Key: entityops.InputKeyAssetWebsite, Expr: "payload.URL"},
	{Key: entityops.InputKeyAssetObservedAt, Expr: "payload.UpdatedAt"},
	{Key: entityops.InputKeyAssetCategories, Expr: `payload.IsPrivate ? ["private", "repository"] : ["public", "repository"]`},
})

// mapExprDirectoryAccount is the CEL mapping expression for GitHub organization member payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryAccountExternalID, Expr: `payload.DatabaseID != 0 ? string(payload.DatabaseID) : payload.Login`},
	{Key: entityops.InputKeyDirectoryAccountCanonicalEmail, Expr: `payload.CanonicalEmail != "" ? payload.CanonicalEmail : payload.Login`},
	{Key: entityops.InputKeyDirectoryAccountEmailAliases, Expr: `'EmailAliases' in payload && payload.EmailAliases != null ? payload.EmailAliases : []`},
	{Key: entityops.InputKeyDirectoryAccountDisplayName, Expr: `payload.Name != "" ? payload.Name : payload.Login`},
	{Key: entityops.InputKeyDirectoryAccountAvatarRemoteURL, Expr: `payload.AvatarURL`},
	{Key: entityops.InputKeyDirectoryAccountDirectoryInstanceID, Expr: `payload.Org`},
	{Key: entityops.InputKeyDirectoryAccountGivenName, Expr: `payload.GivenName`},
	{Key: entityops.InputKeyDirectoryAccountFamilyName, Expr: `payload.FamilyName`},
	{Key: entityops.InputKeyDirectoryAccountProfile, Expr: "payload"},
})

// mapExprDirectoryGroup is the CEL mapping expression for GitHub teams mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryGroupExternalID, Expr: `payload.DatabaseID != 0 ? string(payload.DatabaseID) : payload.Slug`},
	{Key: entityops.InputKeyDirectoryGroupDisplayName, Expr: `payload.Name != "" ? payload.Name : payload.Slug`},
	{Key: entityops.InputKeyDirectoryGroupDirectoryInstanceID, Expr: `payload.Org`},
	{Key: entityops.InputKeyDirectoryGroupClassification, Expr: `dyn("TEAM")`},
	{Key: entityops.InputKeyDirectoryGroupStatus, Expr: `dyn("ACTIVE")`},
	{Key: entityops.InputKeyDirectoryGroupProfile, Expr: "payload"},
})

// mapExprDirectoryMembership is the CEL mapping expression for GitHub team memberships mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryMembershipDirectoryAccountID, Expr: `payload.Member.DatabaseID != 0 ? string(payload.Member.DatabaseID) : payload.Member.Login`},
	{Key: entityops.InputKeyDirectoryMembershipDirectoryGroupID, Expr: `payload.Team.DatabaseID != 0 ? string(payload.Team.DatabaseID) : payload.Team.Slug`},
	{Key: entityops.InputKeyDirectoryMembershipDirectoryInstanceID, Expr: `payload.Org`},
	{Key: entityops.InputKeyDirectoryMembershipRole, Expr: `dyn(payload.Role != "" ? payload.Role : "MEMBER")`},
	{Key: entityops.InputKeyDirectoryMembershipMetadata, Expr: "payload"},
})

// githubAppMappings returns all built-in ingest mappings for the GitHub App definition
func githubAppMappings() []types.MappingRegistration {
	overrides := map[string]types.MappingOverride{
		githubAlertTypeDependabot: {
			FilterExpr: "true",
			MapExpr:    mapExprDependabot,
		},
		githubAlertTypeDependabotPoll: {
			FilterExpr: "true",
			MapExpr:    mapExprDependabotPoll,
		},
		githubAlertTypeCodeScanning: {
			FilterExpr: "true",
			MapExpr:    mapExprCodeScanning,
		},
		githubAlertTypeSecretScan: {
			FilterExpr: "true",
			MapExpr:    mapExprSecretScanning,
		},
	}

	vulnMappings := lo.MapToSlice(overrides, func(variant string, override types.MappingOverride) types.MappingRegistration {
		return types.MappingRegistration{
			Schema:  entityops.SchemaVulnerability.Name,
			Variant: variant,
			Spec:    override,
		}
	})

	return append(vulnMappings,
		types.MappingRegistration{
			Schema:  entityops.SchemaAsset.Name,
			Variant: repositoryAssetVariant,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprRepositoryAsset,
			},
		},
		types.MappingRegistration{
			Schema: entityops.SchemaDirectoryAccount.Name,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryAccount,
			},
		},
		types.MappingRegistration{
			Schema: entityops.SchemaDirectoryGroup.Name,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryGroup,
			},
		},
		types.MappingRegistration{
			Schema: entityops.SchemaDirectoryMembership.Name,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryMembership,
			},
		},
	)
}
