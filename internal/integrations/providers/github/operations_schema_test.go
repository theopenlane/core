package github

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGitHubRepoConfigSchema(t *testing.T) {
	schema := githubRepoConfigSchema
	require.NotNil(t, schema)
	require.Equal(t, "object", schema["type"])

	props := schemaProperties(t, schema)
	require.Contains(t, props, "visibility")
	require.Contains(t, props, "per_page")
}

func TestGitHubVulnerabilityConfigSchema(t *testing.T) {
	schema := githubVulnerabilityConfigSchema
	require.NotNil(t, schema)
	require.Equal(t, "object", schema["type"])

	props := schemaProperties(t, schema)
	for _, key := range []string{
		"alert_types",
		"repositories",
		"visibility",
		"affiliation",
		"per_page",
		"max_repos",
		"include_payloads",
		"alert_state",
		"severity",
		"ecosystem",
	} {
		require.Contains(t, props, key)
	}
}

func schemaProperties(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()

	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok, "expected properties to be a map")
	return props
}
