package feature

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing feature",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "feature id to update")
	updateCmd.Flags().String("display-name", "", "human friendly name of the feature")
	updateCmd.Flags().StringP("description", "d", "", "description of the feature")
	updateCmd.Flags().BoolP("enabled", "e", true, "enable or disable the feature")
	updateCmd.Flags().StringSlice("tags", []string{}, "tags associated with the feature")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateFeatureInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("feature id")
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	displayName := cmd.Config.String("display-name")
	if displayName != "" {
		input.DisplayName = &displayName
	}

	enabled := cmd.Config.Bool("enabled")
	if !enabled {
		input.Enabled = &enabled
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	return id, input, nil
}

// update an existing feature in the platform
func update(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	f, err := client.UpdateFeature(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(f)
}
