//go:build cli

package usersetting

import (
	"context"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	openlaneclient "github.com/theopenlane/go-client"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing user setting",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "user setting id to update")
	updateCmd.Flags().String("status", "", "status of the user - active, inactive, suspended")
	updateCmd.Flags().StringP("default-org", "o", "", "default organization id")
	updateCmd.Flags().StringSliceP("tags", "t", []string{}, "tags associated with the user")
	updateCmd.Flags().BoolP("silence-notifications", "s", false, "silence notifications")
	updateCmd.Flags().Bool("enable-2fa", false, "enable 2fa authentication")
	updateCmd.Flags().Bool("disable-2fa", false, "disable 2fa authentication")
}

// updateValidation validates the input flags provided by the user
func updateValidation() (id string, input openlaneclient.UpdateUserSettingInput, err error) {
	id = cmd.Config.String("id")

	// set default org if provided
	defaultOrg := cmd.Config.String("default-org")
	if defaultOrg != "" {
		input.DefaultOrgID = &defaultOrg
	}

	// set silenced at time if silence flag is set
	if cmd.Config.Bool("silence") {
		now := time.Now().UTC()
		input.SilencedAt = &now
	} else {
		input.SilencedAt = nil
	}

	// explicitly set 2fa if provided to avoid wiping setting unintentionally
	if cmd.Config.Bool("enable-2fa") {
		input.IsTfaEnabled = lo.ToPtr(true)
	} else if cmd.Config.Bool("disable-2fa") {
		input.IsTfaEnabled = lo.ToPtr(false)
	}

	// add tags to the input if provided
	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	// add status to the input if provided
	status := cmd.Config.String("status")
	if status != "" {
		input.Status = enums.ToUserStatus(status)
	}

	return id, input, nil
}

// update an existing user setting
func update(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	if id == "" {
		// get the user settings id
		settings, err := client.GetAllUserSettings(ctx)
		cobra.CheckErr(err)

		// this should never happen, but just in case
		if len(settings.GetUserSettings().Edges) == 0 {
			return cmd.ErrNotFound
		}

		id = settings.GetUserSettings().Edges[0].Node.ID
	}

	// update the user settings
	o, err := client.UpdateUserSetting(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
