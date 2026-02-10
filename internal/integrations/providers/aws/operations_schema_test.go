package aws

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/providers/schematest"
)

// TestSecurityHubFindingsConfigSchema verifies the Security Hub schema exposes expected fields
func TestSecurityHubFindingsConfigSchema(t *testing.T) {
	schema := securityHubFindingsSchema
	require.NotNil(t, schema)

	props := schematest.Properties(t, schema)
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
