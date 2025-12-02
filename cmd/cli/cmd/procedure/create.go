//go:build cli

package procedure

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
	Short: "create a new procedure",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("name", "n", "", "name of the procedure")
	createCmd.Flags().StringP("details", "d", "", "details of the procedure")
	createCmd.Flags().StringP("status", "s", "", "status of the procedure e.g. draft, published, archived, etc.")
	createCmd.Flags().StringP("type", "t", "", "type of the procedure")
	createCmd.Flags().StringP("revision", "v", models.DefaultRevision, "revision of the procedure")
	createCmd.Flags().StringP("file", "f", "", "local path to file to upload as the procedure details")
	createCmd.Flags().StringP("url", "u", "", "url to use as the procedure details")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateProcedureInput, detailsFile *graphql.Upload, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	detailsFileLoc := cmd.Config.String("file")
	if detailsFileLoc != "" {
		file, err := storage.NewUploadFile(detailsFileLoc)
		if err != nil {
			return input, nil, err
		}

		detailsFile = &graphql.Upload{
			File:        file.RawFile,
			Filename:    file.OriginalName,
			Size:        file.Size,
			ContentType: file.ContentType,
		}
	}

	input.Name = cmd.Config.String("name")
	if input.Name == "" && detailsFile == nil {
		return input, detailsFile, cmd.NewRequiredFieldMissingError("name")
	}

	details := cmd.Config.String("details")
	if details != "" {
		input.Details = &details
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = enums.ToDocumentStatus(status)
	}

	procedureType := cmd.Config.String("type")
	if procedureType != "" {
		input.ProcedureType = &procedureType
	}

	revision := cmd.Config.String("revision")
	if revision != "" {
		input.Revision = &revision
	}

	url := cmd.Config.String("url")
	if url != "" {
		input.URL = &url
	}

	return input, detailsFile, nil
}

// create a new procedure
func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, detailsFile, err := createValidation()
	cobra.CheckErr(err)

	if detailsFile != nil {
		o, err := client.CreateUploadProcedure(ctx, *detailsFile, nil)
		cobra.CheckErr(err)

		return consoleOutput(o)

	}

	o, err := client.CreateProcedure(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
