package hooks

import (
	"testing"

	"github.com/theopenlane/core/common/enums"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func Test_impactLevelFromScore(t *testing.T) {
	tests := []struct {
		name  string
		score int
		want  enums.RiskImpact
	}{
		{
			name:  "score of 0 should return low impact",
			score: 0,
			want:  enums.RiskImpactLow,
		},
		{
			name:  "score of 5 should return low impact",
			score: 5,
			want:  enums.RiskImpactModerate,
		},
		{
			name:  "score of 6 should return moderate impact",
			score: 6,
			want:  enums.RiskImpactModerate,
		},
		{
			name:  "score of 10 should return moderate impact",
			score: 10,
			want:  enums.RiskImpactHigh,
		},
		{
			name:  "score of 11 should return high impact",
			score: 11,
			want:  enums.RiskImpactHigh,
		},
		{
			name:  "score of 15 should return high impact",
			score: 15,
			want:  enums.RiskImpactHigh,
		},
		{
			name:  "score of 18 should return critical impact",
			score: 18,
			want:  enums.RiskImpactCritical,
		},
		{
			name:  "negative score should return low impact",
			score: -1,
			want:  enums.RiskImpactLow,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := impactLevelFromScore(tt.score)
			assert.Check(t, is.Equal(got, tt.want))
		})
	}
}
