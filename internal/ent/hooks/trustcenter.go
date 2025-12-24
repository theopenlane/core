package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/privacy"

	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/core/pkg/enums"
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

			// create trust center settings automatically unless setting IDs were provided
			settingIDs := m.SettingIDs()

			createLive, createPreview := false, false
			switch len(settingIDs) {
			case 0:
				createLive, createPreview = true, true
			case 1:
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

			default:
				logx.FromContext(ctx).Debug().Msg("trust center setting IDs provided, skipping default setting creation")

				return retVal, nil
			}

			trustCenter := retVal.(*generated.TrustCenter)

			// If settings were not created, create default settings
			id, err := GetObjectIDFromEntValue(retVal)
			if err != nil {
				return retVal, err
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
					if _, err = m.Job.Insert(
						ctx,
						corejobs.DeletePreviewDomainArgs{
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

			// Create system tuple for system admin access
			systemTuple := fgax.GetTupleKey(fgax.TupleRequest{
				SubjectID:   "openlane_core",
				SubjectType: "system",
				ObjectID:    trustCenter.ID,
				ObjectType:  "trust_center",
				Relation:    "system",
			})

			if _, err := m.Authz.WriteTupleKeys(ctx, append(wildcardTuples, systemTuple), nil); err != nil {
				return nil, fmt.Errorf("failed to create file access permissions: %w", err)
			}

			if trustCenter.CustomDomainID != nil {
				if _, err = m.Job.Insert(ctx, corejobs.CreatePirschDomainArgs{
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

This is the default overview for your trust center. You can customize this by editing the trust center settings.
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
				_, err := m.Job.Insert(ctx, corejobs.DeletePirschDomainArgs{
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
				if _, err := m.Job.Insert(ctx, corejobs.DeletePreviewDomainArgs{
					CustomDomainID:           previewDomainID,
					TrustCenterPreviewZoneID: trustCenterConfig.PreviewZoneID,
				}, nil); err != nil {
					return nil, err
				}
			}

			return retVal, nil
		})
	}
}

func HookTrustCenterUpdate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterFunc(func(ctx context.Context, m *generated.TrustCenterMutation) (generated.Value, error) {
			// Get the previous custom domain ID
			// If the custom domain ID has been updated from nothing/nil, then kick off the CreatePirschDomain job
			previousCustomDomainID, err := m.OldCustomDomainID(ctx)
			if err != nil {
				return nil, err
			}

			mutationCustomDomainID, mutationCustomDomainIDExists := m.CustomDomainID()

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			tcID, err := GetObjectIDFromEntValue(v)
			if err != nil {
				return v, err
			}

			if mutationCustomDomainIDExists && previousCustomDomainID == nil && mutationCustomDomainID != "" {
				if _, err := m.Job.Insert(ctx, corejobs.CreatePirschDomainArgs{
					TrustCenterID: tcID,
				}, nil); err != nil {
					return nil, err
				}
			} else if mutationCustomDomainIDExists && previousCustomDomainID != nil && mutationCustomDomainID != *previousCustomDomainID {
				if _, err := m.Job.Insert(ctx, corejobs.UpdatePirschDomainArgs{
					TrustCenterID: tcID,
				}, nil); err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpUpdateOne)
}
