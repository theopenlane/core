package models

import (
	"testing"
)

func TestSemverVersion_String(t *testing.T) {
	tests := []struct {
		name     string
		version  SemverVersion
		expected string
	}{
		{
			name: "Major.Minor.Patch with PreRelease",
			version: SemverVersion{
				Major:      1,
				Minor:      2,
				Patch:      3,
				PreRelease: "alpha",
			},
			expected: "v1.2.3-alpha",
		},
		{
			name: "Patch only",
			version: SemverVersion{

				Patch: 3,
			},
			expected: "v0.0.3",
		},
		{
			name: "Major.Minor.Patch without PreRelease",
			version: SemverVersion{
				Major:      2,
				Minor:      0,
				Patch:      1,
				PreRelease: "",
			},
			expected: "v2.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.String(); got != tt.expected {
				t.Errorf("SemverVersion.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
