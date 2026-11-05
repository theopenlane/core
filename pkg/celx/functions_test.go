package celx

import (
	"context"
	"testing"
	"time"

	"github.com/google/cel-go/cel"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestParagraphs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "already has newlines is unchanged",
			input: "### Impact\n\nSQL Injection can occur when:\n\n1. The non-default simple protocol is used.",
			want:  "### Impact\n\nSQL Injection can occur when:\n\n1. The non-default simple protocol is used.",
		},
		{
			name:  "multiple separators normalized",
			input: "A Git repository can be crafted in a dangerous way.     Update Instructions:     Run `sudo pro fix CVE-2025-27614` to fix the vulnerability.",
			want:  "A Git repository can be crafted in a dangerous way.\n\nUpdate Instructions:\n\nRun `sudo pro fix CVE-2025-27614` to fix the vulnerability.",
		},
		{
			name:  "leading and trailing whitespace trimmed",
			input: "   Some description.     More details.   ",
			want:  "Some description.\n\nMore details.",
		},
		{
			name:  "double space is not a separator",
			input: "Sentence one.  Sentence two.",
			want:  "Sentence one.  Sentence two.",
		},
		{
			name:  "triple space is a separator",
			input: "Section one.   Section two.",
			want:  "Section one.\n\nSection two.",
		},
		{
			name:  "empty string is unchanged",
			input: "",
			want:  "",
		},
		{
			name:  "plain text with no formatting is unchanged",
			input: "Plain description with no formatting.",
			want:  "Plain description with no formatting.",
		},
		{
			name:  "existing newline short-circuits separator normalization",
			input: "First line.\nSecond line.     Not normalized.",
			want:  "First line.\nSecond line.     Not normalized.",
		},
	}

	env, err := NewEnv(EnvConfig{}, cel.Variable("s", cel.StringType))
	assert.Check(t, err)

	eval := NewEvaluator(env, EvalConfig{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, _, err := eval.Evaluate(context.Background(), "paragraphs(s)", map[string]any{"s": tt.input})
			assert.Check(t, err)
			assert.Check(t, is.Equal(tt.want, out.Value()))
		})
	}
}

func TestParagraphsNull(t *testing.T) {
	env, err := NewEnv(EnvConfig{}, cel.Variable("s", cel.DynType))
	assert.Check(t, err)

	eval := NewEvaluator(env, EvalConfig{})

	out, _, err := eval.Evaluate(context.Background(), "paragraphs(s)", map[string]any{"s": nil})
	assert.Check(t, err)
	assert.Check(t, is.Equal("", out.Value()))
}

func TestMerge(t *testing.T) {
	env, err := NewEnv(EnvConfig{}, cel.Variable("payload", cel.DynType))
	assert.NilError(t, err)

	eval := NewEvaluator(env, EvalConfig{Timeout: time.Second})

	out, err := eval.EvaluateJSONMap(context.Background(), `merge(payload, {"status": "DRAFT", "approver_group_name": "Admins"})`, map[string]any{
		"payload": map[string]any{"name": "policy", "status": "PUBLISHED"},
	})
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(map[string]any{"name": "policy", "status": "DRAFT", "approver_group_name": "Admins"}, out))

	out, err = eval.EvaluateJSONMap(context.Background(), `merge(payload, {"status": "DRAFT"})`, map[string]any{"payload": nil})
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(map[string]any{"status": "DRAFT"}, out))
}
