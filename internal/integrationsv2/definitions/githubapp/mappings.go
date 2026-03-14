package githubapp

import (
	"strconv"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrationsv2/types"
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

// githubBaseEntries returns CEL map entries common to all GitHub alert types
func githubBaseEntries(category, externalIDExpr string) []celMapEntry {
	return []celMapEntry{
		{key: integrationgenerated.IntegrationMappingVulnerabilityExternalID, expr: externalIDExpr},
		{key: integrationgenerated.IntegrationMappingVulnerabilityExternalOwnerID, expr: "resource"},
		{key: integrationgenerated.IntegrationMappingVulnerabilitySource, expr: `"github"`},
		{key: integrationgenerated.IntegrationMappingVulnerabilityCategory, expr: strconv.Quote(category)},
		{key: integrationgenerated.IntegrationMappingVulnerabilityStatus, expr: "payload.state"},
		{key: integrationgenerated.IntegrationMappingVulnerabilityExternalURI, expr: "payload.html_url"},
		{key: integrationgenerated.IntegrationMappingVulnerabilitySourceUpdatedAt, expr: "payload.updated_at"},
		{key: integrationgenerated.IntegrationMappingVulnerabilityDiscoveredAt, expr: "payload.created_at"},
		{key: integrationgenerated.IntegrationMappingVulnerabilityOpen, expr: `payload.state == "open"`},
		{key: integrationgenerated.IntegrationMappingVulnerabilityRawPayload, expr: "payload"},
	}
}

// buildMappingExpr constructs a CEL mapping expression with base fields and type-specific entries
func buildMappingExpr(category, externalIDExpr string, extras []celMapEntry) string {
	entries := githubBaseEntries(category, externalIDExpr)
	entries = append(entries, extras...)

	return celMapExpr(entries)
}

var (
	mapExprDependabot = buildMappingExpr(
		githubAlertTypeDependabot,
		`"github:" + resource + ":dependabot:" + (payload.number != 0 ? string(payload.number) : (payload.security_advisory.ghsa_id != "" ? payload.security_advisory.ghsa_id : "unknown"))`,
		[]celMapEntry{
			{key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, expr: "payload.security_advisory.severity"},
			{key: integrationgenerated.IntegrationMappingVulnerabilitySummary, expr: "payload.security_advisory.summary"},
			{key: integrationgenerated.IntegrationMappingVulnerabilityDescription, expr: "payload.security_advisory.description"},
			{key: integrationgenerated.IntegrationMappingVulnerabilityCveID, expr: "payload.security_advisory.cve_id"},
		},
	)

	mapExprCodeScanning = buildMappingExpr(
		githubAlertTypeCodeScanning,
		`"github:" + resource + ":code_scanning:" + (payload.number != 0 ? string(payload.number) : (payload.id != 0 ? string(payload.id) : "unknown"))`,
		[]celMapEntry{
			{key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, expr: `payload.rule.security_severity_level != "" ? payload.rule.security_severity_level : payload.rule.severity`},
			{key: integrationgenerated.IntegrationMappingVulnerabilitySummary, expr: `payload.rule.description != "" ? payload.rule.description : payload.rule.name`},
		},
	)

	mapExprSecretScanning = buildMappingExpr(
		githubAlertTypeSecretScanning,
		`"github:" + resource + ":secret_scanning:" + (payload.number != 0 ? string(payload.number) : "unknown")`,
		[]celMapEntry{
			{key: integrationgenerated.IntegrationMappingVulnerabilitySummary, expr: `payload.secret_type_display_name != "" ? payload.secret_type_display_name : payload.secret_type`},
		},
	)
)

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
		githubAlertTypeSecretScanning: {
			FilterExpr: "true",
			MapExpr:    mapExprSecretScanning,
		},
	}

	return lo.MapToSlice(overrides, func(variant string, override types.MappingOverride) types.MappingRegistration {
		return types.MappingRegistration{
			Schema:  integrationgenerated.IntegrationMappingSchemaVulnerability,
			Variant: variant,
			Spec:    override,
		}
	})
}
