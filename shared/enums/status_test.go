package enums_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/shared/enums"
)

func TestToUserStatus(t *testing.T) {
	testCases := []struct {
		input    string
		expected enums.UserStatus
	}{
		{
			input:    "active",
			expected: enums.UserStatusActive,
		},
		{
			input:    "inactive",
			expected: enums.UserStatusInactive,
		},
		{
			input:    "DEACTIVATED",
			expected: enums.UserStatusDeactivated,
		},
		{
			input:    "suspended",
			expected: enums.UserStatusSuspended,
		},
		{
			input:    "onboarding",
			expected: enums.UserStatusOnboarding,
		},
		{
			input:    "UNKNOWN",
			expected: enums.UserStatusInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Convert %s to UserStatus", tc.input), func(t *testing.T) {
			result := enums.ToUserStatus(tc.input)
			assert.Equal(t, tc.expected, *result)
		})
	}
}
