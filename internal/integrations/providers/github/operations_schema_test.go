package github

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

// TestGitHubRepoConfigSchema verifies the repository config schema validates expected payloads.
func TestGitHubRepoConfigSchema(t *testing.T) {
	schema := githubRepoConfigSchema
	require.NotNil(t, schema)

	validResult, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema),
		gojsonschema.NewBytesLoader([]byte(`{"visibility":"private","per_page":50}`)),
	)
	require.NoError(t, err)
	require.True(t, validResult.Valid(), "expected valid config, got errors: %v", validResult.Errors())

	invalidResult, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema),
		gojsonschema.NewBytesLoader([]byte(`{"per_page":"not-an-int"}`)),
	)
	require.NoError(t, err)
	require.False(t, invalidResult.Valid(), "expected invalid config to fail schema validation")
}

// TestGitHubOrgRepoConfigSchema verifies the organization GraphQL repo config schema has expected fields.
func TestGitHubOrgRepoConfigSchema(t *testing.T) {
	schema := githubOrgRepoConfigSchema
	require.NotNil(t, schema)

	validResult, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema),
		gojsonschema.NewBytesLoader([]byte(`{
			"organization": "acme-org",
			"per_page": 100,
			"include_payloads": true
		}`)),
	)
	require.NoError(t, err)
	require.True(t, validResult.Valid(), "expected valid config, got errors: %v", validResult.Errors())

	invalidResult, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema),
		gojsonschema.NewBytesLoader([]byte(`{"organization":101}`)),
	)
	require.NoError(t, err)
	require.False(t, invalidResult.Valid(), "expected invalid config to fail schema validation")
}

// TestGitHubVulnerabilityConfigSchema verifies the vulnerability schema validates expected payloads.
func TestGitHubVulnerabilityConfigSchema(t *testing.T) {
	schema := githubVulnerabilityConfigSchema
	require.NotNil(t, schema)

	validResult, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema),
		gojsonschema.NewBytesLoader([]byte(`{
			"alert_types": ["dependabot", "code_scanning"],
			"repositories": ["acme/repo-1", "acme/repo-2"],
			"visibility": "all",
			"affiliation": "owner",
			"per_page": 50,
			"max_repos": 200,
			"include_payloads": true,
			"alert_state": "open",
			"severity": "critical",
			"ecosystem": "npm"
		}`)),
	)
	require.NoError(t, err)
	require.True(t, validResult.Valid(), "expected valid config, got errors: %v", validResult.Errors())

	invalidResult, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema),
		gojsonschema.NewBytesLoader([]byte(`{"alert_types":"dependabot"}`)),
	)
	require.NoError(t, err)
	require.False(t, invalidResult.Valid(), "expected invalid config to fail schema validation")
}
