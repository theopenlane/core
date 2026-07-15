package hooks

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestNewDomains(t *testing.T) {
	tests := []struct {
		name       string
		oldDomains []string
		candidates []string
		want       []string
	}{
		{
			name:       "no old domains, all candidates are new",
			oldDomains: nil,
			candidates: []string{"example.com", "example.org"},
			want:       []string{"example.com", "example.org"},
		},
		{
			name:       "candidate already present in old domains is not new",
			oldDomains: []string{"example.com"},
			candidates: []string{"example.com"},
			want:       []string{},
		},
		{
			name:       "only the newly added domain is returned",
			oldDomains: []string{"example.com"},
			candidates: []string{"example.com", "example.org"},
			want:       []string{"example.org"},
		},
		{
			name:       "duplicate candidates are deduplicated",
			oldDomains: []string{"example.com"},
			candidates: []string{"example.org", "example.org"},
			want:       []string{"example.org"},
		},
		{
			name:       "no candidates returns no new domains",
			oldDomains: []string{"example.com"},
			candidates: nil,
			want:       []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newDomains(tt.oldDomains, tt.candidates)

			assert.Check(t, is.DeepEqual(tt.want, got))
		})
	}
}
