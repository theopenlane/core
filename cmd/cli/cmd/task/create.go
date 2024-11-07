package task

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new task",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("title", "t", "", "title of the task")
	createCmd.Flags().StringP("description", "d", "", "description of the task")
	createCmd.Flags().StringP("status", "s", "", "status of the task")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateTaskInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.Title = cmd.Config.String("title")
	if input.Title == "" {
		return input, cmd.NewRequiredFieldMissingError("task title")
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = enums.ToTaskStatus(status)
	}

	// TODO: add additional fields for the task entity

	return input, nil
}

// create a new task
func create(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateTask(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
