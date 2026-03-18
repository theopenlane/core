package graphapi

import (
	"testing"

	"github.com/theopenlane/core/internal/graphapi/common"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestGetStandardRefCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     []string
		expected map[string][]string
		wantErr  bool
	}{
		{
			name:     "Empty input",
			data:     []string{},
			expected: nil,
			wantErr:  false,
		},
		{
			name: "Valid input with single standard",
			data: []string{
				"ISO 27001::A.5.1.1",
				"ISO 27001::A.5.1.2",
			},
			expected: map[string][]string{
				"ISO 27001": {"A.5.1.1", "A.5.1.2"},
			},
			wantErr: false,
		},
		{
			name: "Valid input with multiple standards",
			data: []string{
				"ISO27001::A.5.1.1",
				"NIST800-53::AC-1",
				"ISO27001::A.6.1.1",
				"NIST800-53::AC-2",
			},
			expected: map[string][]string{
				"ISO27001":   {"A.5.1.1", "A.6.1.1"},
				"NIST800-53": {"AC-1", "AC-2"},
			},
			wantErr: false,
		},
		{
			name: "Invalid format - missing colon",
			data: []string{
				"ISO27001A.5.1.1",
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Invalid format - too many colons",
			data: []string{
				"ISO27001::A.5.1.1::extra",
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Mixed valid and invalid inputs",
			data: []string{
				"ISO27001:A.5.1.1",
				"Invalid-Format",
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getStandardRefCodes(tt.data)

			if tt.wantErr {
				assert.ErrorIs(t, err, common.ErrInvalidInput)
				assert.Check(t, result == nil)

				return
			}
			assert.NilError(t, err)
			assert.Check(t, is.DeepEqual(result, tt.expected))
		})
	}
}
