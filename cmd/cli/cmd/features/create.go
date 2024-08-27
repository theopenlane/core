package feature

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new feature",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	cmd.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "name of the feature")
	createCmd.Flags().String("display-name", "", "human friendly name of the feature")
	createCmd.Flags().StringP("description", "d", "", "description of the feature")
	createCmd.Flags().BoolP("enabled", "e", true, "enabled status of the feature")
	createCmd.Flags().StringSlice("tags", []string{}, "tags associated with the plan")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateFeatureInput, err error) {
	name := cmd.Config.String("name")
	if name == "" {
		return input, cmd.NewRequiredFieldMissingError("feature name")
	}

	enabled := cmd.Config.Bool("enabled")

	input = openlaneclient.CreateFeatureInput{
		Name:    name,
		Enabled: &enabled,
	}

	displayName := cmd.Config.String("display-name")
	if displayName != "" {
		input.DisplayName = &displayName
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	return input, nil
}

// create a new feature in the platform
func create(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	input, err := createValidation()
	cobra.CheckErr(err)

	f, err := client.CreateFeature(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(f)
}
