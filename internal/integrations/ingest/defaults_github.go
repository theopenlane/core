package ingest

import (
	"strings"

	integrationtypes "github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
)

const (
	githubAlertTypeDependabot     = "dependabot"
	githubAlertTypeCodeScanning   = "code_scanning"
	githubAlertTypeSecretScanning = "secret_scanning"
)

const (
	mapExprGitHubDependabot = `{
  "externalID": "github:" + resource + ":dependabot:" + (payload.number != 0 ? string(payload.number) : (payload.security_advisory.ghsa_id != "" ? payload.security_advisory.ghsa_id : "unknown")),
  "externalOwnerID": resource,
  "source": "github",
  "category": "dependabot",
  "severity": payload.security_advisory.severity,
  "status": payload.state,
  "summary": payload.security_advisory.summary,
  "description": payload.security_advisory.description,
  "cveID": payload.security_advisory.cve_id,
  "externalURI": payload.html_url,
  "sourceUpdatedAt": payload.updated_at,
  "discoveredAt": payload.created_at,
  "open": payload.state == "open",
  "rawPayload": payload
}`

	mapExprGitHubCodeScanning = `{
  "externalID": "github:" + resource + ":code_scanning:" + (payload.number != 0 ? string(payload.number) : (payload.id != 0 ? string(payload.id) : "unknown")),
  "externalOwnerID": resource,
  "source": "github",
  "category": "code_scanning",
  "severity": payload.rule.security_severity_level != "" ? payload.rule.security_severity_level : payload.rule.severity,
  "status": payload.state,
  "summary": payload.rule.description != "" ? payload.rule.description : payload.rule.name,
  "externalURI": payload.html_url,
  "sourceUpdatedAt": payload.updated_at,
  "discoveredAt": payload.created_at,
  "open": payload.state == "open",
  "rawPayload": payload
}`

	mapExprGitHubSecretScanning = `{
  "externalID": "github:" + resource + ":secret_scanning:" + (payload.number != 0 ? string(payload.number) : "unknown"),
  "externalOwnerID": resource,
  "source": "github",
  "category": "secret_scanning",
  "status": payload.state,
  "summary": payload.secret_type_display_name != "" ? payload.secret_type_display_name : payload.secret_type,
  "externalURI": payload.html_url,
  "sourceUpdatedAt": payload.updated_at,
  "discoveredAt": payload.created_at,
  "open": payload.state == "open",
  "rawPayload": payload
}`
)

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
	if normalizeMappingKey(schemaName) != normalizeMappingKey(mappingSchemaVulnerability) {
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
	if normalizeMappingKey(schemaName) != normalizeMappingKey(mappingSchemaVulnerability) {
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
