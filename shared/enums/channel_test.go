package enums_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/shared/enums"
)

func TestToChannel(t *testing.T) {
	testCases := []struct {
		input    string
		expected enums.Channel
	}{
		{
			input:    "in_app",
			expected: enums.ChannelInApp,
		},
		{
			input:    "slack",
			expected: enums.ChannelSlack,
		},
		{
			input:    "email",
			expected: enums.ChannelEmail,
		},
		{
			input:    "IN_APP",
			expected: enums.ChannelInApp,
		},
		{
			input:    "Slack",
			expected: enums.ChannelSlack,
		},
		{
			input:    "invalid",
			expected: enums.ChannelInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Convert %s to Channel", tc.input), func(t *testing.T) {
			result := enums.ToChannel(tc.input)
			assert.Equal(t, tc.expected, *result)
		})
	}
}
