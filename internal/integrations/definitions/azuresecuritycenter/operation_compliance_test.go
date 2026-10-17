package azuresecuritycenter

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
)

func TestIsCategoryRollup(t *testing.T) {
	tests := []struct {
		name           string
		assessmentType string
		want           bool
	}{
		{
			name:           "category rollup",
			assessmentType: "AssessmentCategory",
			want:           true,
		},
		{
			name:           "case insensitive",
			assessmentType: "assessmentcategory",
			want:           true,
		},
		{
			name:           "single recommendation",
			assessmentType: "AssessmentResult",
			want:           false,
		},
		{
			name:           "empty type",
			assessmentType: "",
			want:           false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Check(t, isCategoryRollup(complianceAssessment{Properties: complianceAssessmentProperties{AssessmentType: tc.assessmentType}}) == tc.want)
		})
	}
}

func TestComplianceCheckStatus(t *testing.T) {
	tests := []struct {
		name  string
		state string
		want  enums.CheckStatus
	}{
		{
			name:  "passed",
			state: "Passed",
			want:  enums.CheckStatusPass,
		},
		{
			name:  "failed lowercase",
			state: "failed",
			want:  enums.CheckStatusFail,
		},
		{
			name:  "skipped",
			state: "Skipped",
			want:  enums.CheckStatusUnknown,
		},
		{
			name:  "unsupported",
			state: "Unsupported",
			want:  enums.CheckStatusUnknown,
		},
		{
			name:  "empty state",
			state: "",
			want:  enums.CheckStatusUnknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Check(t, complianceCheckStatus(tc.state) == tc.want)
		})
	}
}

func TestStandardSelected(t *testing.T) {
	tests := []struct {
		name      string
		standards []string
		standard  string
		want      bool
	}{
		{
			name:      "nil selects all",
			standards: nil,
			standard:  "Azure-SOC-2023",
			want:      true,
		},
		{
			name:      "empty selects all",
			standards: []string{},
			standard:  "Azure-SOC-2023",
			want:      true,
		},
		{
			name:      "case insensitive match",
			standards: []string{"azure-soc-2023"},
			standard:  "Azure-SOC-2023",
			want:      true,
		},
		{
			name:      "not listed",
			standards: []string{"Azure-SOC-2023"},
			standard:  "Microsoft-cloud-security-benchmark",
			want:      false,
		},
		{
			name:      "empty standard not listed",
			standards: []string{"Azure-SOC-2023"},
			standard:  "",
			want:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Check(t, standardSelected(tc.standards, tc.standard) == tc.want)
		})
	}
}
