package googleworkspace

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/jsonx"
)

// TestResolveDirectorySyncConfig verifies that installation defaults and per-run overrides merge correctly
func TestResolveDirectorySyncConfig(t *testing.T) {
	t.Run("applies installation defaults", func(t *testing.T) {
		installation := testIntegrationWithUserInput(t, UserInput{
			CustomerID:         "customer-123",
			OrganizationalUnit: "/Engineering",
			IncludeSuspended:   true,
			EnableGroupSync:    false,
		})

		cfg := resolveDirectorySyncConfig(installation, DirectorySyncConfig{
			Query: "isSuspended=false",
		})

		assert.Equal(t, "customer-123", cfg.Customer)
		assert.Equal(t, "/Engineering", cfg.OrganizationalUnit)
		assert.Equal(t, "isSuspended=false", cfg.Query)
		require.NotNil(t, cfg.IncludeSuspended)
		assert.True(t, *cfg.IncludeSuspended)
		require.NotNil(t, cfg.IncludeGroups)
		assert.False(t, *cfg.IncludeGroups)
	})

	t.Run("per run overrides win", func(t *testing.T) {
		installation := testIntegrationWithUserInput(t, UserInput{
			CustomerID:         "customer-123",
			OrganizationalUnit: "/Engineering",
			IncludeSuspended:   true,
			EnableGroupSync:    false,
		})

		cfg := resolveDirectorySyncConfig(installation, DirectorySyncConfig{
			Customer:           "customer-override",
			OrganizationalUnit: "/Security",
			IncludeSuspended:   lo.ToPtr(false),
			IncludeGroups:      lo.ToPtr(true),
		})

		assert.Equal(t, "customer-override", cfg.Customer)
		assert.Equal(t, "/Security", cfg.OrganizationalUnit)
		require.NotNil(t, cfg.IncludeSuspended)
		assert.False(t, *cfg.IncludeSuspended)
		require.NotNil(t, cfg.IncludeGroups)
		assert.True(t, *cfg.IncludeGroups)
	})
}

// testIntegrationWithUserInput constructs a minimal Integration with the given user input encoded as ClientConfig
func testIntegrationWithUserInput(t *testing.T, input UserInput) *generated.Integration {
	t.Helper()

	raw, err := jsonx.ToRawMessage(input)
	require.NoError(t, err)

	return &generated.Integration{
		Config: openapi.IntegrationConfig{
			ClientConfig: raw,
		},
	}
}
