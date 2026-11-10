package enums_test

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
)

func TestToTier(t *testing.T) {
	testCases := []struct {
		input    string
		expected enums.Tier
	}{
		{
			input:    "free",
			expected: enums.TierFree,
		},
		{
			input:    "pro",
			expected: enums.TierPro,
		},
		{
			input:    "enterprise",
			expected: enums.TierEnterprise,
		},
		{
			input:    "UNKNOWN",
			expected: enums.TierInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Convert %s to Tier", tc.input), func(t *testing.T) {
			result := enums.ToTier(tc.input)
			assert.Check(t, is.Equal(tc.expected, *result))
		})
	}
}
