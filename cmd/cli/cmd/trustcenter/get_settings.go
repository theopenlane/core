//go:build cli

package trustcenter

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
)

func newGetSettingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-settings",
		Short: "get trust center settings",
		RunE: func(c *cobra.Command, _ []string) error {
			return getSettings(c.Context())
		},
	}

	cmd.Flags().StringP("id", "i", "", "trust center setting id to get")
	cmd.Flags().StringP("trust-center-id", "c", "", "trust center id to get settings for (alternative to setting id)")

	return cmd
}

func getSettingsValidation() (id string, trustCenterID string, err error) {
	id = cmd.Config.String("id")
	trustCenterID = cmd.Config.String("trust-center-id")

	if id == "" && trustCenterID == "" {
		return id, trustCenterID, cmd.NewRequiredFieldMissingError("id or trust-center-id")
	}

	return id, trustCenterID, nil
}

func getSettings(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		if err != nil {
			return err
		}
		defer cmd.StoreSessionCookies(client)
	}

	id, trustCenterID, err := getSettingsValidation()
	if err != nil {
		return err
	}

	if id == "" && trustCenterID != "" {
		trustCenter, err := client.GetTrustCenterByID(ctx, trustCenterID)
		if err != nil {
			return err
		}

		if trustCenter.TrustCenter.Setting == nil {
			return cmd.NewRequiredFieldMissingError("trust center has no settings")
		}

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

	o, err := client.GetTrustCenterSettingByID(ctx, id)
	if err != nil {
		return err
	}

	return consoleSettingsOutput(o)
}
