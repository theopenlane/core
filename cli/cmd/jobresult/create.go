//go:build cli

package jobresult

import (
	"context"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new jobresult",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context(), cmd)
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringArrayP("files", "f", []string{}, "files to upload as evidence")
	createCmd.Flags().StringP("status", "s", "", "job execution status (required)")
	createCmd.Flags().Int64P("exit-code", "e", 0, "exit code from the script execution")
	createCmd.Flags().StringP("scheduled-job-id", "j", "", "scheduled job ID (required)")
	createCmd.Flags().StringP("log", "l", "", "log output from the job")
	createCmd.Flags().StringP("started-at", "", "", "job start time (RFC3339 format)")
	createCmd.Flags().StringP("finished-at", "", "", "job finish time (RFC3339 format)")
	createCmd.Flags().StringP("owner-id", "o", "", "owner ID")
	createCmd.Flags().StringP("file-id", "", "", "file ID reference (required)")
}

// createValidation validates the required fields for the command
func createValidation(cobraCmd *cobra.Command) (input graphclient.CreateJobResultInput, uploads []*graphql.Upload, err error) {
	// Parse required fields
	statusStr, err := cobraCmd.Flags().GetString("status")
	if err != nil {
		return input, nil, err
	}
	if statusStr == "" {
		return input, nil, cmd.NewRequiredFieldMissingError("status")
	}

	scheduledJobID, err := cobraCmd.Flags().GetString("scheduled-job-id")
	if err != nil {
		return input, nil, err
	}
	if scheduledJobID == "" {
		return input, nil, cmd.NewRequiredFieldMissingError("scheduled-job-id")
	}

	exitCode, err := cobraCmd.Flags().GetInt64("exit-code")
	if err != nil {
		return input, nil, err
	}

	fileID, err := cobraCmd.Flags().GetString("file-id")
	if err != nil {
		return input, nil, err
	}
	if fileID == "" {
		return input, nil, cmd.NewRequiredFieldMissingError("file-id")
	}

	// Parse files
	files, err := cobraCmd.Flags().GetStringArray("files")
	if err != nil {
		return input, nil, err
	}

	for _, file := range files {
		u, err := storage.NewUploadFile(file)
		if err != nil {
			return input, nil, err
		}
		uploads = append(uploads, &graphql.Upload{
			File:        u.RawFile,
			Filename:    u.OriginalName,
			Size:        u.Size,
			ContentType: u.ContentType,
		})
	}

	// Parse status enum
	statusEnum := enums.ToJobExecutionStatus(statusStr)
	if *statusEnum == enums.JobExecutionStatusInvalid {
		return input, nil, fmt.Errorf("invalid status value: %s (valid values: success, failed, pending, canceled)", statusStr)
	}

	status := *statusEnum

	// Build input struct with required fields
	input = graphclient.CreateJobResultInput{
		Status:         status,
		ExitCode:       exitCode,
		ScheduledJobID: scheduledJobID,
		FileID:         fileID,
	}

	// Parse optional fields
	if log, err := cobraCmd.Flags().GetString("log"); err == nil && log != "" {
		input.Log = &log
	}

	if ownerID, err := cobraCmd.Flags().GetString("owner-id"); err == nil && ownerID != "" {
		input.OwnerID = &ownerID
	}

	if startedAtStr, err := cobraCmd.Flags().GetString("started-at"); err == nil && startedAtStr != "" {
		if startedAt, parseErr := time.Parse(time.RFC3339, startedAtStr); parseErr == nil {
			input.StartedAt = &startedAt
		}
	}

	if finishedAtStr, err := cobraCmd.Flags().GetString("finished-at"); err == nil && finishedAtStr != "" {
		if finishedAt, parseErr := time.Parse(time.RFC3339, finishedAtStr); parseErr == nil {
			input.FinishedAt = &finishedAt
		}
	}

	return input, uploads, nil
}

// create a new jobresult
func create(ctx context.Context, cobraCmd *cobra.Command) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	// Pass the cobra.Command to validation so we can get flags
	input, uploads, err := createValidation(cobraCmd)
	cobra.CheckErr(err)

	o, err := client.CreateJobResult(ctx, input, uploads)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
