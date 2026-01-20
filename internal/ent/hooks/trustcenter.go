package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/privacy"

	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/pkg/logx"
)

// HookTrustCenter runs on trust center create mutations
func HookTrustCenter() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterFunc(func(ctx context.Context, m *generated.TrustCenterMutation) (generated.Value, error) {
			orgID, err := auth.GetOrganizationIDFromContext(ctx)
			if err != nil {
				return nil, err
			}

			exists, err := m.Client().TrustCenter.Query().
				Exist(ctx)
			if err != nil {
				return nil, err
			}

			if exists {
				return nil, ErrNotSingularTrustCenter
			}

			org, err := m.Client().Organization.Query().
				Where(organization.ID(orgID)).
				Select(organization.FieldName).
				Only(ctx)
			if err != nil {
				return nil, err
			}

			m.SetSlug(strcase.KebabCase(org.Name))

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			trustCenter, ok := retVal.(*generated.TrustCenter)
			if !ok {
				return retVal, nil
			}

			id := trustCenter.ID

			// create trust center settings automatically unless setting IDs were provided
			settingIDs := m.SettingIDs()
			previewSettingIDs := m.PreviewSettingIDs()

			if len(settingIDs) > 0 || len(previewSettingIDs) > 0 {
				for _, settingID := range append(settingIDs, previewSettingIDs...) {
					if settingID == "" {
						continue
					}

					if err := m.Client().TrustCenterSetting.UpdateOneID(settingID).
						SetTrustCenterID(id).
						Exec(privacy.DecisionContext(ctx, privacy.Allow)); err != nil {
						return nil, err
					}
				}
			}

			createLive, createPreview := false, false
			switch {
			case len(settingIDs) == 0 && len(previewSettingIDs) == 0:
				createLive, createPreview = true, true
			case len(settingIDs) > 1 || len(previewSettingIDs) > 1:
				logx.FromContext(ctx).Debug().Msg("trust center setting IDs provided, skipping default setting creation")
			case len(settingIDs) == 1 && len(previewSettingIDs) == 0:
				setting, err := m.Client().TrustCenterSetting.Get(ctx, settingIDs[0])
				if err != nil {
					return nil, err
				}

				switch setting.Environment {
				case enums.TrustCenterEnvironmentLive:
					createPreview = true
				case enums.TrustCenterEnvironmentPreview:
					createLive = true
				}
			case len(settingIDs) == 0 && len(previewSettingIDs) == 1:
				setting, err := m.Client().TrustCenterSetting.Get(ctx, previewSettingIDs[0])
				if err != nil {
					return nil, err
				}

				switch setting.Environment {
				case enums.TrustCenterEnvironmentLive:
					createPreview = true
				case enums.TrustCenterEnvironmentPreview:
					createLive = true
				}
			}

			if createLive {
				setting, err := m.Client().TrustCenterSetting.Create().
					SetTrustCenterID(id).
					SetTitle(fmt.Sprintf("%s Trust Center", org.Name)).
					SetOverview(defaultOverview).
					SetEnvironment(enums.TrustCenterEnvironmentLive).
					Save(privacy.DecisionContext(ctx, privacy.Allow))
				if err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to create live trust center setting")

					return nil, err
				}

				if err := m.Client().TrustCenter.UpdateOne(trustCenter).SetSettingID(setting.ID).Exec(ctx); err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to set live setting ID on trust center")

					return nil, err
				}

				trustCenter.Edges.Setting = setting
			}

			if createPreview {
				if trustCenter.PreviewDomainID != "" {
					// delete the old preview if it exists
					if err = enqueueJob(ctx, m.Job,
						jobspec.DeletePreviewDomainArgs{
							CustomDomainID:           trustCenter.PreviewDomainID,
							TrustCenterPreviewZoneID: trustCenterConfig.PreviewZoneID,
						}, nil,
					); err != nil {
						return nil, err
					}
				}

				// create preview settings with same values but environment set to "preview"
				previewSetting, err := m.Client().TrustCenterSetting.Create().
					SetTrustCenterID(id).
					SetTitle(fmt.Sprintf("%s Trust Center", org.Name)).
					SetOverview(defaultOverview).
					SetEnvironment(enums.TrustCenterEnvironmentPreview).
					Save(privacy.DecisionContext(ctx, privacy.Allow))
				if err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to create preview trust center setting")

					return nil, err
				}

				if err := m.Client().TrustCenter.UpdateOne(trustCenter).SetPreviewSettingID(previewSetting.ID).Exec(ctx); err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to set preview setting ID on trust center")

					return nil, err
				}

				trustCenter.Edges.PreviewSetting = previewSetting
			}

			// create watermark config for trust center with default values
			if id, ok := m.WatermarkConfigID(); ok && id != "" {
				// watermark config was provided, skip creation
				return trustCenter, nil
			}

			if ids := m.WatermarkConfigIDs(); len(ids) > 0 {
				// watermark config IDs were provided, skip creation
				return trustCenter, nil
			}

			input := generated.CreateTrustCenterWatermarkConfigInput{
				TrustCenterID:  &id,
				Text:           &defaultWatermarkText,
				OwnerID:        &orgID,
				TrustCenterIDs: []string{trustCenter.ID},
			}

			if err := m.Client().TrustCenterWatermarkConfig.Create().
				SetInput(input).
				Exec(privacy.DecisionContext(ctx, privacy.Allow)); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to create trust center watermark config")

				return nil, err
			}

			wildcardTuples := fgax.CreateWildcardViewerTuple(trustCenter.ID, "trust_center")

			if _, err := m.Authz.WriteTupleKeys(ctx, wildcardTuples, nil); err != nil {
				return nil, fmt.Errorf("failed to create file access permissions: %w", err)
			}

			if trustCenter.CustomDomainID != nil {
				if err = enqueueJob(ctx, m.Job, jobspec.CreatePirschDomainArgs{
					TrustCenterID: trustCenter.ID,
				}, nil); err != nil {
					return nil, err
				}
			}

			return trustCenter, nil
		})
	}, ent.OpCreate)
}

const (
	defaultOverview = `
# Welcome to your Trust Center

This is the default overview for your Trust Center. You can customize this by editing the Trust Center settings.
`
)

var (
	defaultWatermarkText = "Controlled Copy — Watermark Required"
)

// HookTrustCenterDelete runs on trust center delete mutations
func HookTrustCenterDelete() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterFunc(func(ctx context.Context, m *generated.TrustCenterMutation) (generated.Value, error) {
			// Only run on delete operations (including soft delete)
			if !isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			// Get the trust center ID from the mutation
			id, ok := m.ID()
			if !ok {
				// If we can't get the ID, just proceed with the deletion
				return next.Mutate(ctx, m)
			}

			// Query the trust center to get the pirsch_domain_id before deletion
			tc, err := m.Client().TrustCenter.Get(ctx, id)
			if err != nil {
				// If we can't find the trust center, just proceed
				return next.Mutate(ctx, m)
			}

			// If the trust center has a Pirsch domain ID, kick off the DeletePirschDomain job
			if tc.PirschDomainID != "" {
				err := enqueueJob(ctx, m.Job, jobspec.DeletePirschDomainArgs{
					PirschDomainID: tc.PirschDomainID,
				}, nil)
				if err != nil {
					return nil, err
				}
			}

			// Store the domain IDs before deletion
			previewDomainID := tc.PreviewDomainID

			// Execute the trust center deletion first
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// If preview domain is set, kick off the delete preview job
			if previewDomainID != "" {
				if err := enqueueJob(ctx, m.Job, jobspec.DeletePreviewDomainArgs{
					CustomDomainID:           previewDomainID,
					TrustCenterPreviewZoneID: trustCenterConfig.PreviewZoneID,
				}, nil); err != nil {
					return nil, err
				}
			}

			// Trigger cache refresh for the deleted trust center
			var customDomain string
			if tc.CustomDomainID != nil {
				if cd, err := m.Client().CustomDomain.Get(ctx, *tc.CustomDomainID); err == nil && cd.CnameRecord != "" {
					customDomain = cd.CnameRecord
				}
			}

			if targetURL := buildTrustCenterURL(customDomain, tc.Slug); targetURL != "" {
				if err := triggerCacheRefresh(ctx, targetURL); err != nil {
					return nil, err
				}
			}

			if tc.PreviewDomainID != "" {
				if cd, err := m.Client().CustomDomain.Get(ctx, tc.PreviewDomainID); err == nil && cd.CnameRecord != "" {
					if targetURL := buildTrustCenterURL(cd.CnameRecord, ""); targetURL != "" {
						if err := triggerCacheRefresh(ctx, targetURL); err != nil {
							return nil, err
						}
					}
				}
			}

			return retVal, nil
		})
	}
}

// HookTrustCenterUpdate runs on trust center update mutations
func HookTrustCenterUpdate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterFunc(func(ctx context.Context, m *generated.TrustCenterMutation) (generated.Value, error) {
			// Get the previous custom domain ID
			// If the custom domain ID has been updated from nothing/nil, then kick off the CreatePirschDomain job
			previousCustomDomainID, err := m.OldCustomDomainID(ctx)
			if err != nil {
				return nil, err
			}

			previousPirschDomainID, err := m.OldPirschDomainID(ctx)
			if err != nil {
				return nil, err
			}

			customDomainCleared := m.CustomDomainIDCleared()
			mutationCustomDomainID, mutationCustomDomainIDExists := m.CustomDomainID()

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			tcID, err := GetObjectIDFromEntValue(v)
			if err != nil {
				return v, err
			}

			if customDomainCleared || (mutationCustomDomainIDExists && mutationCustomDomainID == "") {
				if previousCustomDomainID != nil && previousPirschDomainID != "" {
					if err := enqueueJob(ctx, m.Job, jobspec.DeletePirschDomainArgs{
						PirschDomainID: previousPirschDomainID,
					}, nil); err != nil {
						return nil, err
					}
				}

				if previousCustomDomainID != nil {
					if cd, err := m.Client().CustomDomain.Get(ctx, *previousCustomDomainID); err == nil && cd.CnameRecord != "" {
						if targetURL := buildTrustCenterURL(cd.CnameRecord, ""); targetURL != "" {
							if err := triggerCacheRefresh(ctx, targetURL); err != nil {
								return nil, err
							}
						}
					}
				}

				return v, nil
			}

			if mutationCustomDomainIDExists && previousCustomDomainID == nil && mutationCustomDomainID != "" {
				if err := enqueueJob(ctx, m.Job, jobspec.CreatePirschDomainArgs{
					TrustCenterID: tcID,
				}, nil); err != nil {
					return nil, err
				}
			} else if mutationCustomDomainIDExists && previousCustomDomainID != nil && mutationCustomDomainID != "" && mutationCustomDomainID != *previousCustomDomainID {
				if err := enqueueJob(ctx, m.Job, jobspec.UpdatePirschDomainArgs{
					TrustCenterID: tcID,
				}, nil); err != nil {
					return nil, err
				}

				if cd, err := m.Client().CustomDomain.Get(ctx, *previousCustomDomainID); err == nil && cd.CnameRecord != "" {
					if targetURL := buildTrustCenterURL(cd.CnameRecord, ""); targetURL != "" {
						if err := triggerCacheRefresh(ctx, targetURL); err != nil {
							return nil, err
						}
					}
				}
			}

			return v, nil
		})
	}, ent.OpUpdateOne)
}
