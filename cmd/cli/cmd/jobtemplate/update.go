//go:build cli

package jobtemplate

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing jobtemplate",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "jobtemplate id to update")

	// command line flags for the update command

	// example:
	// updateCmd.Flags().StringP("name", "n", "", "name of the jobtemplate")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateJobTemplateInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("jobtemplate id")
	}

	// validation of required fields for the update command
	// output the input struct with the required fields and optional fields based on the command line flags

	return id, input, nil
}

// update an existing jobtemplate in the platform
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

	o, err := client.UpdateJobTemplate(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
