package enums_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/common/enums"
)

func TestToPriority(t *testing.T) {
	testCases := []struct {
		input    string
		expected enums.Priority
	}{
		{
			input:    "medium",
			expected: enums.PriorityMedium,
		},
		{
			input:    "low",
			expected: enums.PriorityLow,
		},
		{
			input:    "HIGH",
			expected: enums.PriorityHigh,
		},
		{
			input:    "Critical",
			expected: enums.PriorityCritical,
		},
		{
			input:    "meow",
			expected: enums.PriorityInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Convert %s to Priority", tc.input), func(t *testing.T) {
			result := enums.ToPriority(tc.input)
			assert.Equal(t, tc.expected, *result)
		})
	}
}
