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

	t.Run("child with non-comparable value does not panic", func(t *testing.T) {
		// []any values can't be safely compared; best-effort returns false rather than panicking
		makeSlateWith := func(child map[string]any) []any {
			return []any{map[string]any{"type": "paragraph", "children": []any{child}}}
		}
		marks := []any{"bold"}
		oldText := makeSlateWith(map[string]any{"text": "hello", "marks": marks})
		newText := makeSlateWith(map[string]any{"text": "hello", "marks": marks})
		assert.Check(t, !slateparser.OnlyCommentsAdded(oldText, newText))
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

	// bold/italic/underline are stored as booleans on leaf nodes in Slate
	t.Run("bold mark unchanged, comment added", func(t *testing.T) {
		oldText := makeSlate(map[string]any{"text": "hello", "bold": true})
		newText := makeSlate(map[string]any{"text": "hello", "bold": true, "comment": true})
		assert.Check(t, slateparser.OnlyCommentsAdded(oldText, newText))
	})

	t.Run("bold mark changed", func(t *testing.T) {
		oldText := makeSlate(map[string]any{"text": "hello", "bold": true})
		newText := makeSlate(map[string]any{"text": "hello", "bold": false})
		assert.Check(t, !slateparser.OnlyCommentsAdded(oldText, newText))
	})

	t.Run("bold mark added (formatting change, not comment)", func(t *testing.T) {
		oldText := makeSlate(map[string]any{"text": "hello"})
		newText := makeSlate(map[string]any{"text": "hello", "bold": true})
		assert.Check(t, !slateparser.OnlyCommentsAdded(oldText, newText))
	})

	t.Run("multiple marks unchanged, comment added", func(t *testing.T) {
		oldText := makeSlate(map[string]any{"text": "hello", "bold": true, "italic": true, "underline": true})
		newText := makeSlate(map[string]any{"text": "hello", "bold": true, "italic": true, "underline": true, "comment": true})
		assert.Check(t, slateparser.OnlyCommentsAdded(oldText, newText))
	})

	// Slate adds both "comment" and "comment_<id>" keys when creating a comment
	t.Run("comment and comment_id keys added", func(t *testing.T) {
		oldText := makeSlate(map[string]any{"text": "hello"})
		newText := makeSlate(map[string]any{"text": "hello", "comment": true, "comment_MDHGnHfbfTfX": true})
		assert.Check(t, slateparser.OnlyCommentsAdded(oldText, newText))
	})

	t.Run("multiple comment_id keys added", func(t *testing.T) {
		oldText := makeSlate(map[string]any{"text": "hello"})
		newText := makeSlate(map[string]any{"text": "hello", "comment": true, "comment_abc": true, "comment_xyz": true})
		assert.Check(t, slateparser.OnlyCommentsAdded(oldText, newText))
	})

	t.Run("bold with comment_id added", func(t *testing.T) {
		oldText := makeSlate(map[string]any{"text": "hello", "bold": true})
		newText := makeSlate(map[string]any{"text": "hello", "bold": true, "comment": true, "comment_abc123": true})
		assert.Check(t, slateparser.OnlyCommentsAdded(oldText, newText))
	})

	// Slate table elements have colSizes []float64 on the element node (not on leaf nodes),
	// and deeply nested children: table → tr → td → p → {text: "..."}
	t.Run("table with colSizes, no changes", func(t *testing.T) {
		makeTable := func() []any {
			return []any{map[string]any{
				"type":     "table",
				"id":       "tbl1",
				"colSizes": []any{0.0, 190.45703125, 296.11328125},
				"children": []any{
					map[string]any{
						"type": "tr", "id": "tr1",
						"children": []any{
							map[string]any{
								"type": "td", "id": "td1",
								"children": []any{
									map[string]any{
										"type": "p", "id": "p1",
										"children": []any{map[string]any{"text": "Version"}},
									},
								},
							},
							map[string]any{
								"type": "td", "id": "td2",
								"children": []any{
									map[string]any{
										"type": "p", "id": "p2",
										"children": []any{map[string]any{"text": "Date"}},
									},
								},
							},
						},
					},
				},
			}}
		}
		assert.Check(t, slateparser.OnlyCommentsAdded(makeTable(), makeTable()))
	})

	t.Run("table with colSizes, comment added to leaf", func(t *testing.T) {
		makeTableLeaf := func(leaf map[string]any) []any {
			return []any{map[string]any{
				"type":     "table",
				"id":       "tbl1",
				"colSizes": []any{0.0, 190.45703125},
				"children": []any{
					map[string]any{
						"type": "tr", "id": "tr1",
						"children": []any{
							map[string]any{
								"type": "td", "id": "td1",
								"children": []any{
									map[string]any{
										"type": "p", "id": "p1",
										"children": []any{leaf},
									},
								},
							},
						},
					},
				},
			}}
		}
		oldText := makeTableLeaf(map[string]any{"text": "Version"})
		newText := makeTableLeaf(map[string]any{"text": "Version", "comment": true, "comment_abc": true})
		assert.Check(t, slateparser.OnlyCommentsAdded(oldText, newText))
	})

	t.Run("table with colSizes, text changed in leaf", func(t *testing.T) {
		makeTableLeaf := func(text string) []any {
			return []any{map[string]any{
				"type":     "table",
				"id":       "tbl1",
				"colSizes": []any{0.0, 190.45703125},
				"children": []any{
					map[string]any{
						"type": "tr", "id": "tr1",
						"children": []any{
							map[string]any{
								"type": "td", "id": "td1",
								"children": []any{
									map[string]any{
										"type": "p", "id": "p1",
										"children": []any{map[string]any{"text": text}},
									},
								},
							},
						},
					},
				},
			}}
		}
		assert.Check(t, !slateparser.OnlyCommentsAdded(makeTableLeaf("Version"), makeTableLeaf("Changed")))
	})

	// td cells in some Slate table plugins carry colSpan/rowSpan as float64
	t.Run("table cell with colSpan and rowSpan, comment added to leaf", func(t *testing.T) {
		makeCell := func(leaf map[string]any) []any {
			return []any{map[string]any{
				"type": "table", "id": "tbl1",
				"children": []any{
					map[string]any{
						"type": "tr", "id": "tr1",
						"children": []any{
							map[string]any{
								"type": "td", "id": "td1",
								"colSpan": float64(1), "rowSpan": float64(1),
								"children": []any{
									map[string]any{
										"type": "p", "id": "p1",
										"children": []any{leaf},
									},
								},
							},
						},
					},
				},
			}}
		}
		oldText := makeCell(map[string]any{"text": "hello"})
		newText := makeCell(map[string]any{"text": "hello", "comment": true})
		assert.Check(t, slateparser.OnlyCommentsAdded(oldText, newText))
	})

	// list items use indent (float64) and listStyleType (string) on the element, not the leaf
	t.Run("list item with indent, no changes", func(t *testing.T) {
		makeList := func(leaf map[string]any) []any {
			return []any{map[string]any{
				"type":          "p",
				"indent":        float64(1),
				"listStyleType": "disc",
				"children":      []any{leaf},
			}}
		}
		oldText := makeList(map[string]any{"text": "Confidentiality Policy"})
		newText := makeList(map[string]any{"text": "Confidentiality Policy"})
		assert.Check(t, slateparser.OnlyCommentsAdded(oldText, newText))
	})

	t.Run("list item with indent, comment added to leaf", func(t *testing.T) {
		makeList := func(leaf map[string]any) []any {
			return []any{map[string]any{
				"type":          "p",
				"indent":        float64(1),
				"listStyleType": "disc",
				"children":      []any{leaf},
			}}
		}
		oldText := makeList(map[string]any{"text": "Confidentiality Policy"})
		newText := makeList(map[string]any{"text": "Confidentiality Policy", "comment": true, "comment_xyz": true})
		assert.Check(t, slateparser.OnlyCommentsAdded(oldText, newText))
	})
}
