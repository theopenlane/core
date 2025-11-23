package fgamodel

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRelationsForService(t *testing.T) {
	rels, err := RelationsForService()
	require.NoError(t, err)
	require.NotEmpty(t, rels)

	// spot check
	require.Contains(t, rels, "can_view_control")
	require.Contains(t, rels, "can_edit_control")
	require.Contains(t, rels, "can_view_evidence")
	require.Contains(t, rels, "can_edit_evidence")
	require.Contains(t, rels, "can_view_api_token")
	require.Contains(t, rels, "can_edit_api_token")
	require.Contains(t, rels, "can_delete_api_token")
}

func TestNormalizeScope(t *testing.T) {
	require.Equal(t, "can_view", NormalizeScope("read"))
	require.Equal(t, "can_edit", NormalizeScope("write"))
	require.Equal(t, "can_delete", NormalizeScope("delete"))
	require.Equal(t, "can_edit_control", NormalizeScope("write:control"))
	require.Equal(t, "can_view_evidence", NormalizeScope("read:evidence"))
	require.Equal(t, "can_view_api_token", NormalizeScope("read:api_token"))
	require.Equal(t, "can_edit_api_token", NormalizeScope("write:api_token"))
}

func TestScopeOptions(t *testing.T) {
	opts, err := ScopeOptions()
	require.NoError(t, err)
	require.NotEmpty(t, opts)

	require.Contains(t, opts, "organization")
	require.Contains(t, opts["organization"], "read")
	require.Contains(t, opts["organization"], "write")

	require.Contains(t, opts, "control")
	require.Contains(t, opts["control"], "read")
	require.Contains(t, opts["control"], "write")
}
