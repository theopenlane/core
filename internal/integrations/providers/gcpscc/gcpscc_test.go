package gcpscc

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/integrations/types"
)

func TestResolveSecurityCenterParents(t *testing.T) {
	t.Run("specific project scope returns project parents", func(t *testing.T) {
		parents, err := resolveSecurityCenterParents(credentialMetadata{
			ProjectScope: projectScopeSpecific,
			ProjectIDs:   []string{"proj-a", "proj-b"},
		})
		require.NoError(t, err)
		require.Equal(t, []string{"projects/proj-a", "projects/proj-b"}, parents)
	})

	t.Run("organization scope uses organization parent", func(t *testing.T) {
		parents, err := resolveSecurityCenterParents(credentialMetadata{
			OrganizationID: "123456789",
			ProjectScope:   projectScopeAll,
		})
		require.NoError(t, err)
		require.Equal(t, []string{"organizations/123456789"}, parents)
	})

	t.Run("requires project id when no org or specific scope", func(t *testing.T) {
		_, err := resolveSecurityCenterParents(credentialMetadata{
			ProjectScope: projectScopeAll,
		})
		require.ErrorIs(t, err, ErrProjectIDRequired)
	})
}

func TestResolveSecurityCenterSources(t *testing.T) {
	meta := credentialMetadata{
		ProjectScope: projectScopeSpecific,
		ProjectIDs:   []string{"proj-a", "proj-b"},
		SourceID:     types.TrimmedString("456"),
	}

	t.Run("expands short source id across project parents", func(t *testing.T) {
		sources, err := resolveSecurityCenterSources(meta, securityCenterFindingsConfig{})
		require.NoError(t, err)
		require.Equal(t, []string{
			"projects/proj-a/sources/456",
			"projects/proj-b/sources/456",
		}, sources)
	})

	t.Run("uses config source ids when provided", func(t *testing.T) {
		sources, err := resolveSecurityCenterSources(meta, securityCenterFindingsConfig{
			SourceIDs: []string{"projects/proj-x/sources/999", "123"},
		})
		require.NoError(t, err)
		require.Equal(t, []string{
			"projects/proj-x/sources/999",
			"projects/proj-a/sources/123",
			"projects/proj-b/sources/123",
		}, sources)
	})
}
