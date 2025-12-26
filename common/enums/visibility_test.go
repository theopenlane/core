package enums_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/common/enums"
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
			assert.Equal(t, tc.expected, *result)
		})
	}
}
