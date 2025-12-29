package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/objects"
)

var trustCenterConfig TrustCenterConfig

// SetTrustCenterConfig sets the trust center configuration
func SetTrustCenterConfig(cfg TrustCenterConfig) {
	trustCenterConfig = cfg
}

// TrustCenterConfig holds the trust center configuration
type TrustCenterConfig struct {
	PreviewZoneID string
	CnameTarget   string
}

// HookTrustCenterSetting process files for trust center settings
func HookTrustCenterSetting() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterSettingFunc(func(ctx context.Context, m *generated.TrustCenterSettingMutation) (generated.Value, error) {
			logx.FromContext(ctx).Debug().Msg("trust center setting hook")

			// check for uploaded files (e.g. logo image)
			fileIDs := objects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkTrustCenterFiles(ctx, m)
				if err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// HookTrustCenterSettingCreatePreview is a hook that runs on trust center setting create
// to enqueue a job to create the preview domain
func HookTrustCenterSettingCreatePreview() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterSettingFunc(func(ctx context.Context, m *generated.TrustCenterSettingMutation) (generated.Value, error) {
			logx.FromContext(ctx).Debug().Msg("trust center setting create preview hook")
			v, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// check the environment
			if env, ok := m.Environment(); !ok || env != enums.TrustCenterEnvironmentPreview {
				return v, nil
			}

			trustCenterID, hasTc := m.TrustCenterID()
			if !hasTc {
				// should never happen
				return nil, ErrMissingTrustCenterID
			}

			trustCenter, err := m.Client().TrustCenter.Get(ctx, trustCenterID)
			if err != nil {
				return nil, err
			}

			if trustCenter.PreviewDomainID != "" {
				logx.FromContext(ctx).Debug().Str("trust_center_id", trustCenterID).Msg("preview domain already exists, skipping creation job")
				return v, nil
			}

			// Insert job to create preview domain with config values
			if err = enqueueJob(ctx, m.Job, jobspec.CreatePreviewDomainArgs{
				TrustCenterID:            trustCenterID,
				TrustCenterPreviewZoneID: trustCenterConfig.PreviewZoneID,
				TrustCenterCnameTarget:   trustCenterConfig.CnameTarget,
			}, nil); err != nil {
				return nil, err
			}

			return v, nil

		})
	}, ent.OpCreate)
}

// checkTrustCenterFiles checks for logo and favicon files in the context
// and processes them for the trust center setting mutation
func checkTrustCenterFiles(ctx context.Context, m *generated.TrustCenterSettingMutation) (context.Context, error) {
	logoKey := "logoFile"
	faviconKey := "faviconFile"

	var err error

	ctx, err = processSingleMutationFile(ctx, m, logoKey, "trust_center_setting", ErrTooManyLogoFiles,
		func(mut *generated.TrustCenterSettingMutation, id string) { mut.SetLogoLocalFileID(id) },
		func(mut *generated.TrustCenterSettingMutation) (string, bool) { return mut.ID() },
		func(mut *generated.TrustCenterSettingMutation) string { return mut.Type() },
	)
	if err != nil {
		return ctx, err
	}

	ctx, err = processSingleMutationFile(ctx, m, faviconKey, "trust_center_setting", ErrTooManyFaviconFiles,
		func(mut *generated.TrustCenterSettingMutation, id string) { mut.SetFaviconLocalFileID(id) },
		func(mut *generated.TrustCenterSettingMutation) (string, bool) { return mut.ID() },
		func(mut *generated.TrustCenterSettingMutation) string { return mut.Type() },
	)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}
