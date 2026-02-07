package ingest

import (
	"strconv"
	"strings"

	integrationtypes "github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
)

const (
	githubAlertTypeDependabot     = "dependabot"
	githubAlertTypeCodeScanning   = "code_scanning"
	githubAlertTypeSecretScanning = "secret_scanning"
)

type celMapEntry struct {
	key  string
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

var (
	mapExprGitHubDependabot = celMapExpr([]celMapEntry{
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityExternalID,
			expr: `"github:" + resource + ":dependabot:" + (payload.number != 0 ? string(payload.number) : (payload.security_advisory.ghsa_id != "" ? payload.security_advisory.ghsa_id : "unknown"))`,
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityExternalOwnerID,
			expr: "resource",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilitySource,
			expr: `"github"`,
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityCategory,
			expr: `"dependabot"`,
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilitySeverity,
			expr: "payload.security_advisory.severity",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityStatus,
			expr: "payload.state",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilitySummary,
			expr: "payload.security_advisory.summary",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityDescription,
			expr: "payload.security_advisory.description",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityCveID,
			expr: "payload.security_advisory.cve_id",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityExternalURI,
			expr: "payload.html_url",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt,
			expr: "payload.updated_at",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt,
			expr: "payload.created_at",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityOpen,
			expr: `payload.state == "open"`,
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityRawPayload,
			expr: "payload",
		},
	})

	mapExprGitHubCodeScanning = celMapExpr([]celMapEntry{
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityExternalID,
			expr: `"github:" + resource + ":code_scanning:" + (payload.number != 0 ? string(payload.number) : (payload.id != 0 ? string(payload.id) : "unknown"))`,
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityExternalOwnerID,
			expr: "resource",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilitySource,
			expr: `"github"`,
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityCategory,
			expr: `"code_scanning"`,
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilitySeverity,
			expr: `payload.rule.security_severity_level != "" ? payload.rule.security_severity_level : payload.rule.severity`,
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityStatus,
			expr: "payload.state",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilitySummary,
			expr: `payload.rule.description != "" ? payload.rule.description : payload.rule.name`,
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityExternalURI,
			expr: "payload.html_url",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt,
			expr: "payload.updated_at",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt,
			expr: "payload.created_at",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityOpen,
			expr: `payload.state == "open"`,
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityRawPayload,
			expr: "payload",
		},
	})

	mapExprGitHubSecretScanning = celMapExpr([]celMapEntry{
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityExternalID,
			expr: `"github:" + resource + ":secret_scanning:" + (payload.number != 0 ? string(payload.number) : "unknown")`,
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityExternalOwnerID,
			expr: "resource",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilitySource,
			expr: `"github"`,
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityCategory,
			expr: `"secret_scanning"`,
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityStatus,
			expr: "payload.state",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilitySummary,
			expr: `payload.secret_type_display_name != "" ? payload.secret_type_display_name : payload.secret_type`,
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityExternalURI,
			expr: "payload.html_url",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt,
			expr: "payload.updated_at",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt,
			expr: "payload.created_at",
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityOpen,
			expr: `payload.state == "open"`,
		},
		{
			key:  integrationgenerated.IntegrationMappingVulnerabilityRawPayload,
			expr: "payload",
		},
	})
)

var normalizedVulnerabilitySchema = normalizeMappingKey(mappingSchemaVulnerability)

var githubVulnerabilityMappings = map[string]openapi.IntegrationMappingOverride{
	githubAlertTypeDependabot: {
		FilterExpr: "true",
		MapExpr:    mapExprGitHubDependabot,
	},
	githubAlertTypeCodeScanning: {
		FilterExpr: "true",
		MapExpr:    mapExprGitHubCodeScanning,
	},
	githubAlertTypeSecretScanning: {
		FilterExpr: "true",
		MapExpr:    mapExprGitHubSecretScanning,
	},
}

// defaultMappingSpec returns built-in mappings for supported providers
func defaultMappingSpec(provider integrationtypes.ProviderType, schemaName string, variant string) (openapi.IntegrationMappingOverride, bool) {
	if normalizeMappingKey(schemaName) != normalizedVulnerabilitySchema {
		return openapi.IntegrationMappingOverride{}, false
	}

	switch provider {
	case integrationtypes.ProviderType("github"), integrationtypes.ProviderType("github_app"):
		return githubMappingSpec(variant)
	default:
		return openapi.IntegrationMappingOverride{}, false
	}
}

// supportsDefaultMapping reports whether built-in mappings exist for a schema
func supportsDefaultMapping(provider integrationtypes.ProviderType, schemaName string) bool {
	if normalizeMappingKey(schemaName) != normalizedVulnerabilitySchema {
		return false
	}

	switch provider {
	case integrationtypes.ProviderType("github"), integrationtypes.ProviderType("github_app"):
		return true
	default:
		return false
	}
}

// githubMappingSpec selects the GitHub mapping for a specific alert type
func githubMappingSpec(alertType string) (openapi.IntegrationMappingOverride, bool) {
	alertType = normalizeAlertType(alertType)
	spec, ok := githubVulnerabilityMappings[alertType]
	if !ok {
		return openapi.IntegrationMappingOverride{}, false
	}

	return spec, true
}

// normalizeAlertType standardizes alert type identifiers
func normalizeAlertType(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")
	switch value {
	case "dependabot_alerts":
		return githubAlertTypeDependabot
	case "code_scanning_alerts":
		return githubAlertTypeCodeScanning
	case "secret_scanning_alerts":
		return githubAlertTypeSecretScanning
	}

	return value
}
