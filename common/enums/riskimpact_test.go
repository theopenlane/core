package enums

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestToRiskImpact(t *testing.T) {
	tests := []struct {
		input    string
		expected RiskImpact
	}{
		{"LOW", RiskImpactLow},
		{"low", RiskImpactLow},
		{"MODERATE", RiskImpactModerate},
		{"moderate", RiskImpactModerate},
		{"HIGH", RiskImpactHigh},
		{"high", RiskImpactHigh},
		{"UNKNOWN", RiskImpactInvalid},
		{"", RiskImpactInvalid},
	}

	for _, test := range tests {
		result := ToRiskImpact(test.input)
		assert.Check(t, is.Equal(test.expected, *result), "ToRiskImpact(%q)", test.input)
	}
}
