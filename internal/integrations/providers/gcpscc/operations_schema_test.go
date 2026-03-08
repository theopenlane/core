package gcpscc

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

// TestSecurityCenterFindingsConfigSchema verifies the Security Center schema validates expected payloads.
func TestSecurityCenterFindingsConfigSchema(t *testing.T) {
	schema := securityCenterFindingsConfigSchema
	require.NotNil(t, schema)

	validResult, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema),
		gojsonschema.NewBytesLoader([]byte(`{
			"sourceId": "organizations/123/sources/456",
			"sourceIds": ["456", "789"],
			"filter": "state=\"ACTIVE\"",
			"page_size": 200,
			"max_findings": 500,
			"include_payloads": true
		}`)),
	)
	require.NoError(t, err)
	require.True(t, validResult.Valid(), "expected valid config, got errors: %v", validResult.Errors())

	invalidResult, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema),
		gojsonschema.NewBytesLoader([]byte(`{"sourceIds":"not-an-array"}`)),
	)
	require.NoError(t, err)
	require.False(t, invalidResult.Valid(), "expected invalid config to fail schema validation")
}
