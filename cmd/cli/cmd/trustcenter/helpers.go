//go:build cli

package trustcenter

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func findTrustCenterCommand() *cobra.Command {
	for _, c := range cmdpkg.RootCmd.Commands() {
		if c.Use == "trustcenter" {
			return c
		}
	}

	return nil
}

func attachTrustCenterExtras(parent *cobra.Command) {
	parent.AddCommand(newGetSettingsCmd())
	parent.AddCommand(newUpdateSettingsCmd())
}

func consoleSettingsOutput(out any) error {
	if strings.EqualFold(cmdpkg.OutputFormat, cmdpkg.JSONOutput) {
		return speccli.PrintJSON(out)
	}

	switch v := out.(type) {
	case *openlaneclient.UpdateTrustCenterSetting:
		tableSettingsOutput(v.UpdateTrustCenterSetting.TrustCenterSetting)
	case *openlaneclient.GetTrustCenterSettingByID:
		tableSettingsOutputFromGeneric(openlaneclient.GetTrustCenterSettingByID_TrustCenterSetting{
			ID:            v.TrustCenterSetting.ID,
			TrustCenterID: v.TrustCenterSetting.TrustCenterID,
			Title:         v.TrustCenterSetting.Title,
			Overview:      v.TrustCenterSetting.Overview,
			PrimaryColor:  v.TrustCenterSetting.PrimaryColor,
			CreatedAt:     v.TrustCenterSetting.CreatedAt,
			UpdatedAt:     v.TrustCenterSetting.UpdatedAt,
		})
	default:
		payload, err := json.Marshal(out)
		if err != nil {
			return err
		}

		var setting openlaneclient.UpdateTrustCenterSetting_UpdateTrustCenterSetting_TrustCenterSetting
		if err := json.Unmarshal(payload, &setting); err != nil {
			return err
		}

		tableSettingsOutput(setting)
	}

	return nil
}

func tableSettingsOutput(setting openlaneclient.UpdateTrustCenterSetting_UpdateTrustCenterSetting_TrustCenterSetting) {
	writer := tables.NewTableWriter(rootCmd.OutOrStdout(), "ID", "TrustCenterID", "Title", "Overview", "PrimaryColor", "ThemeMode", "Font", "ForegroundColor", "BackgroundColor", "AccentColor", "CreatedAt", "UpdatedAt")

	title := derefString(setting.Title)
	overview := truncate(derefString(setting.Overview), 50)
	primaryColor := derefString(setting.PrimaryColor)

	themeMode := ""
	if setting.ThemeMode != nil {
		themeMode = setting.ThemeMode.String()
	}

	font := truncate(derefString(setting.Font), 20)
	foregroundColor := derefString(setting.ForegroundColor)
	backgroundColor := derefString(setting.BackgroundColor)
	accentColor := derefString(setting.AccentColor)
	trustCenterID := derefString(setting.TrustCenterID)
	createdAt := formatTime(setting.CreatedAt)
	updatedAt := formatTime(setting.UpdatedAt)

	writer.AddRow(setting.ID, trustCenterID, title, overview, primaryColor, themeMode, font, foregroundColor, backgroundColor, accentColor, createdAt, updatedAt)
	writer.Render()
}

func tableSettingsOutputFromGeneric(setting openlaneclient.GetTrustCenterSettingByID_TrustCenterSetting) {
	writer := tables.NewTableWriter(rootCmd.OutOrStdout(), "ID", "TrustCenterID", "Title", "Overview", "PrimaryColor", "CreatedAt", "UpdatedAt")

	title := derefString(setting.Title)
	overview := truncate(derefString(setting.Overview), 50)
	primaryColor := derefString(setting.PrimaryColor)
	trustCenterID := derefString(setting.TrustCenterID)
	createdAt := formatTime(setting.CreatedAt)
	updatedAt := formatTime(setting.UpdatedAt)

	writer.AddRow(setting.ID, trustCenterID, title, overview, primaryColor, createdAt, updatedAt)
	writer.Render()
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func formatTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}
