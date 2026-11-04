package enums_test

import (
	"testing"

	"github.com/theopenlane/core/common/enums"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestControlImplementationStatus(t *testing.T) {
	testCases := []struct {
		name     string
		status   string
		expected enums.ControlImplementationStatus
	}{
		{
			name:     "planned",
			status:   "planned",
			expected: enums.ControlImplementationStatusPlanned,
		},
		{
			name:     "implemented",
			status:   "implemented",
			expected: enums.ControlImplementationStatusImplemented,
		},
		{
			name:     "partially implemented",
			status:   "partially_implemented",
			expected: enums.ControlImplementationStatusPartiallyImplemented,
		},
		{
			name:     "inherited",
			status:   "INHERITED",
			expected: enums.ControlImplementationStatusInherited,
		},
		{
			name:     "not applicable",
			status:   "NOT_APPLICABLE",
			expected: enums.ControlImplementationStatusNotApplicable,
		},
		{
			name:     "empty",
			status:   "",
			expected: enums.ControlImplementationStatusPlanned,
		},
		{
			name:     "invalid",
			status:   "UNKNOWN",
			expected: enums.ControlImplementationStatusInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := enums.ToControlImplementationStatus(tc.status)
			assert.Check(t, is.Equal(tc.expected, *res))
		})
	}
}
