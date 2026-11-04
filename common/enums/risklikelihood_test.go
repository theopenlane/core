package enums

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestToRiskLikelihood(t *testing.T) {
	tests := []struct {
		input    string
		expected RiskLikelihood
	}{
		{"UNLIKELY", RiskLikelihoodLow},
		{"LIKELY", RiskLikelihoodMid},
		{"HIGHLY_LIKELY", RiskLikelihoodHigh},
		{"unknown", RiskLikelihoodInvalid},
		{"", RiskLikelihoodInvalid},
		{"unlikely", RiskLikelihoodLow},
		{"likely", RiskLikelihoodMid},
		{"highly_likely", RiskLikelihoodHigh},
	}

	for _, test := range tests {
		result := ToRiskLikelihood(test.input)
		assert.Check(t, is.Equal(test.expected, *result), "ToRiskLikelihood(%q)", test.input)
	}
}
