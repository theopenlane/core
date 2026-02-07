package gcpscc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSecurityCenterFindingsConfigSchema verifies the Security Center schema exposes expected fields
func TestSecurityCenterFindingsConfigSchema(t *testing.T) {
	schema := securityCenterFindingsConfigSchema
	require.NotNil(t, schema)
	require.Equal(t, "object", schema["type"])

	props := schemaProperties(t, schema)
	for _, key := range []string{
		"sourceId",
		"filter",
		"page_size",
		"max_findings",
		"include_payloads",
	} {
		require.Contains(t, props, key)
	}
}

// schemaProperties extracts schema properties and fails the test if missing
func schemaProperties(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()

	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok, "expected properties to be a map")
	return props
}
