package objects

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/common/storagetypes"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

func TestPopulateProviderHints(t *testing.T) {
	file := pkgobjects.File{
		OriginalName:         "evidence.json",
		FieldName:            "uploadFile",
		CorrelatedObjectType: "evidence",
		FileMetadata: pkgobjects.FileMetadata{
			Size:        1024,
			ContentType: "application/json",
		},
	}

	PopulateProviderHints(&file, "org-123")

	require.NotNil(t, file.ProviderHints)
	assert.Equal(t, "org-123", file.ProviderHints.OrganizationID)
	assert.Equal(t, models.CatalogComplianceModule, file.ProviderHints.Module)
	assert.Equal(t, "uploadFile", file.ProviderHints.Metadata["key"])
	assert.Equal(t, "evidence", file.ProviderHints.Metadata["object_type"])
	assert.Equal(t, "1024", file.ProviderHints.Metadata["size_bytes"])
	assert.Equal(t, string(models.CatalogComplianceModule), file.ProviderHints.Metadata["module"])
}

func TestApplyProviderHints(t *testing.T) {
	hints := &storagetypes.ProviderHints{
		PreferredProvider: storagetypes.ProviderType("s3"),
		KnownProvider:     storagetypes.ProviderType("disk"),
		Metadata: map[string]string{
			"size_bytes":            "2048",
			TemplateKindMetadataKey: enums.TemplateKindTrustCenterNda.String(),
		},
	}

	module := models.CatalogComplianceModule
	hints.Module = module

	ctx := ApplyProviderHints(context.Background(), hints)

	pref, ok := PreferredProviderHintFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, storagetypes.ProviderType(pref), storagetypes.ProviderType("s3"))

	known, ok := KnownProviderHintFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, storagetypes.ProviderType(known), storagetypes.ProviderType("disk"))

	resModule, ok := ModuleHintFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, models.OrgModule(resModule), module)

	size, ok := SizeBytesHintFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, int64(size), int64(2048))

	kind, ok := TemplateKindHintFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, enums.TemplateKindTrustCenterNda, enums.TemplateKind(kind))
}
