package notifications

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectVersionBump(t *testing.T) {
	tests := []struct {
		name        string
		oldRevision string
		newRevision string
		expected    string
	}{
		{
			name:        "major bump",
			oldRevision: "v1.0.0",
			newRevision: "v2.0.0",
			expected:    "major",
		},
		{
			name:        "minor bump",
			oldRevision: "v1.0.0",
			newRevision: "v1.1.0",
			expected:    "minor",
		},
		{
			name:        "patch only bump",
			oldRevision: "v1.0.0",
			newRevision: "v1.0.1",
			expected:    "",
		},
		{
			name:        "same version",
			oldRevision: "v1.2.3",
			newRevision: "v1.2.3",
			expected:    "",
		},
		{
			name:        "major and minor bump",
			oldRevision: "v1.2.3",
			newRevision: "v2.0.0",
			expected:    "major",
		},
		{
			name:        "prerelease to release minor bump",
			oldRevision: "v1.0.0-draft",
			newRevision: "v1.1.0",
			expected:    "minor",
		},
		{
			name:        "prerelease same major minor",
			oldRevision: "v1.0.0",
			newRevision: "v1.0.1-draft",
			expected:    "",
		},
		{
			name:        "empty old revision",
			oldRevision: "",
			newRevision: "v1.0.0",
			expected:    "major",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, detectVersionBump(tt.oldRevision, tt.newRevision))
		})
	}
}
