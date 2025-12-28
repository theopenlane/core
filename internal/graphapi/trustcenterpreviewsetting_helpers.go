package graphapi

import (
	"context"
	"sort"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersetting"
	"github.com/theopenlane/core/internal/graphapi/model"
)

func resolveTrustCenterForPreviewSetting(ctx context.Context, client *generated.Client, trustCenterID string) (*generated.TrustCenter, error) {
	if trustCenterID != "" {
		return client.TrustCenter.Get(ctx, trustCenterID)
	}

	return client.TrustCenter.Query().Only(ctx)
}

func buildPreviewSettingCreateInput(trustCenterID string, input *model.CreateTrustCenterPreviewSettingInput) generated.CreateTrustCenterSettingInput {
	previewEnv := enums.TrustCenterEnvironmentPreview
	createInput := generated.CreateTrustCenterSettingInput{
		TrustCenterID: &trustCenterID,
		Environment:   &previewEnv,
	}

	if input == nil {
		return createInput
	}

	createInput.Title = input.Title
	createInput.Overview = input.Overview
	createInput.PrimaryColor = input.PrimaryColor
	createInput.LogoRemoteURL = input.LogoRemoteURL
	createInput.LogoFileID = input.LogoFileID
	createInput.FaviconRemoteURL = input.FaviconRemoteURL
	createInput.FaviconFileID = input.FaviconFileID
	createInput.ThemeMode = input.ThemeMode
	createInput.Font = input.Font
	createInput.ForegroundColor = input.ForegroundColor
	createInput.BackgroundColor = input.BackgroundColor
	createInput.AccentColor = input.AccentColor

	return createInput
}

func buildPreviewSettingUpdateInput(input model.CreateTrustCenterPreviewSettingInput) generated.UpdateTrustCenterSettingInput {
	return generated.UpdateTrustCenterSettingInput{
		Title:            input.Title,
		Overview:         input.Overview,
		PrimaryColor:     input.PrimaryColor,
		LogoRemoteURL:    input.LogoRemoteURL,
		LogoFileID:       input.LogoFileID,
		FaviconRemoteURL: input.FaviconRemoteURL,
		FaviconFileID:    input.FaviconFileID,
		ThemeMode:        input.ThemeMode,
		Font:             input.Font,
		ForegroundColor:  input.ForegroundColor,
		BackgroundColor:  input.BackgroundColor,
		AccentColor:      input.AccentColor,
	}
}

func upsertPreviewSetting(ctx context.Context, client *generated.Client, trustCenter *generated.TrustCenter, createInput *model.CreateTrustCenterPreviewSettingInput, updateInput *generated.UpdateTrustCenterSettingInput) (*generated.TrustCenterSetting, error) {
	previewSettings, err := client.TrustCenterSetting.Query().
		Where(
			trustcentersetting.TrustCenterIDEQ(trustCenter.ID),
			trustcentersetting.EnvironmentEQ(enums.TrustCenterEnvironmentPreview),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	var previewSetting *generated.TrustCenterSetting
	if len(previewSettings) > 0 {
		sort.Slice(previewSettings, func(i, j int) bool {
			return previewSettings[i].UpdatedAt.After(previewSettings[j].UpdatedAt)
		})
		previewSetting = previewSettings[0]
	}

	if len(previewSettings) > 1 {
		for _, extra := range previewSettings[1:] {
			if err := client.TrustCenterSetting.DeleteOneID(extra.ID).Exec(ctx); err != nil {
				return nil, err
			}
		}
	}

	if previewSetting == nil {
		create := buildPreviewSettingCreateInput(trustCenter.ID, createInput)
		previewSetting, err = client.TrustCenterSetting.Create().SetInput(create).Save(ctx)
		if err != nil {
			return nil, err
		}

		if updateInput != nil {
			previewSetting, err = previewSetting.Update().SetInput(*updateInput).Save(ctx)
			if err != nil {
				return nil, err
			}
		}
	} else {
		if updateInput != nil {
			previewSetting, err = previewSetting.Update().SetInput(*updateInput).Save(ctx)
			if err != nil {
				return nil, err
			}
		} else if createInput != nil {
			update := buildPreviewSettingUpdateInput(*createInput)
			previewSetting, err = previewSetting.Update().SetInput(update).Save(ctx)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := client.TrustCenter.UpdateOneID(trustCenter.ID).
		SetPreviewSettingID(previewSetting.ID).
		Exec(ctx); err != nil {
		return nil, err
	}

	return previewSetting, nil
}
