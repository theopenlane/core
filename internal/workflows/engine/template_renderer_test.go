package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/workflows"
)

// newTestCELEvaluator builds a CEL evaluator for template tests
func newTestCELEvaluator(t *testing.T) *CELEvaluator {
	t.Helper()

	cfg := workflows.NewDefaultConfig()
	env, err := workflows.NewCELEnv(cfg, workflows.CELScopeAction)
	require.NoError(t, err)

	return NewCELEvaluator(env, cfg)
}

// TestRenderTemplateTextBareIdentifiers verifies bare identifier substitution
func TestRenderTemplateTextBareIdentifiers(t *testing.T) {
	eval := newTestCELEvaluator(t)

	vars := map[string]any{
		"data": map[string]any{
			"review_url": "https://example.com/review",
		},
		"object_id": "obj-123",
	}

	out, err := renderTemplateText(context.Background(), eval, "Review {{review_url}} for {{object_id}}", vars)
	require.NoError(t, err)
	require.Equal(t, "Review https://example.com/review for obj-123", out)
}

// TestRenderTemplateTextCELExpression verifies CEL expressions in template text
func TestRenderTemplateTextCELExpression(t *testing.T) {
	eval := newTestCELEvaluator(t)

	vars := map[string]any{
		"data": map[string]any{
			"count": 2,
		},
	}

	out, err := renderTemplateText(context.Background(), eval, "Total {{data.count + 1}}", vars)
	require.NoError(t, err)
	require.Equal(t, "Total 3", out)
}

// TestRenderTemplateValueSingleExpressionReturnsTypedValue verifies typed value passthrough
func TestRenderTemplateValueSingleExpressionReturnsTypedValue(t *testing.T) {
	eval := newTestCELEvaluator(t)

	payload := map[string]any{"key": "value"}
	vars := map[string]any{
		"data": map[string]any{
			"payload": payload,
		},
	}

	out, err := renderTemplateValue(context.Background(), eval, "{{data.payload}}", vars)
	require.NoError(t, err)
	require.Equal(t, payload, out)
}

// TestRenderTemplateValueWalksCompositeValues verifies composite rendering behavior
func TestRenderTemplateValueWalksCompositeValues(t *testing.T) {
	eval := newTestCELEvaluator(t)

	vars := map[string]any{
		"data": map[string]any{
			"name":  "Ada",
			"count": 5,
		},
	}

	input := map[string]any{
		"text": "{{name}}",
		"items": []any{
			"{{data.count}}",
			"plain",
		},
	}

	out, err := renderTemplateValue(context.Background(), eval, input, vars)
	require.NoError(t, err)

	result := out.(map[string]any)
	require.Equal(t, "Ada", result["text"])

	items := result["items"].([]any)
	require.EqualValues(t, 5, items[0])
	require.Equal(t, "plain", items[1])
}

// TestRenderTemplateTextInvalidExpressionReturnsError verifies invalid expression errors
func TestRenderTemplateTextInvalidExpressionReturnsError(t *testing.T) {
	eval := newTestCELEvaluator(t)

	_, err := renderTemplateText(context.Background(), eval, "Bad {{data.}}", map[string]any{"data": map[string]any{}})
	require.Error(t, err)
}

// TestRenderTemplateValueDepthLimit verifies recursion depth limits are enforced
func TestRenderTemplateValueDepthLimit(t *testing.T) {
	eval := newTestCELEvaluator(t)
	vars := map[string]any{"data": map[string]any{}}

	// Build deeply nested structure exceeding maxTemplateRenderDepth
	var nested any = "leaf"
	for range maxTemplateRenderDepth + 10 {
		nested = map[string]any{"level": nested}
	}

	_, err := renderTemplateValue(context.Background(), eval, nested, vars)
	require.ErrorIs(t, err, ErrTemplateRenderDepthExceeded)
}
