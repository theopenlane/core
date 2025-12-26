package enums

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, test.expected, *result, "ToRiskLikelihood(%q)", test.input)
	}
}
