package enums_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/shared/enums"
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
			assert.Equal(t, tc.expected, *result)
		})
	}
}
