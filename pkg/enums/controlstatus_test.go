package enums_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/core/pkg/enums"
)

func TestControlStatus(t *testing.T) {
	testCases := []struct {
		name     string
		status   string
		expected enums.ControlStatus
	}{
		{
			name:     "preparing",
			status:   "preparing",
			expected: enums.ControlStatusPreparing,
		},
		{
			name:     "needs approval",
			status:   "needs_approval",
			expected: enums.ControlStatusNeedsApproval,
		},
		{
			name:     "request changes",
			status:   "changes_requested",
			expected: enums.ControlStatusChangesRequested,
		},
		{
			name:     "approved",
			status:   "approved",
			expected: enums.ControlStatusApproved,
		},
		{
			name:     "archived",
			status:   "archived",
			expected: enums.ControlStatusArchived,
		},
		{
			name:     "empty",
			status:   "",
			expected: enums.ControlStatusNull,
		},
		{
			name:     "null",
			status:   "NULL",
			expected: enums.ControlStatusNull,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := enums.ToControlStatus(tc.status)
			assert.Equal(t, tc.expected, *res)
		})
	}
}
