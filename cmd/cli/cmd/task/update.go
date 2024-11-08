package task

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
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
	updateCmd.Flags().StringP("description", "d", "", "description of the task")
	updateCmd.Flags().StringP("status", "s", "", "status of the task")
	updateCmd.Flags().StringP("assignee", "a", "", "assignee (user ID) of the task")
	updateCmd.Flags().Duration("due", 0, "time until due date of the task")
	updateCmd.Flags().StringP("organization", "o", "", "organization ID of the task to own the task, this will give the organization access to the task")
	updateCmd.Flags().StringP("group", "g", "", "group ID of the task to own the task, this will give the group access to the task")
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

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = enums.ToTaskStatus(status)
	}

	assignee := cmd.Config.String("assignee")
	if assignee != "" {
		input.Assignee = &assignee
	}

	due := cmd.Config.Duration("due")
	if due != 0 {
		dueDate := time.Now().Add(due)
		input.Due = &dueDate
	}

	organization := cmd.Config.String("organization")
	if organization != "" {
		input.AddOrganizationIDs = []string{organization}
	}

	group := cmd.Config.String("group")
	if group != "" {
		input.AddGroupIDs = []string{group}
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
