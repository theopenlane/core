//go:build cli

package task

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects/storage"
	openlaneclient "github.com/theopenlane/go-client/genclient"
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
	createCmd.Flags().StringP("details", "d", "", "details of the task")
	createCmd.Flags().StringP("status", "s", "", "status of the task")
	createCmd.Flags().StringP("assignee", "a", "", "assignee (user ID) of the task")
	createCmd.Flags().String("due", "", "time until due date of the task")
	createCmd.Flags().StringP("group", "g", "", "group ID to own the task, this will give the group access to the task")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateTaskInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.Title = cmd.Config.String("title")
	if input.Title == "" {
		return input, cmd.NewRequiredFieldMissingError("task title")
	}

	details := cmd.Config.String("details")
	if details != "" {
		input.Details = &details
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = enums.ToTaskStatus(status)
	}

	assignee := cmd.Config.String("assignee")
	if assignee != "" {
		input.AssigneeID = &assignee
	}

	due := cmd.Config.String("due")
	if due != "" {
		var err error
		input.Due, err = models.ToDateTime(due)
		if err != nil {
			return input, err
		}
	}

	group := cmd.Config.String("group")
	if group != "" {
		input.GroupIDs = []string{group}
	}

	return input, nil
}

// create a new task
func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	if cmd.InputFile != "" {
		u, err := storage.NewUploadFile(cmd.InputFile)
		cobra.CheckErr(err)

		o, err := client.CreateBulkCSVTask(ctx, graphql.Upload{
			File:        u.RawFile,
			Filename:    u.OriginalName,
			Size:        u.Size,
			ContentType: u.ContentType,
		})
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateTask(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
