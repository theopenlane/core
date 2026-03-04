package ingest

import (
	"testing"

	"github.com/stretchr/testify/require"

	integrationtypes "github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	googleworkspaceprovider "github.com/theopenlane/core/internal/integrations/providers/googleworkspace"
)

// stubMappingIndex is a minimal MappingIndex for use in tests.
type stubMappingIndex struct {
	vulnProviders   map[integrationtypes.ProviderType]map[string]integrationtypes.MappingSpec
	dirAccProviders map[integrationtypes.ProviderType]map[string]integrationtypes.MappingSpec
}

func (s *stubMappingIndex) SupportsIngest(provider integrationtypes.ProviderType, schema integrationtypes.MappingSchema) bool {
	switch schema {
	case integrationtypes.MappingSchemaVulnerability:
		return len(s.vulnProviders[provider]) > 0
	case integrationtypes.MappingSchemaDirectoryAccount:
		return len(s.dirAccProviders[provider]) > 0
	default:
		return false
	}
}

func (s *stubMappingIndex) DefaultMapping(provider integrationtypes.ProviderType, schema integrationtypes.MappingSchema, variant string) (integrationtypes.MappingSpec, bool) {
	var mappings map[string]integrationtypes.MappingSpec
	switch schema {
	case integrationtypes.MappingSchemaVulnerability:
		mappings = s.vulnProviders[provider]
	case integrationtypes.MappingSchemaDirectoryAccount:
		mappings = s.dirAccProviders[provider]
	default:
		return integrationtypes.MappingSpec{}, false
	}

	if len(mappings) == 0 {
		return integrationtypes.MappingSpec{}, false
	}

	if variant != "" {
		if spec, ok := mappings[variant]; ok {
			return spec, true
		}
	}

	spec, ok := mappings[""]

	return spec, ok
}

// TestSupportsDirectoryAccountIngestGoogleWorkspace verifies Google Workspace default mappings are enabled
func TestSupportsDirectoryAccountIngestGoogleWorkspace(t *testing.T) {
	index := &stubMappingIndex{
		dirAccProviders: map[integrationtypes.ProviderType]map[string]integrationtypes.MappingSpec{
			googleworkspaceprovider.TypeGoogleWorkspace: {
				"": {FilterExpr: "true", MapExpr: "{}"},
			},
		},
	}
	SetMappingIndex(index)
	t.Cleanup(func() { SetMappingIndex(nil) })

	require.True(t, SupportsDirectoryAccountIngest(googleworkspaceprovider.TypeGoogleWorkspace, openapi.IntegrationConfig{}))
}

// TestSupportsDirectoryAccountIngestOverride verifies directory account overrides enable ingest for custom providers
func TestSupportsDirectoryAccountIngestOverride(t *testing.T) {
	SetMappingIndex(nil)
	t.Cleanup(func() { SetMappingIndex(nil) })

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
