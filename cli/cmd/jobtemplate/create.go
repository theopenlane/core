//go:build cli

package jobtemplate

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new jobtemplate",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command

	// example:
	// createCmd.Flags().StringP("name", "n", "", "name of the jobtemplate")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateJobTemplateInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags

	return input, nil
}

// create a new jobtemplate
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

	o, err := client.CreateJobTemplate(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
