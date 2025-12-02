package hooks

import (
	"context"
	"errors"

	"entgo.io/ent"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/jobspec"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
)

var ErrTooManyFaviconFiles = errors.New("too many favicon files uploaded, only one is allowed")
var ErrMissingTrustCenterID = errors.New("trust center id is required")

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

			// Insert job to create preview domain with config values
			if _, err = m.Job.Insert(ctx, jobspec.CreatePreviewDomainArgs{
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

func checkTrustCenterFiles(ctx context.Context, m *generated.TrustCenterSettingMutation) (context.Context, error) {
	logoKey := "logoFile"
	faviconKey := "faviconFile"

	logoFiles, _ := objects.FilesFromContextWithKey(ctx, logoKey)
	if len(logoFiles) > 1 {
		return ctx, ErrTooManyLogoFiles
	}
	if len(logoFiles) == 1 {
		m.SetLogoLocalFileID(logoFiles[0].ID)

		adapter := objects.NewGenericMutationAdapter(m,
			func(mut *generated.TrustCenterSettingMutation) (string, bool) { return mut.ID() },
			func(mut *generated.TrustCenterSettingMutation) string { return mut.Type() },
		)

		ctx, _ = objects.ProcessFilesForMutation(ctx, adapter, logoKey, "trust_center_setting")
	}

	faviconFiles, _ := objects.FilesFromContextWithKey(ctx, faviconKey)
	if len(faviconFiles) > 1 {
		return ctx, ErrTooManyFaviconFiles
	}
	if len(faviconFiles) == 1 {
		m.SetFaviconLocalFileID(faviconFiles[0].ID)

		adapter := objects.NewGenericMutationAdapter(m,
			func(mut *generated.TrustCenterSettingMutation) (string, bool) { return mut.ID() },
			func(mut *generated.TrustCenterSettingMutation) string { return mut.Type() },
		)

		ctx, _ = objects.ProcessFilesForMutation(ctx, adapter, faviconKey, "trust_center_setting")
	}

	return ctx, nil
}
