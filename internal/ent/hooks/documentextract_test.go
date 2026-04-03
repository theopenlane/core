package hooks

import (
	"slices"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestFindControlMatches(t *testing.T) {
	tests := []struct {
		name     string
		details  string
		controls controlMapping
		want     *edgeLinks
	}{
		{
			name:    "matches control ref code exactly",
			details: "This is a reference to control AC-2 in the document details.",
			controls: controlMapping{
				"AC-2": {
					id:      "control-id-1",
					fefCode: "AC-2",
				},
			},
			want: &edgeLinks{
				controlIDs: []string{"control-id-1"},
			},
		},
		{
			name:    "matches control ref code with punctuation around it",
			details: "The document details mention control AC-2, which is important.",
			controls: controlMapping{
				"AC-2": {
					id:      "control-id-1",
					fefCode: "AC-2",
				},
			},
			want: &edgeLinks{
				controlIDs: []string{"control-id-1"},
			},
		},
		{
			name:    "does not match control ref code when it's part of another word",
			details: "The document details mention control AC-200, which is important.",
			controls: controlMapping{
				"AC-2": {
					id:      "control-id-1",
					fefCode: "AC-2",
				},
			},
			want: &edgeLinks{},
		},
		{
			name:    "matches multiple control ref codes",
			details: "The document details mention controls AC-2 and AC-3, which are important.",
			controls: controlMapping{
				"AC-2": {
					id:      "control-id-1",
					fefCode: "AC-2",
				},
				"AC-3": {
					id:      "control-id-2",
					fefCode: "AC-3",
				},
			},
			want: &edgeLinks{
				controlIDs: []string{"control-id-1", "control-id-2"},
			},
		},
		{
			name:    "matches control ref code in different cases",
			details: "The document details mention control ac-2, which is important.",
			controls: controlMapping{
				"AC-2": {
					id:      "control-id-1",
					fefCode: "AC-2",
				},
			},
			want: &edgeLinks{
				controlIDs: []string{"control-id-1"},
			},
		},
		{
			name:    "ignore numeric markdown headers",
			details: "## 4.1 this is not a control reference",
			controls: controlMapping{
				"4.1": {
					id:      "control-id-1",
					fefCode: "4.1",
				},
			},
			want: &edgeLinks{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findControlMatches(tt.details, tt.controls)

			// sort for consistent comparison since order of matches is not guaranteed
			slices.Sort(got.controlIDs)
			slices.Sort(got.subcontrolIDs)

			assert.Check(t, is.DeepEqual(tt.want.controlIDs, got.controlIDs))
			assert.Check(t, is.DeepEqual(tt.want.subcontrolIDs, got.subcontrolIDs))
		})
	}
}

func TestFindVersion(t *testing.T) {
	tests := []struct {
		name    string
		details string
		want    string
	}{
		{
			name:    "finds version with 'Version:' prefix",
			details: "Version: 1.0",
			want:    "v1.0.0",
		},
		{
			name:    "finds version with 'version:' prefix in different case",
			details: "version: 2.5",
			want:    "v2.5.0",
		},
		{
			name:    "finds version with 'version:' prefix in different case, only single digit",
			details: "version: 2",
			want:    "v2.0.0",
		},
		{
			name:    "returns empty string when no version is present",
			details: "This document is about control AC-2.",
			want:    "",
		},
		{
			name:    "version in middle of details",
			details: "This document is about control AC-2. Version: 1.0",
			want:    "",
		},
		{
			name:    "finds version with 'version:' but with a string that cannot be converted to semver",
			details: "version: a",
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findVersion(tt.details)
			assert.Check(t, is.Equal(tt.want, got))
		})
	}
}
