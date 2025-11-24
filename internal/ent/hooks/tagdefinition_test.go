package hooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSluggify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple word",
			input:    "test",
			expected: "test",
		},
	{
		name:     "camelCase",
		input:    "testCase",
		expected: "testcase",
	},
		{
			name:     "snake_case",
			input:    "test_case",
			expected: "test-case",
		},
		{
			name:     "multiple words with spaces",
			input:    "Test Case Example",
			expected: "test-case-example",
		},
	{
		name:     "all caps",
		input:    "TEST CASE",
		expected: "test-case",
	},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "already kebab-case",
			input:    "test-case",
			expected: "test-case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sluggify(tt.input)
			assert.Equal(t, tt.expected, result, "sluggify(%q) = %q, want %q", tt.input, result, tt.expected)
		})
	}
}
