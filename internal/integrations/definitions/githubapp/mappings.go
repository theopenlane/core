package githubapp

import (
	"strconv"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// githubBaseEntries returns CEL map entries common to all GitHub alert types
func githubBaseEntries(category, externalIDExpr string) []providerkit.CelMapEntry {
	return []providerkit.CelMapEntry{
		{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalID, Expr: externalIDExpr},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalOwnerID, Expr: "resource"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySource, Expr: `"github"`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityCategory, Expr: strconv.Quote(category)},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityStatus, Expr: "payload.state"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityExternalURI, Expr: "payload.html_url"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt, Expr: "payload.updated_at"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt, Expr: "payload.created_at"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityOpen, Expr: `payload.state == "OPEN"`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityRawPayload, Expr: "payload"},
	}
}

// buildMappingExpr constructs a CEL mapping expression with base fields and type-specific entries
func buildMappingExpr(category, externalIDExpr string, extras []providerkit.CelMapEntry) string {
	entries := githubBaseEntries(category, externalIDExpr)
	entries = append(entries, extras...)

	return providerkit.CelMapExpr(entries)
}

var (
	// mapExprDependabot is the CEL mapping expression for Dependabot alert payloads
	mapExprDependabot = buildMappingExpr(githubAlertTypeDependabot, `"github:" + resource + ":dependabot:" + (payload.number != 0 ? string(payload.number) : (payload.security_advisory.ghsa_id != "" ? payload.security_advisory.ghsa_id : "unknown"))`, []providerkit.CelMapEntry{
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, Expr: "payload.security_advisory.severity"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySummary, Expr: "payload.security_advisory.summary"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDescription, Expr: "payload.security_advisory.description"},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityCveID, Expr: "payload.security_advisory.cve_id"},
	},
	)
	// mapExprCodeScanning is the CEL mapping expression for code scanning alert payloads
	mapExprCodeScanning = buildMappingExpr(githubAlertTypeCodeScanning, `"github:" + resource + ":code_scanning:" + (payload.number != 0 ? string(payload.number) : "unknown")`, []providerkit.CelMapEntry{
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, Expr: `payload.rule.security_severity_level != "" ? payload.rule.security_severity_level : payload.rule.severity`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySummary, Expr: `payload.rule.description != "" ? payload.rule.description : payload.rule.name`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDescription, Expr: `payload.most_recent_instance.message.text`},
	},
	)
	// mapExprSecretScanning is the CEL mapping expression for secret scanning alert payloads
	mapExprSecretScanning = buildMappingExpr(githubAlertTypeSecretScan, `"github:" + resource + ":secret_scanning:" + (payload.number != 0 ? string(payload.number) : "unknown")`, []providerkit.CelMapEntry{
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, Expr: `"high"`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilitySummary, Expr: `payload.secret_type_display_name != "" ? payload.secret_type_display_name : payload.secret_type`},
		{Key: integrationgenerated.IntegrationMappingVulnerabilityDescription, Expr: `payload.resolution`},
	},
	)
)

// mapExprRepositoryAsset is the CEL mapping expression for GitHub repository payloads mapped to Asset
var mapExprRepositoryAsset = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingAssetSourceIdentifier, Expr: "payload.nameWithOwner"},
	{Key: integrationgenerated.IntegrationMappingAssetDisplayName, Expr: "payload.nameWithOwner"},
	{Key: integrationgenerated.IntegrationMappingAssetAssetType, Expr: `"repository"`},
	{Key: integrationgenerated.IntegrationMappingAssetSourceType, Expr: `"github"`},
	{Key: integrationgenerated.IntegrationMappingAssetWebsite, Expr: "payload.url"},
	{Key: integrationgenerated.IntegrationMappingAssetObservedAt, Expr: "payload.updatedAt"},
	{Key: integrationgenerated.IntegrationMappingAssetCategories, Expr: `payload.isPrivate ? ["private"] : ["public"]`},
})

// mapExprDirectoryAccount is the CEL mapping expression for GitHub organization member payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountExternalID, Expr: `'database_id' in payload && payload.database_id != 0 ? string(payload.database_id) : ('login' in payload ? payload.login : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail, Expr: `'email' in payload ? payload.email : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDisplayName, Expr: `'name' in payload && payload.name != "" ? payload.name : ('login' in payload ? payload.login : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountAvatarRemoteURL, Expr: `'avatar_url' in payload ? payload.avatar_url : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDirectoryName, Expr: `'org' in payload ? payload.org : ""`},
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
