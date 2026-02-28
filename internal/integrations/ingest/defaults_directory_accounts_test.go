package ingest

import (
	"testing"

	"github.com/stretchr/testify/require"

	openapi "github.com/theopenlane/core/common/openapi"
	googleworkspaceprovider "github.com/theopenlane/core/internal/integrations/providers/googleworkspace"
)

// TestSupportsDirectoryAccountIngestGoogleWorkspace verifies Google Workspace default mappings are enabled
func TestSupportsDirectoryAccountIngestGoogleWorkspace(t *testing.T) {
	require.True(t, SupportsDirectoryAccountIngest(googleworkspaceprovider.TypeGoogleWorkspace, openapi.IntegrationConfig{}))
}

// TestSupportsDirectoryAccountIngestOverride verifies directory account overrides enable ingest for custom providers
func TestSupportsDirectoryAccountIngestOverride(t *testing.T) {
	config := openapi.IntegrationConfig{
		MappingOverrides: map[string]openapi.IntegrationMappingOverride{
			"DirectoryAccount": {
				FilterExpr: "true",
				MapExpr:    `{"externalID":"id","status":"ACTIVE","mfaState":"UNKNOWN","observedAt":"2026-01-01T00:00:00Z","profileHash":"id"}`,
			},
		},
	}

	require.True(t, SupportsDirectoryAccountIngest("custom", config))
}
