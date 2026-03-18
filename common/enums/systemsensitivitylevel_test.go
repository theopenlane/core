package enums_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/core/common/enums"
)

func TestSystemSensitivityLevel(t *testing.T) {
	testCases := []struct {
		name     string
		level    string
		expected enums.SystemSensitivityLevel
	}{
		{
			name:     "low",
			level:    "low",
			expected: enums.SystemSensitivityLevelLow,
		},
		{
			name:     "moderate",
			level:    "MODERATE",
			expected: enums.SystemSensitivityLevelModerate,
		},
		{
			name:     "high",
			level:    "HIGH",
			expected: enums.SystemSensitivityLevelHigh,
		},
		{
			name:     "unknown",
			level:    "UNKNOWN",
			expected: enums.SystemSensitivityLevelUnknown,
		},
		{
			name:     "empty",
			level:    "",
			expected: enums.SystemSensitivityLevelUnknown,
		},
		{
			name:     "invalid",
			level:    "SECRET",
			expected: enums.SystemSensitivityLevelInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := enums.ToSystemSensitivityLevel(tc.level)
			assert.Equal(t, tc.expected, *res)
		})
	}
}
