package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/common/models"
)

func TestSemverVersionString(t *testing.T) {
	tests := []struct {
		name     string
		version  models.SemverVersion
		expected string
	}{
		{
			name: "Major.Minor.Patch with PreRelease",
			version: models.SemverVersion{
				Major:      1,
				Minor:      2,
				Patch:      3,
				PreRelease: "alpha",
			},
			expected: "v1.2.3-alpha",
		},
		{
			name: "Patch only",
			version: models.SemverVersion{

				Patch: 3,
			},
			expected: "v0.0.3",
		},
		{
			name: "Major.Minor.Patch without PreRelease",
			version: models.SemverVersion{
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
			got := tt.version.String()
			assert.Equal(t, tt.expected, got)
		})
	}
}
func TestToSemverVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *models.SemverVersion
	}{
		{
			name:  "v1.0.0",
			input: "v1.0.0",
			expected: &models.SemverVersion{
				Major: 1,
				Minor: 0,
				Patch: 0,
			},
		},
		{
			name:  "1.0.0",
			input: "1.0.0",
			expected: &models.SemverVersion{
				Major: 1,
				Minor: 0,
				Patch: 0,
			},
		},
		{
			name:  "v1.0.0-alpha",
			input: "v1.0.0-alpha",
			expected: &models.SemverVersion{
				Major:      1,
				Minor:      0,
				Patch:      0,
				PreRelease: "alpha",
			},
		},
		{
			name:  "1.0.0-alpha",
			input: "1.0.0-alpha",
			expected: &models.SemverVersion{
				Major:      1,
				Minor:      0,
				Patch:      0,
				PreRelease: "alpha",
			},
		},
		{
			name:  "v2.1.3-beta",
			input: "v2.1.3-beta",
			expected: &models.SemverVersion{
				Major:      2,
				Minor:      1,
				Patch:      3,
				PreRelease: "beta",
			},
		},
		{
			name:  "2.1.3-beta",
			input: "2.1.3-beta",
			expected: &models.SemverVersion{
				Major:      2,
				Minor:      1,
				Patch:      3,
				PreRelease: "beta",
			},
		},
		{
			name:  "v0.0.1",
			input: "v0.0.1",
			expected: &models.SemverVersion{
				Major: 0,
				Minor: 0,
				Patch: 1,
			},
		},
		{
			name:  "0.0.1",
			input: "0.0.1",
			expected: &models.SemverVersion{
				Major: 0,
				Minor: 0,
				Patch: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := models.ToSemverVersion(&tt.input)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
func TestBumpMajor(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "Bump major version",
			version:  "v1.2.3",
			expected: "v2.0.0",
		},
		{
			name:     "Bump major version with pre-release",
			version:  "v1.2.3-alpha",
			expected: "v2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := models.BumpMajor(tt.version)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, version)
		})
	}
}

func TestBumpMinor(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "Bump minor version",
			version:  "v1.2.3",
			expected: "v1.3.0",
		},
		{
			name:     "Bump minor version with pre-release",
			version:  "v1.2.3-alpha",
			expected: "v1.3.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := models.BumpMinor(tt.version)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, version)
		})
	}
}

func TestBumpPatch(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "Bump patch version",
			version:  "v1.2.3",
			expected: "v1.2.4",
		},
		{
			name:     "Bump patch version with pre-release",
			version:  "v1.2.3-alpha",
			expected: "v1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := models.BumpPatch(tt.version)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, version)
		})
	}
}
func TestSetPreRelease(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "Set pre-release on version without pre-release",
			version:  "v1.2.3",
			expected: "v1.2.4-draft",
		},
		{
			name:     "Set pre-release on version with pre-release",
			version:  "v1.2.3-draft",
			expected: "v1.2.3-draft-1",
		},
		{
			name:     "Set pre-release on version with pre-release",
			version:  "v1.2.3-draft-1",
			expected: "v1.2.3-draft-2",
		},
		{
			name:     "Set pre-release on version with no patch",
			version:  "v1.2.0",
			expected: "v1.2.1-draft",
		},
		{
			name:     "Set pre-release on version with no minor and patch",
			version:  "v1.0.0",
			expected: "v1.0.1-draft",
		},
		{
			name:     "Set pre-release on version with no major, minor and patch",
			version:  "v0.0.0",
			expected: "v0.0.1-draft",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := models.SetPreRelease(tt.version)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, version)
		})
	}
}
