package gcpscc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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

func schemaProperties(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()

	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok, "expected properties to be a map")
	return props
}
