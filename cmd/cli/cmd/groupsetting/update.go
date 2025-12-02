//go:build cli

package groupsetting

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing group setting",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "group setting id to update")
	updateCmd.Flags().StringP("visibility", "v", "", "visibility of the group")
	updateCmd.Flags().StringP("join-policy", "j", "", "join policy of the group")
	updateCmd.Flags().BoolP("sync-to-slack", "s", false, "sync group members to slack")
	updateCmd.Flags().BoolP("sync-to-github", "g", false, "sync group members to github")
}

// updateValidation validates the input flags provided by the user
func updateValidation() (id string, input openlaneclient.UpdateGroupSettingInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("setting id")
	}

	visibility := cmd.Config.String("visibility")
	if visibility != "" {
		input.Visibility = enums.ToGroupVisibility(visibility)
	}

	joinPolicy := cmd.Config.String("join-policy")
	if joinPolicy != "" {
		input.JoinPolicy = enums.ToGroupJoinPolicy(joinPolicy)
	}

	syncToSlack := cmd.Config.Bool("sync-to-slack")
	if syncToSlack {
		input.SyncToSlack = &syncToSlack
	}

	syncToGithub := cmd.Config.Bool("sync-to-github")
	if syncToGithub {
		input.SyncToGithub = &syncToGithub
	}

	return id, input, nil
}

// update an existing group setting in the platform
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

	o, err := client.UpdateGroupSetting(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
