//go:build cli

package jobrunnerregistrationtoken

import (
	"context"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new job runner registration token",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("owner-id", "o", "", "organization ID for the runner registration token")
	createCmd.Flags().StringSliceP("tags", "t", []string{}, "tags for the runner registration token")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateJobRunnerRegistrationTokenInput, err error) {
	owner := cmd.Config.String("owner-id")
	if owner == "" {
		return input, cmd.NewRequiredFieldMissingError("owner-id")
	}
	input.OwnerID = lo.ToPtr(owner)

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	return input, nil
}

// create a new job runner registration token
func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateJobRunnerRegistrationToken(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
