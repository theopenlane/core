package enums_test

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
)

func TestToRegion(t *testing.T) {
	testCases := []struct {
		input    string
		expected enums.Region
	}{
		{
			input:    "amer",
			expected: enums.Amer,
		},
		{
			input:    "EMEA",
			expected: enums.Emea,
		},
		{
			input:    "Apac",
			expected: enums.Apac,
		},
		{
			input:    "UNKNOWN",
			expected: enums.InvalidRegion,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Convert %s to Region", tc.input), func(t *testing.T) {
			result := enums.ToRegion(tc.input)
			assert.Check(t, is.Equal(tc.expected, *result))
		})
	}
}
