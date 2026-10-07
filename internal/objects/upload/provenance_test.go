package upload

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	pkgobjects "github.com/theopenlane/core/v2/pkg/objects"
)

const testSHA256 = "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"

func TestValidateProvenance(t *testing.T) {
	tests := []struct {
		name       string
		objectType string
		claims     map[string]any
		noContent  bool
		wantErr    error
	}{
		{
			name:   "matching hash is accepted",
			claims: map[string]any{"artifact_sha256": testSHA256},
		},
		{
			name:   "hash comparison ignores hex case",
			claims: map[string]any{"artifact_sha256": strings.ToUpper(testSHA256)},
		},
		{
			name:   "absent hash is accepted",
			claims: map[string]any{"schema_version": 1},
		},
		{
			name:   "empty claims are accepted",
			claims: map[string]any{},
		},
		{
			name: "no claims are accepted",
		},
		{
			name:       "no claims on non evidence uploads are accepted",
			objectType: "TrustCenterDoc",
		},
		{
			name:       "non evidence uploads are rejected",
			objectType: "TrustCenterDoc",
			claims:     map[string]any{},
			wantErr:    ErrProvenanceNotSupported,
		},
		{
			name:    "mismatched hash is rejected",
			claims:  map[string]any{"artifact_sha256": strings.Repeat("0", len(testSHA256))},
			wantErr: ErrProvenanceHashMismatch,
		},
		{
			name:    "malformed hash is rejected",
			claims:  map[string]any{"artifact_sha256": "sha256:" + testSHA256},
			wantErr: ErrInvalidProvenanceHash,
		},
		{
			name:    "non-string hash is rejected",
			claims:  map[string]any{"artifact_sha256": 42},
			wantErr: ErrInvalidProvenanceHash,
		},
		{
			name:    "claims over the size limit are rejected",
			claims:  map[string]any{"page_title": strings.Repeat("a", maxProvenanceBytes)},
			wantErr: ErrProvenanceTooLarge,
		},
		{
			name:      "missing content is rejected",
			claims:    map[string]any{},
			noContent: true,
			wantErr:   ErrProvenanceWithoutContent,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			file := pkgobjects.File{
				CorrelatedObjectType: ent.TypeEvidence,
				SHA256:               []byte(testSHA256),
				ProvenanceClaims:     tc.claims,
			}

			if tc.objectType != "" {
				file.CorrelatedObjectType = tc.objectType
			}

			if tc.noContent {
				file.SHA256 = nil
			}

			err := validateProvenance(file)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)

				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestProvenanceErrorsShareParentSentinel(t *testing.T) {
	for _, err := range []error{
		ErrProvenanceNotSupported,
		ErrProvenanceTooLarge,
		ErrProvenanceWithoutContent,
		ErrInvalidProvenanceHash,
		ErrProvenanceHashMismatch,
	} {
		assert.ErrorIs(t, err, ErrInvalidProvenance)
	}
}
