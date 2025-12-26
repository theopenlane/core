//go:build cli

package scheduledjob

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new scheduledjob",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command

	// example:
	// createCmd.Flags().StringP("name", "n", "", "name of the scheduledjob")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateScheduledJobInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags

	return input, nil
}

// create a new scheduledjob
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

	o, err := client.CreateScheduledJob(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
