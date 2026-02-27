package slateparser_test

import (
	"testing"

	"github.com/theopenlane/core/pkg/slateparser"
	"gotest.tools/v3/assert"
)

func TestContainsCommentsInTextJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    []any
		expected bool
	}{
		{
			name: "returns true when comment metadata present",
			input: []any{
				`{
                    "children": [
                        { "text": "asfsadfsd" }
                    ],
                    "id": "b_bwtnb9l8",
                    "type": "p"
                }`,
				`{
                    "children": [
                        { "text": "" }
                    ],
                    "id": "lqbGHj_l70",
                    "type": "p"
                }`,
				`{
                    "children": [
                        { "text": "for a " },
                        { "comment": true, "comment_MDHGnHfbfTfX-amk1Gugp": true, "text": "comment" }
                    ],
                    "id": "qfPeKFLe13",
                    "type": "p"
                }`,
			},
			expected: true,
		},
		{
			name: "returns false when no comment metadata present",
			input: []any{
				`{
                    "children": [
                        { "text": "lets just have text here and see but I added something else here so thats my fault" }
                    ],
                    "id": "kK9tZ5Tllq",
                    "type": "p"
                }`,
			},
			expected: false,
		},
		{
			name: "returns true for comment metadata with different key",
			input: []any{
				`{
            "children": [
                {
                    "text": "another one with "
                },
                {
                    "comment": true,
                    "comment_ribPYmTf5N1Ckx-Y3b0h0": true,
                    "text": "update"
                }
            ],
            "id": "yfmcr1tmAR",
            "type": "p"
        }`,
				`{
            "children": [
                {
                    "text": ""
                }
            ],
            "id": "JlnobxJxJE",
            "type": "p"
        }`,
			},
			expected: true,
		},
		{
			name:     "return false when empty input",
			input:    []any{},
			expected: false,
		},
		{
			name: "returns false for unexpected types",
			input: []any{
				123,
				[]string{"not", "a", "map"},
				map[string]interface{}{
					"children": "this should be a list",
				},
			},
			expected: false,
		},
		{
			name:     "returns false when nil input",
			input:    nil,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := slateparser.ContainsCommentsInTextJSON(tc.input)
			assert.Equal(t, got, tc.expected)
		})
	}
}

func TestOnlyCommentsAdded(t *testing.T) {
	// Helper to wrap children in Slate element
	makeSlate := func(children ...any) []any {
		return []any{
			map[string]any{
				"type":     "paragraph",
				"children": children,
			},
		}
	}

	t.Run("no changes", func(t *testing.T) {
		oldText := makeSlate(map[string]any{"text": "hello"})
		newText := makeSlate(map[string]any{"text": "hello"})
		assert.Check(t, slateparser.OnlyCommentsAdded(oldText, newText))
	})

	t.Run("comment added", func(t *testing.T) {
		oldText := makeSlate(map[string]any{"text": "hello"})
		newText := makeSlate(map[string]any{"text": "hello", "comment": "my comment"})
		assert.Check(t, slateparser.OnlyCommentsAdded(oldText, newText))
	})

	t.Run("comment changed", func(t *testing.T) {
		oldText := makeSlate(map[string]any{"text": "hello", "comment": "old"})
		newText := makeSlate(map[string]any{"text": "hello", "comment": "new"})
		assert.Check(t, slateparser.OnlyCommentsAdded(oldText, newText))
	})

	t.Run("text changed", func(t *testing.T) {
		oldText := makeSlate(map[string]any{"text": "hello"})
		newText := makeSlate(map[string]any{"text": "world"})
		assert.Check(t, !slateparser.OnlyCommentsAdded(oldText, newText))
	})

	t.Run("comment removed", func(t *testing.T) {
		oldText := makeSlate(map[string]any{"text": "hello", "comment": "gone"})
		newText := makeSlate(map[string]any{"text": "hello"})
		assert.Check(t, slateparser.OnlyCommentsAdded(oldText, newText))
	})

	t.Run("multiple children, only comments added", func(t *testing.T) {
		oldText := makeSlate(
			map[string]any{"text": "a"},
			map[string]any{"text": "b"},
		)
		newText := makeSlate(
			map[string]any{"text": "a", "comment": "c1"},
			map[string]any{"text": "b", "comment": "c2"},
		)
		assert.Check(t, slateparser.OnlyCommentsAdded(oldText, newText))
	})

	t.Run("multiple children, text changed in one", func(t *testing.T) {
		oldText := makeSlate(
			map[string]any{"text": "a"},
			map[string]any{"text": "b"},
		)
		newText := makeSlate(
			map[string]any{"text": "a"},
			map[string]any{"text": "B"},
		)
		assert.Check(t, !slateparser.OnlyCommentsAdded(oldText, newText))
	})

	t.Run("different number of children", func(t *testing.T) {
		oldText := makeSlate(map[string]any{"text": "a"})
		newText := makeSlate(map[string]any{"text": "a"}, map[string]any{"text": "b"})
		assert.Check(t, !slateparser.OnlyCommentsAdded(oldText, newText))
	})

	t.Run("non-map children", func(t *testing.T) {
		oldText := makeSlate("not a map")
		newText := makeSlate("not a map")
		assert.Check(t, !slateparser.OnlyCommentsAdded(oldText, newText))
	})

	t.Run("comment added to one of multiple children", func(t *testing.T) {
		oldText := makeSlate(
			map[string]any{"text": "a"},
			map[string]any{"text": "b"},
		)
		newText := makeSlate(
			map[string]any{"text": "a", "comment": "c"},
			map[string]any{"text": "b"},
		)
		assert.Check(t, slateparser.OnlyCommentsAdded(oldText, newText))
	})

	t.Run("multiple children, extra key added", func(t *testing.T) {
		oldText := makeSlate(
			map[string]any{"text": "a"},
			map[string]any{"text": "b"},
		)
		newText := makeSlate(
			map[string]any{"text": "a", "comment": "c1", "extra": "not allowed"},
			map[string]any{"text": "b", "comment": "c2"},
		)
		assert.Check(t, !slateparser.OnlyCommentsAdded(oldText, newText))
	})
}
