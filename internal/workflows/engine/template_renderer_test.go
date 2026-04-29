package engine

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderTemplateText_Substitution(t *testing.T) {
	vars := map[string]any{
		"review_url": "https://example.com/review",
		"object_id":  "obj-123",
	}

	out, err := renderTemplateText("Review {{.review_url}} for {{.object_id}}", vars)
	require.NoError(t, err)
	require.Equal(t, "Review https://example.com/review for obj-123", out)
}

func TestRenderTemplateText_NestedAccess(t *testing.T) {
	vars := map[string]any{
		"data": map[string]any{"name": "Ada"},
	}

	out, err := renderTemplateText("Hello {{.data.name}}", vars)
	require.NoError(t, err)
	require.Equal(t, "Hello Ada", out)
}

func TestRenderTemplateText_NoExpressions(t *testing.T) {
	out, err := renderTemplateText("plain string", map[string]any{})
	require.NoError(t, err)
	require.Equal(t, "plain string", out)
}

func TestRenderTemplateText_Empty(t *testing.T) {
	out, err := renderTemplateText("", map[string]any{})
	require.NoError(t, err)
	require.Equal(t, "", out)
}

func TestRenderTemplateText_MissingKey(t *testing.T) {
	out, err := renderTemplateText("Hello {{.missing}}", map[string]any{})
	require.NoError(t, err)
	require.Equal(t, "Hello ", out)
}
