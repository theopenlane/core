package domainscan

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestMergeStrings(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want []string
	}{
		{
			name: "unions and dedupes preserving first-seen order",
			a:    []string{"SOC 2", "GDPR"},
			b:    []string{"GDPR", "ISO 27001"},
			want: []string{"SOC 2", "GDPR", "ISO 27001"},
		},
		{
			name: "empty values are dropped",
			a:    []string{"", "SOC 2"},
			b:    []string{""},
			want: []string{"SOC 2"},
		},
		{
			name: "both empty returns empty slice",
			a:    nil,
			b:    nil,
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeStrings(tt.a, tt.b)

			assert.Check(t, is.DeepEqual(tt.want, got))
		})
	}
}
