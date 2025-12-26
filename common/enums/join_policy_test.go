package enums_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/common/enums"
)

func TestToJoinPolicy(t *testing.T) {
	testCases := []struct {
		input    string
		expected enums.JoinPolicy
	}{
		{
			input:    "open",
			expected: enums.JoinPolicyOpen,
		},
		{
			input:    "invite_only",
			expected: enums.JoinPolicyInviteOnly,
		},
		{
			input:    "application_only",
			expected: enums.JoinPolicyApplicationOnly,
		},
		{
			input:    "invite_or_application",
			expected: enums.JoinPolicyInviteOrApplication,
		},
		{
			input:    "UNKNOWN",
			expected: enums.JoinPolicyInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Convert %s to Join Policy", tc.input), func(t *testing.T) {
			result := enums.ToGroupJoinPolicy(tc.input)
			assert.Equal(t, tc.expected, *result)
		})
	}
}
