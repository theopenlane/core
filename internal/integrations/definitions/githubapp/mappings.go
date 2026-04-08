package githubapp

import (
	"strconv"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// webhookBaseEntries returns CEL map entries for webhook alert payloads (snake_case JSON keys)
func webhookBaseEntries(category, externalIDExpr string) []providerkit.CelMapEntry {
	return []providerkit.CelMapEntry{
		{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalID, Expr: externalIDExpr},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalOwnerID, Expr: "resource"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySource, Expr: `"github"`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityCategory, Expr: strconv.Quote(category)},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityVulnerabilityStatusName, Expr: `'state' in payload ? payload.state : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalURI, Expr: `'html_url' in payload ? payload.html_url : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt, Expr: `'updated_at' in payload ? payload.updated_at : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt, Expr: `'created_at' in payload ? payload.created_at : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityPublishedAt, Expr: `'security_advisory' in payload && 'published_at' in payload.security_advisory ? payload.security_advisory.published_at : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDismissedAt, Expr: `'dismissed_at' in payload ? payload.dismissed_at : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDismissedReason, Expr: `'dismissed_reason' in payload ? payload.dismissed_reason : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDismissedComment, Expr: `'dismissed_comment' in payload ? payload.dismissed_comment : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityFixedAt, Expr: `'fixed_at' in payload ? payload.fixed_at : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityAutoDismissedAt, Expr: `'auto_dismissed_at' in payload ? payload.auto_dismissed_at : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityOpen, Expr: `'state' in payload ? payload.state == "open" : false`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityRawPayload, Expr: "payload"},
	}
}

// pollBaseEntries returns CEL map entries for GraphQL poll payloads (PascalCase JSON keys)
func pollBaseEntries(category, externalIDExpr string) []providerkit.CelMapEntry {
	return []providerkit.CelMapEntry{
		{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalID, Expr: externalIDExpr},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalOwnerID, Expr: "resource"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySource, Expr: `"github"`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityCategory, Expr: strconv.Quote(category)},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityVulnerabilityStatusName, Expr: `'State' in payload ? payload.State : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalURI, Expr: `'URL' in payload ? payload.URL : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt, Expr: `'UpdatedAt' in payload ? payload.UpdatedAt : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt, Expr: `'CreatedAt' in payload ? payload.CreatedAt : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityPublishedAt, Expr: `'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'PublishedAt' in payload.SecurityVulnerability.Advisory ? payload.SecurityVulnerability.Advisory.PublishedAt : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDismissedAt, Expr: `'DismissedAt' in payload ? payload.DismissedAt : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDismissedReason, Expr: `'DismissReason' in payload ? payload.DismissReason : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDismissedComment, Expr: `'DismissComment' in payload ? payload.DismissComment : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityFixedAt, Expr: `'FixedAt' in payload ? payload.FixedAt : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityAutoDismissedAt, Expr: `'AutoDismissedAt' in payload ? payload.AutoDismissedAt : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityOpen, Expr: `'State' in payload ? payload.State == "OPEN" : false`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityRawPayload, Expr: "payload"},
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
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, Expr: `'security_advisory' in payload && 'severity' in payload.security_advisory ? payload.security_advisory.severity : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySummary, Expr: `'security_advisory' in payload && 'summary' in payload.security_advisory ? payload.security_advisory.summary : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDescription, Expr: `'security_advisory' in payload && 'description' in payload.security_advisory ? payload.security_advisory.description : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityCveID, Expr: `'security_advisory' in payload && 'cve_id' in payload.security_advisory ? payload.security_advisory.cve_id : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDisplayName, Expr: `'security_advisory' in payload && 'cve_id' in payload.security_advisory ? payload.security_advisory.cve_id : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityScore, Expr: `'security_advisory' in payload && 'cvss' in payload.security_advisory && 'score' in payload.security_advisory.cvss ? payload.security_advisory.cvss.score : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityVector, Expr: `'security_advisory' in payload && 'cvss' in payload.security_advisory && 'vector_string' in payload.security_advisory.cvss ? payload.security_advisory.cvss.vector_string : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityCweIds, Expr: `'security_advisory' in payload && 'cwes' in payload.security_advisory ? payload.security_advisory.cwes.map(c, c.cwe_id) : []`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityReferences, Expr: `'security_advisory' in payload && 'references' in payload.security_advisory ? payload.security_advisory.references.map(r, r.url) : []`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityVulnerableVersionRange, Expr: `'security_vulnerability' in payload && 'vulnerable_version_range' in payload.security_vulnerability ? payload.security_vulnerability.vulnerable_version_range : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityFirstPatchedVersion, Expr: `'security_vulnerability' in payload && 'first_patched_version' in payload.security_vulnerability && 'identifier' in payload.security_vulnerability.first_patched_version ? payload.security_vulnerability.first_patched_version.identifier : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityPackageName, Expr: `'security_vulnerability' in payload && 'package' in payload.security_vulnerability && 'name' in payload.security_vulnerability.package ? payload.security_vulnerability.package.name : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityPackageEcosystem, Expr: `'security_vulnerability' in payload && 'package' in payload.security_vulnerability && 'ecosystem' in payload.security_vulnerability.package ? payload.security_vulnerability.package.ecosystem : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityManifestPath, Expr: `'dependency' in payload && 'manifest_path' in payload.dependency ? payload.dependency.manifest_path : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDependencyScope, Expr: `'dependency' in payload && 'scope' in payload.dependency ? payload.dependency.scope : ""`},
	},
	)
	// mapExprDependabotPoll is the CEL mapping expression for Dependabot alerts collected via GraphQL poll
	mapExprDependabotPoll = buildPollMappingExpr(githubAlertTypeDependabot, `"github:" + resource + ":dependabot:" + ('Number' in payload && payload.Number != 0 ? string(payload.Number) : ('SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'GHSAID' in payload.SecurityVulnerability.Advisory && payload.SecurityVulnerability.Advisory.GHSAID != "" ? payload.SecurityVulnerability.Advisory.GHSAID : "unknown"))`, []providerkit.CelMapEntry{
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, Expr: `'SecurityVulnerability' in payload && 'Severity' in payload.SecurityVulnerability ? payload.SecurityVulnerability.Severity : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySummary, Expr: `'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'Summary' in payload.SecurityVulnerability.Advisory ? payload.SecurityVulnerability.Advisory.Summary : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDescription, Expr: `'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'Description' in payload.SecurityVulnerability.Advisory ? payload.SecurityVulnerability.Advisory.Description : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityCveID, Expr: `'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'Identifiers' in payload.SecurityVulnerability.Advisory && payload.SecurityVulnerability.Advisory.Identifiers.filter(i, i.Type == "CVE").size() > 0 ? payload.SecurityVulnerability.Advisory.Identifiers.filter(i, i.Type == "CVE")[0].Value : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityScore, Expr: `'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'Cvss' in payload.SecurityVulnerability.Advisory && 'Score' in payload.SecurityVulnerability.Advisory.Cvss ? payload.SecurityVulnerability.Advisory.Cvss.Score : null`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityVector, Expr: `'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'Cvss' in payload.SecurityVulnerability.Advisory && 'VectorString' in payload.SecurityVulnerability.Advisory.Cvss ? payload.SecurityVulnerability.Advisory.Cvss.VectorString : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityCweIds, Expr: `'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'Cwes' in payload.SecurityVulnerability.Advisory && 'Nodes' in payload.SecurityVulnerability.Advisory.Cwes ? payload.SecurityVulnerability.Advisory.Cwes.Nodes.map(c, c.CweID) : []`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityReferences, Expr: `'SecurityVulnerability' in payload && 'Advisory' in payload.SecurityVulnerability && 'References' in payload.SecurityVulnerability.Advisory ? payload.SecurityVulnerability.Advisory.References.map(r, r.URL) : []`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityVulnerableVersionRange, Expr: `'SecurityVulnerability' in payload && 'VulnerableVersionRange' in payload.SecurityVulnerability ? payload.SecurityVulnerability.VulnerableVersionRange : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityFirstPatchedVersion, Expr: `'SecurityVulnerability' in payload && 'FirstPatchedVersion' in payload.SecurityVulnerability && 'Identifier' in payload.SecurityVulnerability.FirstPatchedVersion ? payload.SecurityVulnerability.FirstPatchedVersion.Identifier : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityPackageName, Expr: `'SecurityVulnerability' in payload && 'Package' in payload.SecurityVulnerability && 'Name' in payload.SecurityVulnerability.Package ? payload.SecurityVulnerability.Package.Name : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityPackageEcosystem, Expr: `'SecurityVulnerability' in payload && 'Package' in payload.SecurityVulnerability && 'Ecosystem' in payload.SecurityVulnerability.Package ? payload.SecurityVulnerability.Package.Ecosystem : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityManifestPath, Expr: `'VulnerableManifestPath' in payload ? payload.VulnerableManifestPath : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDependencyScope, Expr: `'DependencyScope' in payload ? payload.DependencyScope : ""`},
	},
	)
	// mapExprCodeScanning is the CEL mapping expression for code scanning webhook alert payloads
	mapExprCodeScanning = buildWebhookMappingExpr(githubAlertTypeCodeScanning, `"github:" + resource + ":code_scanning:" + ('number' in payload && payload.number != 0 ? string(payload.number) : "unknown")`, []providerkit.CelMapEntry{
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, Expr: `'rule' in payload && 'security_severity_level' in payload.rule && payload.rule.security_severity_level != "" ? payload.rule.security_severity_level : ('rule' in payload && 'severity' in payload.rule ? payload.rule.severity : "")`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySummary, Expr: `'rule' in payload && 'description' in payload.rule && payload.rule.description != "" ? payload.rule.description : ('rule' in payload && 'name' in payload.rule ? payload.rule.name : "")`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDescription, Expr: `'most_recent_instance' in payload && 'message' in payload.most_recent_instance && 'text' in payload.most_recent_instance.message ? payload.most_recent_instance.message.text : ""`},
	},
	)
	// mapExprSecretScanning is the CEL mapping expression for secret scanning webhook alert payloads
	mapExprSecretScanning = buildWebhookMappingExpr(githubAlertTypeSecretScan, `"github:" + resource + ":secret_scanning:" + ('number' in payload && payload.number != 0 ? string(payload.number) : "unknown")`, []providerkit.CelMapEntry{
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, Expr: `"high"`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySummary, Expr: `'secret_type_display_name' in payload && payload.secret_type_display_name != "" ? payload.secret_type_display_name : ('secret_type' in payload ? payload.secret_type : "")`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDescription, Expr: `'resolution' in payload ? payload.resolution : ""`},
	},
	)
)

// mapExprRepositoryAsset is the CEL mapping expression for GitHub repository payloads mapped to Asset
var mapExprRepositoryAsset = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingAssetSourceIdentifier, Expr: "payload.NameWithOwner"},
	{Key: integrationgenerated.IntegrationMappingAssetDisplayName, Expr: "payload.NameWithOwner"},
	{Key: integrationgenerated.IntegrationMappingAssetName, Expr: "payload.NameWithOwner"},
	{Key: integrationgenerated.IntegrationMappingAssetAssetType, Expr: `"REPOSITORY"`},
	{Key: integrationgenerated.IntegrationMappingAssetSourceType, Expr: `"github"`},
	{Key: integrationgenerated.IntegrationMappingAssetWebsite, Expr: "payload.URL"},
	{Key: integrationgenerated.IntegrationMappingAssetObservedAt, Expr: "payload.UpdatedAt"},
	{Key: integrationgenerated.IntegrationMappingAssetCategories, Expr: `payload.IsPrivate ? ["private"] : ["public"]`},
})

// mapExprDirectoryAccount is the CEL mapping expression for GitHub organization member payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountExternalID, Expr: `payload.DatabaseID != 0 ? string(payload.DatabaseID) : payload.Login`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail, Expr: `payload.Email`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDisplayName, Expr: `payload.Name != "" ? payload.Name : payload.Login`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountAvatarRemoteURL, Expr: `payload.AvatarURL`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDirectoryName, Expr: `payload.Org`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountStatus, Expr: `dyn("ACTIVE")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountProfile, Expr: "payload"},
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
			Schema:  integrationgenerated.IntegrationMappingSchemaVulnerability,
			Variant: variant,
			Spec:    override,
		}
	})

	return append(vulnMappings,
		types.MappingRegistration{
			Schema:  integrationgenerated.IntegrationMappingSchemaAsset,
			Variant: repositoryAssetVariant,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprRepositoryAsset,
			},
		},
		types.MappingRegistration{
			Schema: integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryAccount,
			},
		},
	)
}
