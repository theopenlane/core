package github

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/providers/schematest"
)

// TestGitHubRepoConfigSchema verifies the repository config schema has expected fields
func TestGitHubRepoConfigSchema(t *testing.T) {
	schema := githubRepoConfigSchema
	require.NotNil(t, schema)

	props := schematest.Properties(t, schema)
	require.Contains(t, props, "visibility")
	require.Contains(t, props, "per_page")
}

// TestGitHubVulnerabilityConfigSchema verifies the vulnerability schema has expected fields
func TestGitHubVulnerabilityConfigSchema(t *testing.T) {
	schema := githubVulnerabilityConfigSchema
	require.NotNil(t, schema)

	props := schematest.Properties(t, schema)
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
