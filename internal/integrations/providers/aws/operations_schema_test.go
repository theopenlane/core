package aws

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

// TestSecurityHubFindingsConfigSchema verifies the Security Hub schema validates expected payloads.
func TestSecurityHubFindingsConfigSchema(t *testing.T) {
	schema := securityHubFindingsSchema
	require.NotNil(t, schema)

	validResult, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema),
		gojsonschema.NewBytesLoader([]byte(`{
			"page_size": 25,
			"max_findings": 100,
			"severity": "high",
			"record_state": "ACTIVE",
			"workflow_status": "NEW",
			"include_payloads": true
		}`)),
	)
	require.NoError(t, err)
	require.True(t, validResult.Valid(), "expected valid config, got errors: %v", validResult.Errors())

	invalidResult, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema),
		gojsonschema.NewBytesLoader([]byte(`{"page_size":"not-an-int"}`)),
	)
	require.NoError(t, err)
	require.False(t, invalidResult.Valid(), "expected invalid config to fail schema validation")
}
