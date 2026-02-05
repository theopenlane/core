package awssecurityhub

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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

func schemaProperties(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()

	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok, "expected properties to be a map")
	return props
}
