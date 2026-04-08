package store

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
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

func TestAutoFileCategoryName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		file objects.File
		want string
	}{
		{
			name: "evidence maps to evidence",
			file: objects.File{CorrelatedObjectType: "Evidence"},
			want: "Evidence",
		},
		{
			name: "user maps to avatar",
			file: objects.File{CorrelatedObjectType: "User"},
			want: "Avatar",
		},
		{
			name: "organization maps to avatar",
			file: objects.File{CorrelatedObjectType: "Organization"},
			want: "Avatar",
		},
		{
			name: "trust center setting maps to logo",
			file: objects.File{CorrelatedObjectType: "TrustCenterSetting"},
			want: "Logo",
		},
		{
			name: "subprocessor maps to logo",
			file: objects.File{CorrelatedObjectType: "Subprocessor"},
			want: "Logo",
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
			assert.Check(t, is.Equal(tt.want, autoFileCategoryName(tt.file)))
		})
	}
}
