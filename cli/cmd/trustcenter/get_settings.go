//go:build cli

package trustcenter

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/theopenlane/cli/cmd"
)

var getSettingsCmd = &cobra.Command{
	Use:   "get-settings",
	Short: "get trust center settings",
	Run: func(cmd *cobra.Command, args []string) {
		err := getSettings(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getSettingsCmd)

	getSettingsCmd.Flags().StringP("id", "i", "", "trust center setting id to get")
	getSettingsCmd.Flags().StringP("trust-center-id", "c", "", "trust center id to get settings for (alternative to setting id)")
}

// getSettingsValidation validates the required fields for the command
func getSettingsValidation() (id string, trustCenterID string, err error) {
	id = cmd.Config.String("id")
	trustCenterID = cmd.Config.String("trust-center-id")

	if id == "" && trustCenterID == "" {
		return id, trustCenterID, cmd.NewRequiredFieldMissingError("id or trust-center-id")
	}

	return id, trustCenterID, nil
}

// get trust center settings
func getSettings(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, trustCenterID, err := getSettingsValidation()
	cobra.CheckErr(err)

	// If we have a trust center ID instead of setting ID, get the trust center and extract settings
	if id == "" && trustCenterID != "" {
		trustCenter, err := client.GetTrustCenterByID(ctx, trustCenterID)
		cobra.CheckErr(err)

		if trustCenter.TrustCenter.Setting == nil {
			return cmd.NewRequiredFieldMissingError("trust center has no settings")
		}

		// Convert the setting to the expected format for output
		setting := &struct {
			TrustCenterSetting *struct {
				ID            string  `json:"id"`
				TrustCenterID *string `json:"trustCenterID,omitempty"`
				Title         *string `json:"title,omitempty"`
				Overview      *string `json:"overview,omitempty"`
				PrimaryColor  *string `json:"primaryColor,omitempty"`
				CreatedAt     *string `json:"createdAt,omitempty"`
				UpdatedAt     *string `json:"updatedAt,omitempty"`
			} `json:"trustCenterSetting"`
		}{
			TrustCenterSetting: &struct {
				ID            string  `json:"id"`
				TrustCenterID *string `json:"trustCenterID,omitempty"`
				Title         *string `json:"title,omitempty"`
				Overview      *string `json:"overview,omitempty"`
				PrimaryColor  *string `json:"primaryColor,omitempty"`
				CreatedAt     *string `json:"createdAt,omitempty"`
				UpdatedAt     *string `json:"updatedAt,omitempty"`
			}{
				ID:            trustCenter.TrustCenter.Setting.ID,
				TrustCenterID: &trustCenterID,
				Title:         trustCenter.TrustCenter.Setting.Title,
				Overview:      trustCenter.TrustCenter.Setting.Overview,
				PrimaryColor:  trustCenter.TrustCenter.Setting.PrimaryColor,
			},
		}

		return consoleSettingsOutput(setting)
	}

	// Get by setting ID
	o, err := client.GetTrustCenterSettingByID(ctx, id)
	cobra.CheckErr(err)

	return consoleSettingsOutput(o)
}
