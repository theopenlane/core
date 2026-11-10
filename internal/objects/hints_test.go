package objects

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/common/storagetypes"
	pkgobjects "github.com/theopenlane/core/v2/pkg/objects"
)

func TestPopulateProviderHints(t *testing.T) {
	file := pkgobjects.File{
		OriginalName:         "evidence.json",
		FieldName:            "uploadFile",
		CorrelatedObjectType: "evidence",
		Size:                 1024,
		ContentType:          "application/json",
	}

	PopulateProviderHints(&file, "org-123")

	assert.Assert(t, file.ProviderHints != nil)
	assert.Check(t, is.Equal("org-123", file.ProviderHints.OrganizationID))
	assert.Check(t, is.Equal(models.CatalogComplianceModule, file.ProviderHints.Module))
	assert.Check(t, is.Equal("uploadFile", file.ProviderHints.Metadata["key"]))
	assert.Check(t, is.Equal("evidence", file.ProviderHints.Metadata["object_type"]))
	assert.Check(t, is.Equal("1024", file.ProviderHints.Metadata["size_bytes"]))
	assert.Check(t, is.Equal(string(models.CatalogComplianceModule), file.ProviderHints.Metadata["module"]))
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
	assert.Assert(t, ok)
	assert.Check(t, is.Equal(storagetypes.ProviderType(pref), storagetypes.ProviderType("s3")))

	known, ok := KnownProviderHintFromContext(ctx)
	assert.Assert(t, ok)
	assert.Check(t, is.Equal(storagetypes.ProviderType(known), storagetypes.ProviderType("disk")))

	resModule, ok := ModuleHintFromContext(ctx)
	assert.Assert(t, ok)
	assert.Check(t, is.Equal(models.OrgModule(resModule), module))

	size, ok := SizeBytesHintFromContext(ctx)
	assert.Assert(t, ok)
	assert.Check(t, is.Equal(int64(size), int64(2048)))

	kind, ok := TemplateKindHintFromContext(ctx)
	assert.Assert(t, ok)
	assert.Check(t, is.Equal(enums.TemplateKindTrustCenterNda, enums.TemplateKind(kind)))
}
