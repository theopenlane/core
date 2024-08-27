package entitlementplan

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new entitlement plan",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "short name of the plan, must be unique")
	createCmd.Flags().String("display-name", "", "human friendly name of the plan")
	createCmd.Flags().StringP("description", "d", "", "description of the plan")
	createCmd.Flags().StringP("version", "v", "", "version of the plan")
	createCmd.Flags().StringSlice("tags", []string{}, "tags associated with the plan")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateEntitlementPlanInput, err error) {
	name := cmd.Config.String("name")
	if name == "" {
		return input, cmd.NewRequiredFieldMissingError("plan name")
	}

	version := cmd.Config.String("version")
	if version == "" {
		return input, cmd.NewRequiredFieldMissingError("version")
	}

	input = openlaneclient.CreateEntitlementPlanInput{
		Name:    name,
		Version: version,
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	displayName := cmd.Config.String("display-name")
	if displayName != "" {
		input.DisplayName = &displayName
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	return input, nil
}

// create a plan in the platform
func create(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	input, err := createValidation()
	cobra.CheckErr(err)

	p, err := client.CreateEntitlementPlan(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(p)
}
