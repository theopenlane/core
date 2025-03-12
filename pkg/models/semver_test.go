package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/core/pkg/models"
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
			assert.Equal(t, &tt.expected, got)
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
			got := models.ToSemverVersion(&tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
func TestSemverVersionBumpMajor(t *testing.T) {
	tests := []struct {
		name     string
		version  models.SemverVersion
		expected models.SemverVersion
	}{
		{
			name: "Bump major version",
			version: models.SemverVersion{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			expected: models.SemverVersion{
				Major: 2,
				Minor: 0,
				Patch: 0,
			},
		},
		{
			name: "Bump major version with pre-release",
			version: models.SemverVersion{
				Major:      1,
				Minor:      2,
				Patch:      3,
				PreRelease: "alpha",
			},
			expected: models.SemverVersion{
				Major: 2,
				Minor: 0,
				Patch: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.version.BumpMajor()
			assert.Equal(t, tt.expected, tt.version)
		})
	}
}

func TestSemverVersion_BumpMinor(t *testing.T) {
	tests := []struct {
		name     string
		version  models.SemverVersion
		expected models.SemverVersion
	}{
		{
			name: "Bump minor version",
			version: models.SemverVersion{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			expected: models.SemverVersion{
				Major: 1,
				Minor: 3,
				Patch: 0,
			},
		},
		{
			name: "Bump minor version with pre-release",
			version: models.SemverVersion{
				Major:      1,
				Minor:      2,
				Patch:      3,
				PreRelease: "alpha",
			},
			expected: models.SemverVersion{
				Major: 1,
				Minor: 3,
				Patch: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.version.BumpMinor()
			assert.Equal(t, tt.expected, tt.version)
		})
	}
}

func TestSemverVersion_BumpPatch(t *testing.T) {
	tests := []struct {
		name     string
		version  models.SemverVersion
		expected models.SemverVersion
	}{
		{
			name: "Bump patch version",
			version: models.SemverVersion{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			expected: models.SemverVersion{
				Major: 1,
				Minor: 2,
				Patch: 4,
			},
		},
		{
			name: "Bump patch version with pre-release",
			version: models.SemverVersion{
				Major:      1,
				Minor:      2,
				Patch:      3,
				PreRelease: "alpha",
			},
			expected: models.SemverVersion{
				Major: 1,
				Minor: 2,
				Patch: 4,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.version.BumpPatch()
			assert.Equal(t, tt.expected, tt.version)
		})
	}
}
func TestBumpMajorVersionString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Bump major version",
			input:    "v1.2.3",
			expected: "v2.0.0",
		},
		{
			name:     "Bump major version with pre-release",
			input:    "v1.2.3-alpha",
			expected: "v2.0.0",
		},
		{
			name:     "Bump major version without v prefix",
			input:    "1.2.3",
			expected: "v2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := models.BumpMajorVersionString(&tt.input)
			assert.Equal(t, &tt.expected, got)
		})
	}
}

func TestBumpMinorVersionString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Bump minor version",
			input:    "v1.2.3",
			expected: "v1.3.0",
		},
		{
			name:     "Bump minor version with pre-release",
			input:    "v1.2.3-alpha",
			expected: "v1.3.0",
		},
		{
			name:     "Bump minor version without v prefix",
			input:    "1.2.3",
			expected: "v1.3.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := models.BumpMinorVersionString(&tt.input)
			assert.Equal(t, &tt.expected, got)
		})
	}
}

func TestBumpPatchVersionString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Bump patch version",
			input:    "v1.2.3",
			expected: "v1.2.4",
		},
		{
			name:     "Bump patch version with pre-release",
			input:    "v1.2.3-alpha",
			expected: "v1.2.4",
		},
		{
			name:     "Bump patch version without v prefix",
			input:    "1.2.3",
			expected: "v1.2.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := models.BumpPatchVersionString(&tt.input)
			assert.Equal(t, &tt.expected, got)
		})
	}
}
