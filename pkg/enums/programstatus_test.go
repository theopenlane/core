package enums_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/enums"
)

func TestToProgramStatus(t *testing.T) {
	testCases := []struct {
		input    string
		expected enums.ProgramStatus
	}{
		{
			input:    "not_started",
			expected: enums.ProgramStatusNotStarted,
		},
		{
			input:    "in_progress",
			expected: enums.ProgramStatusInProgress,
		},
		{
			input:    "READY_FOR_AUDITOR",
			expected: enums.ProgramStatusReadyForAuditor,
		},
		{
			input:    "COMPLETED",
			expected: enums.ProgramStatusCompleted,
		},
		{
			input:    "action_required",
			expected: enums.ProgramStatusActionRequired,
		},
		{
			input:    "UNKNOWN",
			expected: enums.ProgramStatusInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Convert %s to ProgramStatus", tc.input), func(t *testing.T) {
			result := enums.ToProgramStatus(tc.input)
			assert.Equal(t, tc.expected, *result)
		})
	}
}
