package trustcenter

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var updateSettingsCmd = &cobra.Command{
	Use:   "update-settings",
	Short: "update trust center settings",
	Run: func(cmd *cobra.Command, args []string) {
		err := updateSettings(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateSettingsCmd)

	updateSettingsCmd.Flags().StringP("id", "i", "", "trust center setting id to update")
	updateSettingsCmd.Flags().StringP("trust-center-id", "c", "", "trust center id to find settings for (alternative to setting id)")

	// command line flags for the update settings command
	updateSettingsCmd.Flags().StringP("title", "t", "", "title of the trust center")
	updateSettingsCmd.Flags().StringP("overview", "o", "", "overview of the trust center")
	updateSettingsCmd.Flags().StringP("primary-color", "p", "", "primary color for the trust center (hex color)")
	updateSettingsCmd.Flags().StringP("logo-file", "l", "", "local of logo file to upload")
}

// updateSettingsValidation validates the required fields for the command
func updateSettingsValidation() (id string, input openlaneclient.UpdateTrustCenterSettingInput, logoFile *graphql.Upload, err error) {
	id = cmd.Config.String("id")
	trustCenterID := cmd.Config.String("trust-center-id")

	if id == "" && trustCenterID == "" {
		return id, input, nil, cmd.NewRequiredFieldMissingError("id or trust-center-id")
	}

	// Build the input based on flags
	title := cmd.Config.String("title")
	if title != "" {
		input.Title = &title
	}

	overview := cmd.Config.String("overview")
	if overview != "" {
		input.Overview = &overview
	}

	primaryColor := cmd.Config.String("primary-color")
	if primaryColor != "" {
		input.PrimaryColor = &primaryColor
	}
	logoFileLoc := cmd.Config.String("logo-file")
	if logoFileLoc != "" {
		file, err := objects.NewUploadFile(logoFileLoc)
		if err != nil {
			return id, input, nil, err
		}

		logoFile = &graphql.Upload{
			File:        file.File,
			Filename:    file.Filename,
			Size:        file.Size,
			ContentType: file.ContentType,
		}
	}

	return id, input, logoFile, nil
}

// findSettingIDByTrustCenter finds the setting ID for a given trust center ID
func findSettingIDByTrustCenter(ctx context.Context, client *openlaneclient.OpenlaneClient, trustCenterID string) (string, error) {
	// Get the trust center to find its setting
	trustCenter, err := client.GetTrustCenterByID(ctx, trustCenterID)
	if err != nil {
		return "", err
	}

	if trustCenter.TrustCenter.Setting == nil {
		return "", cmd.NewRequiredFieldMissingError("trust center has no settings")
	}

	return trustCenter.TrustCenter.Setting.ID, nil
}

// update trust center settings
func updateSettings(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, logoFile, err := updateSettingsValidation()
	cobra.CheckErr(err)

	// If we have a trust center ID instead of setting ID, find the setting ID
	if id == "" {
		trustCenterID := cmd.Config.String("trust-center-id")
		id, err = findSettingIDByTrustCenter(ctx, client, trustCenterID)
		cobra.CheckErr(err)
	}

	o, err := client.UpdateTrustCenterSetting(ctx, id, input, logoFile)
	cobra.CheckErr(err)

	return consoleSettingsOutput(o)
}

// consoleSettingsOutput prints the trust center settings output in the console
func consoleSettingsOutput(e any) error {
	// check if the output format is JSON and print the settings in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the settings and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.UpdateTrustCenterSetting:
		e = v.UpdateTrustCenterSetting.TrustCenterSetting
	case *openlaneclient.GetTrustCenterSettingByID:
		e = v.TrustCenterSetting
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var setting openlaneclient.UpdateTrustCenterSetting_UpdateTrustCenterSetting_TrustCenterSetting
	err = json.Unmarshal(s, &setting)
	cobra.CheckErr(err)

	tableSettingsOutput(setting)

	return nil
}

// tableSettingsOutput prints the trust center settings in a table format
func tableSettingsOutput(setting openlaneclient.UpdateTrustCenterSetting_UpdateTrustCenterSetting_TrustCenterSetting) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "TrustCenterID", "Title", "Overview", "PrimaryColor", "CreatedAt", "UpdatedAt")

	title := ""
	if setting.Title != nil {
		title = *setting.Title
	}

	overview := ""
	if setting.Overview != nil {
		overview = *setting.Overview
	}

	primaryColor := ""
	if setting.PrimaryColor != nil {
		primaryColor = *setting.PrimaryColor
	}

	trustCenterID := ""
	if setting.TrustCenterID != nil {
		trustCenterID = *setting.TrustCenterID
	}

	createdAt := ""
	if setting.CreatedAt != nil {
		createdAt = setting.CreatedAt.Format("2006-01-02 15:04:05")
	}

	updatedAt := ""
	if setting.UpdatedAt != nil {
		updatedAt = setting.UpdatedAt.Format("2006-01-02 15:04:05")
	}

	// Truncate overview if it's too long for table display
	if len(overview) > 50 {
		overview = overview[:47] + "..."
	}

	writer.AddRow(setting.ID, trustCenterID, title, overview, primaryColor, createdAt, updatedAt)

	writer.Render()
}
