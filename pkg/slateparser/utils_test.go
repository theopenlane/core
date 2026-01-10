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
