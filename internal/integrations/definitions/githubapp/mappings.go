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
		{Key: integrationgenerated.IntegrationMappingVulnerabilityVulnerabilityStatusName, Expr: "payload.state"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalURI, Expr: "payload.html_url"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt, Expr: "payload.updated_at"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt, Expr: "payload.created_at"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityPublishedAt, Expr: "payload.security_advisory.published_at"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDismissedAt, Expr: "payload.dismissed_at"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDismissedReason, Expr: "payload.dismissed_reason"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDismissedComment, Expr: "payload.dismissed_comment"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityFixedAt, Expr: "payload.fixed_at"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityAutoDismissedAt, Expr: "payload.auto_dismissed_at"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityOpen, Expr: `payload.state == "OPEN"`},
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
		{Key: integrationgenerated.IntegrationMappingVulnerabilityVulnerabilityStatusName, Expr: "payload.State"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalURI, Expr: "payload.URL"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt, Expr: "payload.UpdatedAt"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt, Expr: "payload.CreatedAt"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityPublishedAt, Expr: "payload.SecurityVulnerability.Advisory.PublishedAt"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDismissedAt, Expr: "payload.DismissedAt"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDismissedReason, Expr: "payload.DismissReason"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDismissedComment, Expr: "payload.DismissComment"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityFixedAt, Expr: "payload.FixedAt"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityAutoDismissedAt, Expr: "payload.AutoDismissedAt"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityOpen, Expr: `payload.State == "OPEN"`},
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
	mapExprDependabot = buildWebhookMappingExpr(githubAlertTypeDependabot, `"github:" + resource + ":dependabot:" + (payload.number != 0 ? string(payload.number) : (payload.security_advisory.ghsa_id != "" ? payload.security_advisory.ghsa_id : "unknown"))`, []providerkit.CelMapEntry{
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, Expr: "payload.security_advisory.severity"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySummary, Expr: "payload.security_advisory.summary"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDescription, Expr: "payload.security_advisory.description"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityCveID, Expr: "payload.security_advisory.cve_id"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityScore, Expr: "payload.security_advisory.cvss.score"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityVector, Expr: "payload.security_advisory.cvss.vector_string"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityCweIds, Expr: `payload.security_advisory.cwes.map(c, c.cwe_id)`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityReferences, Expr: `payload.security_advisory.references.map(r, r.url)`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityVulnerableVersionRange, Expr: "payload.security_vulnerability.vulnerable_version_range"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityFirstPatchedVersion, Expr: "payload.security_vulnerability.first_patched_version.identifier"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityPackageName, Expr: "payload.security_vulnerability.package.name"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityPackageEcosystem, Expr: "payload.security_vulnerability.package.ecosystem"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityManifestPath, Expr: "payload.dependency.manifest_path"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDependencyScope, Expr: "payload.dependency.scope"},
	},
	)
	// mapExprDependabotPoll is the CEL mapping expression for Dependabot alerts collected via GraphQL poll
	mapExprDependabotPoll = buildPollMappingExpr(githubAlertTypeDependabot, `"github:" + resource + ":dependabot:" + (payload.Number != 0 ? string(payload.Number) : (payload.SecurityVulnerability.Advisory.GHSAID != "" ? payload.SecurityVulnerability.Advisory.GHSAID : "unknown"))`, []providerkit.CelMapEntry{
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, Expr: "payload.SecurityVulnerability.Severity"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySummary, Expr: "payload.SecurityVulnerability.Advisory.Summary"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDescription, Expr: "payload.SecurityVulnerability.Advisory.Description"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityCveID, Expr: `payload.SecurityVulnerability.Advisory.Identifiers.filter(i, i.Type == "CVE").size() > 0 ? payload.SecurityVulnerability.Advisory.Identifiers.filter(i, i.Type == "CVE")[0].Value : ""`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityScore, Expr: "payload.SecurityVulnerability.Advisory.Cvss.Score"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityVector, Expr: "payload.SecurityVulnerability.Advisory.Cvss.VectorString"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityCweIds, Expr: `payload.SecurityVulnerability.Advisory.Cwes.Nodes.map(c, c.CweID)`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityReferences, Expr: `payload.SecurityVulnerability.Advisory.References.map(r, r.URL)`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityVulnerableVersionRange, Expr: "payload.SecurityVulnerability.VulnerableVersionRange"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityFirstPatchedVersion, Expr: "payload.SecurityVulnerability.FirstPatchedVersion.Identifier"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityPackageName, Expr: "payload.SecurityVulnerability.Package.Name"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityPackageEcosystem, Expr: "payload.SecurityVulnerability.Package.Ecosystem"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityManifestPath, Expr: "payload.VulnerableManifestPath"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDependencyScope, Expr: "payload.DependencyScope"},
	},
	)
	// mapExprCodeScanning is the CEL mapping expression for code scanning webhook alert payloads
	mapExprCodeScanning = buildWebhookMappingExpr(githubAlertTypeCodeScanning, `"github:" + resource + ":code_scanning:" + (payload.number != 0 ? string(payload.number) : "unknown")`, []providerkit.CelMapEntry{
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, Expr: `payload.rule.security_severity_level != "" ? payload.rule.security_severity_level : payload.rule.severity`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySummary, Expr: `payload.rule.description != "" ? payload.rule.description : payload.rule.name`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDescription, Expr: `payload.most_recent_instance.message.text`},
	},
	)
	// mapExprSecretScanning is the CEL mapping expression for secret scanning webhook alert payloads
	mapExprSecretScanning = buildWebhookMappingExpr(githubAlertTypeSecretScan, `"github:" + resource + ":secret_scanning:" + (payload.number != 0 ? string(payload.number) : "unknown")`, []providerkit.CelMapEntry{
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, Expr: `"high"`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySummary, Expr: `payload.secret_type_display_name != "" ? payload.secret_type_display_name : payload.secret_type`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDescription, Expr: `payload.resolution`},
	},
	)
)

// mapExprRepositoryAsset is the CEL mapping expression for GitHub repository payloads mapped to Asset
var mapExprRepositoryAsset = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingAssetSourceIdentifier, Expr: "payload.NameWithOwner"},
	{Key: integrationgenerated.IntegrationMappingAssetDisplayName, Expr: "payload.NameWithOwner"},
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
