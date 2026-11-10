package enums_test

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
)

func TestToNotificationType(t *testing.T) {
	testCases := []struct {
		input    string
		expected enums.NotificationType
	}{
		{
			input:    "organization",
			expected: enums.NotificationTypeOrganization,
		},
		{
			input:    "user",
			expected: enums.NotificationTypeUser,
		},
		{
			input:    "ORGANIZATION",
			expected: enums.NotificationTypeOrganization,
		},
		{
			input:    "User",
			expected: enums.NotificationTypeUser,
		},
		{
			input:    "invalid",
			expected: enums.NotificationTypeInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Convert %s to NotificationType", tc.input), func(t *testing.T) {
			result := enums.ToNotificationType(tc.input)
			assert.Check(t, is.Equal(tc.expected, *result))
		})
	}
}
