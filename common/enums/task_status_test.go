package enums_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/common/enums"
)

func TestToTaskStatus(t *testing.T) {
	testCases := []struct {
		input    string
		expected enums.TaskStatus
	}{
		{
			input:    "open",
			expected: enums.TaskStatusOpen,
		},
		{
			input:    "in_progress",
			expected: enums.TaskStatusInProgress,
		},
		{
			input:    "IN_REVIEW",
			expected: enums.TaskStatusInReview,
		},
		{
			input:    "COMPLETED",
			expected: enums.TaskStatusCompleted,
		},
		{
			input:    "wont_do",
			expected: enums.TaskStatusWontDo,
		},
		{
			input:    "UNKNOWN",
			expected: enums.TaskStatusInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Convert %s to TaskStatus", tc.input), func(t *testing.T) {
			result := enums.ToTaskStatus(tc.input)
			assert.Equal(t, tc.expected, *result)
		})
	}
}
