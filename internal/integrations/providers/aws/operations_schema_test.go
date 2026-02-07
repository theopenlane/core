package aws

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSecurityHubFindingsConfigSchema verifies the Security Hub schema exposes expected fields
func TestSecurityHubFindingsConfigSchema(t *testing.T) {
	schema := securityHubFindingsSchema
	require.NotNil(t, schema)
	require.Equal(t, "object", schema["type"])

	props := schemaProperties(t, schema)
	for _, key := range []string{
		"page_size",
		"max_findings",
		"severity",
		"record_state",
		"workflow_status",
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
