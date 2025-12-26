//go:build cli

package task

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/go-client/graphclient"
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
	updateCmd.Flags().StringP("title", "t", "", "title of the task")
	updateCmd.Flags().StringP("details", "d", "", "details of the task")
	updateCmd.Flags().StringP("status", "s", "", "status of the task")
	updateCmd.Flags().StringP("assignee", "a", "", "assignee (user ID) of the task")
	updateCmd.Flags().Duration("due", 0, "time until due date of the task")
	updateCmd.Flags().StringP("add-group", "g", "", "group ID to own the task, this will give the group access to the task")
	updateCmd.Flags().String("remove-group", "", "group ID to own the task, this will give the group access to the task")
	updateCmd.Flags().StringP("add-comment", "c", "", "add a comment to the task")
	updateCmd.Flags().String("delete-comment", "", "delete a comment by ID from the task")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateTaskInput, err error) {
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

	due := cmd.Config.Duration("due")
	if due > 0 {
		var err error
		input.Due, err = models.ToDateTime(time.Now().Add(due).String())
		if err != nil {
			return "", input, err
		}
	}

	group := cmd.Config.String("add-group")
	if group != "" {
		input.AddGroupIDs = []string{group}
	}

	group = cmd.Config.String("remove-group")
	if group != "" {
		input.RemoveGroupIDs = []string{group}
	}

	comment := cmd.Config.String("add-comment")
	if comment != "" {
		input.AddComment = &graphclient.CreateNoteInput{
			Text: comment,
		}
	}

	deleteCommentID := cmd.Config.String("delete-comment")
	if deleteCommentID != "" {
		input.RemoveCommentIDs = []string{deleteCommentID}
	}

	return id, input, nil
}

// update an existing task in the platform
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

	o, err := client.UpdateTask(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
