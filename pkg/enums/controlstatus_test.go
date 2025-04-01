package enums_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/core/pkg/enums"
)

func TestControlStatus(t *testing.T) {
	testCases := []struct {
		name     string
		role     string
		expected enums.ControlStatus
	}{
		{
			name:     "preparing",
			role:     "preparing",
			expected: enums.ControlStatusPreparing,
		},
		{
			name:     "needs approval",
			role:     "needs_approval",
			expected: enums.ControlStatusNeedsApproval,
		},
		{
			name:     "request changes",
			role:     "changes_requested",
			expected: enums.ControlStatusChangesRequested,
		},
		{
			name:     "approved",
			role:     "approved",
			expected: enums.ControlStatusApproved,
		},
		{
			name:     "archived",
			role:     "archived",
			expected: enums.ControlStatusArchived,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := enums.ToControlStatus(tc.role)
			assert.Equal(t, tc.expected, *res)
		})
	}
}
