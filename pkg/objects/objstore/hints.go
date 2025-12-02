package objstore

import (
	"context"
	"strconv"

	"github.com/samber/lo"

	"github.com/theopenlane/core/pkg/models"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
	"github.com/theopenlane/ent/entitlements/features"
	"github.com/theopenlane/utils/contextx"
)

// Typed context hint strings
type ModuleHint models.OrgModule
type PreferredProviderHint storagetypes.ProviderType
type KnownProviderHint storagetypes.ProviderType
type SizeBytesHint int64

// PopulateProviderHints ensures standard metadata is present on the file's provider hints
func PopulateProviderHints(file *pkgobjects.File, orgID string) {
	if file == nil {
		return
	}

	hints := file.ProviderHints
	if hints == nil {
		hints = &storage.ProviderHints{}
		file.ProviderHints = hints
	}

	if hints.Metadata == nil {
		hints.Metadata = map[string]string{}
	}

	if orgID != "" && hints.OrganizationID == "" {
		hints.OrganizationID = orgID
	}

	if file.FieldName != "" {
		hints.Metadata["key"] = file.FieldName
	}

	if file.CorrelatedObjectType != "" {
		hints.Metadata["object_type"] = file.CorrelatedObjectType
	}

	if size := file.Size; size > 0 {
		hints.Metadata["size_bytes"] = strconv.FormatInt(size, 10)
	}

	if module, ok := ResolveModuleFromFile(*file); ok {
		hints.Module = module
		hints.Metadata["module"] = string(module)
	}
}

// ResolveModuleFromFile attempts to determine the module associated with the upload
func ResolveModuleFromFile(f pkgobjects.File) (models.OrgModule, bool) {
	if module, ok := moduleFromHints(f.ProviderHints); ok {
		return module, true
	}

	if f.CorrelatedObjectType != "" {
		featureKey := lo.PascalCase(f.CorrelatedObjectType)
		if modules, ok := features.FeatureOfType[featureKey]; ok && len(modules) > 0 {
			return modules[0], true
		}
	}

	return "", false
}

// ApplyProviderHints injects hint values into the resolution context using typed context values
func ApplyProviderHints(ctx context.Context, hints *storagetypes.ProviderHints) context.Context {
	if hints == nil {
		return ctx
	}

	if module, ok := moduleFromHints(hints); ok {
		ctx = contextx.With(ctx, ModuleHint(module))
	}

	if hints.PreferredProvider != "" {
		ctx = contextx.With(ctx, PreferredProviderHint(hints.PreferredProvider))
	}

	if hints.KnownProvider != "" {
		ctx = contextx.With(ctx, KnownProviderHint(hints.KnownProvider))
	}

	if sizeStr, ok := hints.Metadata["size_bytes"]; ok {
		if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
			ctx = contextx.With(ctx, SizeBytesHint(size))
		}
	}

	return ctx
}

func moduleFromHints(hints *storagetypes.ProviderHints) (models.OrgModule, bool) {
	if hints == nil {
		return "", false
	}

	if hints.Module != nil {
		switch v := hints.Module.(type) {
		case models.OrgModule:
			if v != "" {
				return v, true
			}
		case string:
			if v != "" {
				return models.OrgModule(v), true
			}
		}
	}

	if hints.Metadata != nil {
		if module, ok := hints.Metadata["module"]; ok && module != "" {
			return models.OrgModule(module), true
		}
	}

	return "", false
}
