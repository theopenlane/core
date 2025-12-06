package enums

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToRiskStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected *RiskStatus
	}{
		{"open", &RiskOpen},
		{"IN_PROGRESS", &RiskInProgress},
		{"ongoing", &RiskOngoing},
		{"identified", &RiskIdentified},
		{"MITIGATED", &RiskMitigated},
		{"accepted", &RiskAccepted},
		{"closed", &RiskClosed},
		{"transferred", &RiskTransferred},
		{"archived", &RiskArchived},
		{"invalid", nil},
		{"", nil},
	}

	for _, test := range tests {
		result := ToRiskStatus(test.input)
		if test.expected == nil {
			assert.Nil(t, result, "ToRiskStatus(%s) should be nil", test.input)
		} else {
			assert.NotNil(t, result, "ToRiskStatus(%s) should not be nil", test.input)
			assert.Equal(t, *test.expected, *result, "ToRiskStatus(%s) should be %s", test.input, *test.expected)
		}
	}
}
