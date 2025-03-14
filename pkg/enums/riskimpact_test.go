package enums

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, test.expected, *result, "ToRiskImpact(%q)", test.input)
	}
}
