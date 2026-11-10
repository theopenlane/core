package enums

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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
			assert.Check(t, is.Nil(result), "ToRiskStatus(%s) should be nil", test.input)
		} else {
			assert.Check(t, result != nil, "ToRiskStatus(%s) should not be nil", test.input)
			assert.Check(t, is.Equal(*test.expected, *result), "ToRiskStatus(%s) should be %s", test.input, *test.expected)
		}
	}
}
