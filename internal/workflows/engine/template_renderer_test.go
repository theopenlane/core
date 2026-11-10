package engine

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestRenderTemplateText_Substitution(t *testing.T) {
	vars := map[string]any{
		"review_url": "https://example.com/review",
		"object_id":  "obj-123",
	}

	out, err := renderTemplateText("Review {{.review_url}} for {{.object_id}}", vars)
	assert.NilError(t, err)
	assert.Assert(t, is.Equal("Review https://example.com/review for obj-123", out))
}

func TestRenderTemplateText_NestedAccess(t *testing.T) {
	vars := map[string]any{
		"data": map[string]any{"name": "Ada"},
	}

	out, err := renderTemplateText("Hello {{.data.name}}", vars)
	assert.NilError(t, err)
	assert.Assert(t, is.Equal("Hello Ada", out))
}

func TestRenderTemplateText_NoExpressions(t *testing.T) {
	out, err := renderTemplateText("plain string", map[string]any{})
	assert.NilError(t, err)
	assert.Assert(t, is.Equal("plain string", out))
}

func TestRenderTemplateText_Empty(t *testing.T) {
	out, err := renderTemplateText("", map[string]any{})
	assert.NilError(t, err)
	assert.Assert(t, is.Equal("", out))
}

func TestRenderTemplateText_MissingKey(t *testing.T) {
	out, err := renderTemplateText("Hello {{.missing}}", map[string]any{})
	assert.NilError(t, err)
	assert.Assert(t, is.Equal("Hello ", out))
}
