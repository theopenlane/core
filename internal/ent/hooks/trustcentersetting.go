package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organization"
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
	PreviewZoneID            string
	CnameTarget              string
	DefaultTrustCenterDomain string
	CacheRefreshScheme       string
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

			if err := setDefaultCompanyName(ctx, m); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// setDefaultCompanyName will add the company name either from the org display name or if provided in the mutation
func setDefaultCompanyName(ctx context.Context, m *generated.TrustCenterSettingMutation) error {

	if m.Op().Is(ent.OpUpdateOne) {
		oldName, err := m.OldCompanyName(ctx)
		if err == nil && oldName != "" {
			return nil
		}
	}

	name, ok := m.CompanyName()
	if ok && name != "" {
		return nil
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	org, err := m.Client().Organization.
		Query().Select(organization.FieldDisplayName).
		Where(organization.ID(orgID)).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).
			Str("owner_id", orgID).
			Msg("failed to get organization for company name default")
		return err
	}

	m.SetCompanyName(org.DisplayName)
	return nil
}

// HookTrustCenterSettingCreatePreview is a hook that runs on trust center setting create or update
// to enqueue a job to create the preview domain
func HookTrustCenterSettingCreatePreview() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterSettingFunc(func(ctx context.Context, m *generated.TrustCenterSettingMutation) (generated.Value, error) {
			v, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// check the environment
			if !isPreviewSetting(ctx, m) {
				logx.FromContext(ctx).Debug().Msg("trust center setting is not for preview environment, skipping preview domain creation job")

				return v, nil
			}

			trustCenterID, hasTc := getTrustCenterSettingID(ctx, m)
			if !hasTc {
				// this should not happen, but just in case
				logx.FromContext(ctx).Warn().Msg("trust center ID missing in trust center setting mutation, skipping preview domain creation job")

				return v, nil
			}

			trustCenter, err := m.Client().TrustCenter.Get(ctx, trustCenterID)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", trustCenterID).Msg("failed to get trust center for trust center setting mutation")

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
				logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", trustCenterID).Msg("failed to enqueue create preview domain job")

				return nil, err
			}

			return v, nil
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// isPreviewSetting checks if the trust center setting mutation is for the preview environment
func isPreviewSetting(ctx context.Context, m *generated.TrustCenterSettingMutation) bool {
	env, ok := m.Environment()
	if ok {
		return env == enums.TrustCenterEnvironmentPreview
	}

	if m.Op().Is(ent.OpCreate) {
		// on create there is no old environment to check
		return false
	}

	logx.FromContext(ctx).Debug().Msg("environment missing in trust center setting mutation")

	// check old environment for updates
	oldEnv, err := m.OldEnvironment(ctx)
	if err != nil {
		return false
	}

	return oldEnv == enums.TrustCenterEnvironmentPreview
}

// getTrustCenterSettingID retrieves the trust center ID from the mutation, handling both create and update cases
// by checking the old value if necessary on updates if its not included in the mutation
func getTrustCenterSettingID(ctx context.Context, m *generated.TrustCenterSettingMutation) (string, bool) {
	trustCenterID, hasTc := m.TrustCenterID()
	if hasTc || m.Op().Is(ent.OpCreate) {
		return trustCenterID, hasTc
	}

	// on update we need to get the old trust center ID
	oldTrustCenterID, err := m.OldTrustCenterID(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get old trust center ID from trust center setting mutation")

		return "", false
	}

	return oldTrustCenterID, oldTrustCenterID != ""
}

// checkTrustCenterFiles checks for logo and favicon files in the context
// and processes them for the trust center setting mutation
func checkTrustCenterFiles(ctx context.Context, m *generated.TrustCenterSettingMutation) (context.Context, error) {
	logoKey := "logoFile"
	faviconKey := "faviconFile"

	var err error

	ctx, err = processSingleMutationFile(ctx, m, logoKey, "trust_center_setting", ErrTooManyLogoFiles,
		func(mut *generated.TrustCenterSettingMutation, id string) { mut.SetLogoLocalFileID(id) },
	)
	if err != nil {
		return ctx, err
	}

	ctx, err = processSingleMutationFile(ctx, m, faviconKey, "trust_center_setting", ErrTooManyFaviconFiles,
		func(mut *generated.TrustCenterSettingMutation, id string) { mut.SetFaviconLocalFileID(id) },
	)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}
