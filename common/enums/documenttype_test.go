package enums_test

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
)

func TestToDocumentType(t *testing.T) {
	testCases := []struct {
		input    string
		expected enums.DocumentType
	}{
		{
			input:    "roottemplate",
			expected: enums.RootTemplate,
		},
		{
			input:    "document",
			expected: enums.Document,
		},
		{
			input:    "UNKNOWN",
			expected: enums.DocumentTypeInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Convert %s to Document Type", tc.input), func(t *testing.T) {
			result := enums.ToDocumentType(tc.input)
			assert.Check(t, is.Equal(tc.expected, *result))
		})
	}
}
