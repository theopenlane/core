package gcpscc

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/providers/schematest"
)

// TestSecurityCenterFindingsConfigSchema verifies the Security Center schema exposes expected fields
func TestSecurityCenterFindingsConfigSchema(t *testing.T) {
	schema := securityCenterFindingsConfigSchema
	require.NotNil(t, schema)

	props := schematest.Properties(t, schema)
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
