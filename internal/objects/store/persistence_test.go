package store

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/pkg/objects"
)

func TestGetOrgOwnerIDUsesSingleAuthorizedOrg(t *testing.T) {
	t.Parallel()

	orgID := ulids.New().String()

	ctx := auth.WithCaller(context.Background(), &auth.Caller{
		OrganizationIDs: []string{orgID},
	})

	id, err := getOrgOwnerID(ctx, objects.File{})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(orgID, id))
}

func TestFileCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		file objects.File
		want string
	}{
		{
			name: "evidence maps to evidence",
			file: objects.File{CorrelatedObjectType: "Evidence"},
			want: "evidence",
		},
		{
			name: "user maps to user",
			file: objects.File{CorrelatedObjectType: "User"},
			want: "user",
		},
		{
			name: "organization maps to organization",
			file: objects.File{CorrelatedObjectType: "Organization"},
			want: "organization",
		},
		{
			name: "trust center setting maps to trust center setting",
			file: objects.File{CorrelatedObjectType: "TrustCenterSetting"},
			want: "trust_center_setting",
		},
		{
			name: "unmapped file has no category",
			file: objects.File{CorrelatedObjectType: "Unknown"},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Check(t, is.Equal(tt.want, getCategoryNameForSchema(tt.file)))
		})
	}
}

func TestResolveProvidedExtension(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		file objects.File
		want string
	}{
		{
			name: "simple png",
			file: objects.File{OriginalName: "avatar.png"},
			want: ".png",
		},
		{
			name: "filename with spaces, parentheses, and multiple dots",
			file: objects.File{OriginalName: "Captura de pantalla 2026-03-26 a la(s) 5.35.31 p. m..png"},
			want: ".png",
		},
		{
			name: "pdf",
			file: objects.File{OriginalName: "Q1-report.final.v2.pdf"},
			want: ".pdf",
		},
		{
			name: "uppercase extension is preserved",
			file: objects.File{OriginalName: "SCAN.PDF"},
			want: ".PDF",
		},
		{
			name: "no extension returns empty",
			file: objects.File{OriginalName: "README"},
			want: "",
		},
		{
			name: "empty filename returns empty",
			file: objects.File{OriginalName: ""},
			want: "",
		},
		{
			name: "explicit ProvidedExtension takes precedence over OriginalName",
			file: objects.File{OriginalName: "avatar.png", ProvidedExtension: ".jpg"},
			want: ".jpg",
		},
		{
			name: "explicit ProvidedExtension used when OriginalName is empty",
			file: objects.File{ProvidedExtension: ".png"},
			want: ".png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Check(t, is.Equal(tt.want, resolveProvidedExtension(tt.file)))
		})
	}
}

func TestResolveFileName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		file objects.File
		want string
	}{
		{
			name: "uses explicit file name",
			file: objects.File{
				FileMetadata: objects.FileMetadata{Name: "Profile Photo"},
				OriginalName: "avatar.png",
			},
			want: "Profile Photo",
		},
		{
			name: "falls back to metadata name",
			file: objects.File{
				Metadata:     map[string]any{"name": "April Invoice"},
				OriginalName: "invoice.pdf",
			},
			want: "April Invoice",
		},
		{
			name: "falls back to provided file name",
			file: objects.File{
				OriginalName: "contract.pdf",
			},
			want: "contract.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Check(t, is.Equal(tt.want, retrieveFileName(tt.file)))
		})
	}
}
