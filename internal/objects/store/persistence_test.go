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
			name: "user maps to avatar",
			file: objects.File{CorrelatedObjectType: "User"},
			want: "avatar",
		},
		{
			name: "organization maps to avatar",
			file: objects.File{CorrelatedObjectType: "Organization"},
			want: "avatar",
		},
		{
			name: "trust center setting maps to logo",
			file: objects.File{CorrelatedObjectType: "TrustCenterSetting"},
			want: "logo",
		},
		{
			name: "subprocessor maps to logo",
			file: objects.File{CorrelatedObjectType: "Subprocessor"},
			want: "logo",
		},
		{
			name: "unmapped file has no category",
			file: objects.File{CorrelatedObjectType: "Entity", FieldName: "entityFiles"},
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
