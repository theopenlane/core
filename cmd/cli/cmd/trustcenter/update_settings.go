//go:build cli

package trustcenter

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func newUpdateSettingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-settings",
		Short: "update trust center settings",
		RunE: func(c *cobra.Command, _ []string) error {
			return updateSettings(c.Context())
		},
	}

	cmd.Flags().StringP("id", "i", "", "trust center setting id to update")
	cmd.Flags().StringP("trust-center-id", "c", "", "trust center id to find settings for (alternative to setting id)")

	cmd.Flags().StringP("title", "t", "", "title of the trust center")
	cmd.Flags().StringP("overview", "o", "", "overview of the trust center")
	cmd.Flags().StringP("primary-color", "p", "", "primary color for the trust center (hex color)")
	cmd.Flags().StringP("logo-file", "l", "", "local of logo file to upload")
	cmd.Flags().StringP("favicon-file", "f", "", "local of favicon file to upload")
	cmd.Flags().StringP("theme-mode", "", "", "theme mode for the trust center (EASY or ADVANCED)")
	cmd.Flags().StringP("font", "", "", "font for the trust center")
	cmd.Flags().StringP("foreground-color", "", "", "foreground color for the trust center (hex color)")
	cmd.Flags().StringP("background-color", "", "", "background color for the trust center (hex color)")
	cmd.Flags().StringP("accent-color", "", "", "accent color for the trust center (hex color)")

	return cmd
}

func updateSettingsValidation() (id string, input openlaneclient.UpdateTrustCenterSettingInput, logoFile *graphql.Upload, faviconFile *graphql.Upload, err error) {
	id = cmd.Config.String("id")
	trustCenterID := cmd.Config.String("trust-center-id")

	if id == "" && trustCenterID == "" {
		return id, input, nil, nil, cmd.NewRequiredFieldMissingError("id or trust-center-id")
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

	// Theme and styling options
	themeMode := cmd.Config.String("theme-mode")
	if themeMode != "" {
		// Validate theme mode
		themeModeEnum := enums.ToTrustCenterThemeMode(themeMode)
		if *themeModeEnum == enums.TrustCenterThemeModeInvalid {
			return id, input, nil, nil, cmd.NewRequiredFieldMissingError("invalid theme-mode, must be EASY or ADVANCED")
		}
		input.ThemeMode = themeModeEnum
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

	logoFileLoc := cmd.Config.String("logo-file")
	if logoFileLoc != "" {
		file, err := storage.NewUploadFile(logoFileLoc)
		if err != nil {
			return id, input, nil, nil, err
		}

		logoFile = &graphql.Upload{
			File:        file.RawFile,
			Filename:    file.OriginalName,
			Size:        file.Size,
			ContentType: file.ContentType,
		}
	}
	faviconFileLoc := cmd.Config.String("favicon-file")
	if faviconFileLoc != "" {
		file, err := storage.NewUploadFile(faviconFileLoc)
		if err != nil {
			return id, input, nil, nil, err
		}

		faviconFile = &graphql.Upload{
			File:        file.RawFile,
			Filename:    file.OriginalName,
			Size:        file.Size,
			ContentType: file.ContentType,
		}
	}

	return id, input, logoFile, faviconFile, nil
}

func findSettingIDByTrustCenter(ctx context.Context, client *openlaneclient.OpenlaneClient, trustCenterID string) (string, error) {
	trustCenter, err := client.GetTrustCenterByID(ctx, trustCenterID)
	if err != nil {
		return "", err
	}

	if trustCenter.TrustCenter.Setting == nil {
		return "", cmd.NewRequiredFieldMissingError("trust center has no settings")
	}

	return trustCenter.TrustCenter.Setting.ID, nil
}

func updateSettings(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		if err != nil {
			return err
		}
		defer cmd.StoreSessionCookies(client)
	}

	id, input, logoFile, faviconFile, err := updateSettingsValidation()
	if err != nil {
		return err
	}

	if id == "" {
		trustCenterID := cmd.Config.String("trust-center-id")
		id, err = findSettingIDByTrustCenter(ctx, client, trustCenterID)
		if err != nil {
			return err
		}
	}

	o, err := client.UpdateTrustCenterSetting(ctx, id, input, logoFile, faviconFile)
	if err != nil {
		return err
	}

	return consoleSettingsOutput(o)
}
