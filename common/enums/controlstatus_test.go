package enums_test

import (
	"testing"

	"github.com/theopenlane/core/common/enums"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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
			expected: enums.ControlStatusNotImplemented,
		},
		{
			name:     "not implemented",
			status:   "NOT_IMPLEMENTED",
			expected: enums.ControlStatusNotImplemented,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := enums.ToControlStatus(tc.status)
			assert.Check(t, is.Equal(tc.expected, *res))
		})
	}
}
