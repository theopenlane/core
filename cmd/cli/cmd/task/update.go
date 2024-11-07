package task

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing task",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "task id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("title", "n", "", "title of the task")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateTaskInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("task id")
	}

	// validation of required fields for the update command
	// output the input struct with the required fields and optional fields based on the command line flags
	title := cmd.Config.String("title")
	if title != "" {
		input.Title = &title
	}

	return id, input, nil
}

// update an existing task in the platform
func update(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateTask(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
