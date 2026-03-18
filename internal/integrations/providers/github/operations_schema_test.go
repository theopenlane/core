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

// TestGitHubOrgRepoConfigSchema verifies the organization GraphQL repo config schema has expected fields.
func TestGitHubOrgRepoConfigSchema(t *testing.T) {
	schema := githubOrgRepoConfigSchema
	require.NotNil(t, schema)

	props := schematest.Properties(t, schema)
	require.Contains(t, props, "organization")
	require.Contains(t, props, "per_page")
	require.Contains(t, props, "page_size")
	require.Contains(t, props, "include_payloads")
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
