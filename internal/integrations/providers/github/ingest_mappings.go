package github

import (
	"strconv"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

type celMapEntry struct {
	key  string
	expr string
}

// celMapExpr renders CEL map entries into a CEL object literal string.
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

// githubBaseEntries returns CEL map entries common to all GitHub alert types.
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

// buildGitHubMappingExpr constructs a CEL mapping expression with base fields and type-specific entries.
func buildGitHubMappingExpr(category, externalIDExpr string, extras []celMapEntry) string {
	entries := githubBaseEntries(category, externalIDExpr)
	entries = append(entries, extras...)

	return celMapExpr(entries)
}

var (
	mapExprGitHubDependabot = buildGitHubMappingExpr(
		githubAlertTypeDependabot,
		`"github:" + resource + ":dependabot:" + (payload.number != 0 ? string(payload.number) : (payload.security_advisory.ghsa_id != "" ? payload.security_advisory.ghsa_id : "unknown"))`,
		[]celMapEntry{
			{key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, expr: "payload.security_advisory.severity"},
			{key: integrationgenerated.IntegrationMappingVulnerabilitySummary, expr: "payload.security_advisory.summary"},
			{key: integrationgenerated.IntegrationMappingVulnerabilityDescription, expr: "payload.security_advisory.description"},
			{key: integrationgenerated.IntegrationMappingVulnerabilityCveID, expr: "payload.security_advisory.cve_id"},
		},
	)

	mapExprGitHubCodeScanning = buildGitHubMappingExpr(
		githubAlertTypeCodeScanning,
		`"github:" + resource + ":code_scanning:" + (payload.number != 0 ? string(payload.number) : (payload.id != 0 ? string(payload.id) : "unknown"))`,
		[]celMapEntry{
			{key: integrationgenerated.IntegrationMappingVulnerabilitySeverity, expr: `payload.rule.security_severity_level != "" ? payload.rule.security_severity_level : payload.rule.severity`},
			{key: integrationgenerated.IntegrationMappingVulnerabilitySummary, expr: `payload.rule.description != "" ? payload.rule.description : payload.rule.name`},
		},
	)

	mapExprGitHubSecretScanning = buildGitHubMappingExpr(
		githubAlertTypeSecretScanning,
		`"github:" + resource + ":secret_scanning:" + (payload.number != 0 ? string(payload.number) : "unknown")`,
		[]celMapEntry{
			{key: integrationgenerated.IntegrationMappingVulnerabilitySummary, expr: `payload.secret_type_display_name != "" ? payload.secret_type_display_name : payload.secret_type`},
		},
	)
)

// githubVulnerabilityMappings returns the built-in vulnerability mapping specs for GitHub providers.
func githubVulnerabilityMappings() map[string]integrationtypes.MappingSpec {
	return map[string]integrationtypes.MappingSpec{
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
}

// githubDefaultMappings returns all built-in ingest mappings published by GitHub providers.
func githubDefaultMappings() []integrationtypes.MappingRegistration {
	return lo.MapToSlice(githubVulnerabilityMappings(), func(variant string, spec integrationtypes.MappingSpec) integrationtypes.MappingRegistration {
		return integrationtypes.MappingRegistration{
			Schema:  integrationtypes.MappingSchemaVulnerability,
			Variant: variant,
			Spec:    spec,
		}
	})
}
