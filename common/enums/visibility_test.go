package enums_test

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
)

func TestToVisibility(t *testing.T) {
	testCases := []struct {
		input    string
		expected enums.Visibility
	}{
		{
			input:    "public",
			expected: enums.VisibilityPublic,
		},
		{
			input:    "private",
			expected: enums.VisibilityPrivate,
		},

		{
			input:    "UNKNOWN",
			expected: enums.VisibilityInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Convert %s to Visibility", tc.input), func(t *testing.T) {
			result := enums.ToGroupVisibility(tc.input)
			assert.Check(t, is.Equal(tc.expected, *result))
		})
	}
}
