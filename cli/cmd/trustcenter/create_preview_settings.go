//go:build cli

package trustcenter

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var createPreviewSettingsCmd = &cobra.Command{
	Use:   "create-preview-settings",
	Short: "create preview environment trust center settings",
	Run: func(cmd *cobra.Command, args []string) {
		err := createPreviewSettings(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createPreviewSettingsCmd)

	// command line flags for the create preview settings command
	createPreviewSettingsCmd.Flags().StringP("trust-center-id", "c", "", "trust center id to create preview settings for (required)")
	createPreviewSettingsCmd.Flags().StringP("title", "t", "", "title of the trust center")
	createPreviewSettingsCmd.Flags().StringP("overview", "o", "", "overview of the trust center")
	createPreviewSettingsCmd.Flags().StringP("primary-color", "p", "", "primary color for the trust center (hex color)")

	// theme and styling options
	createPreviewSettingsCmd.Flags().StringP("theme-mode", "", "", "theme mode for the trust center (EASY or ADVANCED)")
	createPreviewSettingsCmd.Flags().StringP("font", "", "", "font for the trust center")
	createPreviewSettingsCmd.Flags().StringP("foreground-color", "", "", "foreground color for the trust center (hex color)")
	createPreviewSettingsCmd.Flags().StringP("background-color", "", "", "background color for the trust center (hex color)")
	createPreviewSettingsCmd.Flags().StringP("accent-color", "", "", "accent color for the trust center (hex color)")
}

// createPreviewSettingsValidation validates the required fields for the command
func createPreviewSettingsValidation() (input graphclient.CreateTrustCenterPreviewSettingInput, err error) {
	trustCenterID := cmd.Config.String("trust-center-id")
	if trustCenterID == "" {
		return input, cmd.NewRequiredFieldMissingError("trust-center-id")
	}

	input.TrustCenterID = trustCenterID

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

	// Theme and styling options
	themeModeStr := cmd.Config.String("theme-mode")
	if themeModeStr != "" {
		input.ThemeMode = enums.ToTrustCenterThemeMode(themeModeStr)
	}

	font := cmd.Config.String("font")
	if font != "" {
		input.Font = &font
	}

	foregroundColor := cmd.Config.String("foreground-color")
	if foregroundColor != "" {
		input.ForegroundColor = &foregroundColor
	}

	backgroundColor := cmd.Config.String("background-color")
	if backgroundColor != "" {
		input.BackgroundColor = &backgroundColor
	}

	accentColor := cmd.Config.String("accent-color")
	if accentColor != "" {
		input.AccentColor = &accentColor
	}

	return input, nil
}

// create preview trust center settings
func createPreviewSettings(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := createPreviewSettingsValidation()
	cobra.CheckErr(err)

	o, err := client.CreateTrustCenterPreviewSetting(ctx, input)
	cobra.CheckErr(err)

	return consolePreviewSettingsOutput(o)
}

// consolePreviewSettingsOutput prints the preview trust center settings output in the console
func consolePreviewSettingsOutput(e *graphclient.CreateTrustCenterPreviewSetting) error {
	return consoleSettingsOutput(e)
}
