//go:build cli

package jobresult

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects/storage"
	openlaneclient "github.com/theopenlane/go-client"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing jobresult",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context(), cmd)
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "jobresult id to update")
	updateCmd.Flags().StringArrayP("files", "f", []string{}, "files to upload as evidence")
	updateCmd.Flags().StringP("status", "s", "", "job execution status")
	updateCmd.Flags().StringP("scheduled-job-id", "j", "", "scheduled job ID")
	updateCmd.Flags().StringP("log", "l", "", "log output from the job")
	updateCmd.Flags().BoolP("clear-log", "", false, "clear the log field")
	updateCmd.Flags().StringP("owner-id", "o", "", "owner ID")
	updateCmd.Flags().BoolP("clear-owner", "", false, "clear the owner field")
	updateCmd.Flags().StringP("file-id", "", "", "file ID reference")
}

// updateValidation validates the required fields for the command
func updateValidation(cobraCmd *cobra.Command) (id string, input openlaneclient.UpdateJobResultInput, uploads []*graphql.Upload, err error) {
	id, err = cobraCmd.Flags().GetString("id")
	if err != nil {
		return id, input, nil, err
	}
	if id == "" {
		return id, input, nil, cmd.NewRequiredFieldMissingError("id")
	}

	// Parse optional status
	if statusStr, err := cobraCmd.Flags().GetString("status"); err == nil && statusStr != "" {
		statusEnum := enums.ToJobExecutionStatus(statusStr)
		if *statusEnum == enums.JobExecutionStatusInvalid {
			return id, input, nil, fmt.Errorf("invalid status value: %s (valid values: success, failed, pending, canceled)", statusStr)
		}
		status := *statusEnum
		input.Status = &status
	}

	// Parse optional fields
	if scheduledJobID, err := cobraCmd.Flags().GetString("scheduled-job-id"); err == nil && scheduledJobID != "" {
		input.ScheduledJobID = &scheduledJobID
	}

	if log, err := cobraCmd.Flags().GetString("log"); err == nil && log != "" {
		input.Log = &log
	}

	if clearLog, err := cobraCmd.Flags().GetBool("clear-log"); err == nil && clearLog {
		input.ClearLog = &clearLog
	}

	if ownerID, err := cobraCmd.Flags().GetString("owner-id"); err == nil && ownerID != "" {
		input.OwnerID = &ownerID
	}

	if clearOwner, err := cobraCmd.Flags().GetBool("clear-owner"); err == nil && clearOwner {
		input.ClearOwner = &clearOwner
	}

	if fileID, err := cobraCmd.Flags().GetString("file-id"); err == nil && fileID != "" {
		input.FileID = &fileID
	}

	// Parse files
	files, err := cobraCmd.Flags().GetStringArray("files")
	if err != nil {
		return id, input, nil, err
	}

	for _, file := range files {
		u, err := storage.NewUploadFile(file)
		if err != nil {
			return id, input, nil, err
		}
		uploads = append(uploads, &graphql.Upload{
			File:        u.RawFile,
			Filename:    u.OriginalName,
			Size:        u.Size,
			ContentType: u.ContentType,
		})
	}

	return id, input, uploads, nil
}

// update an existing jobresult in the platform
func update(ctx context.Context, cobraCmd *cobra.Command) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, uploads, err := updateValidation(cobraCmd)
	cobra.CheckErr(err)

	o, err := client.UpdateJobResult(ctx, id, input, uploads)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
