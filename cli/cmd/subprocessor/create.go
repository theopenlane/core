//go:build cli

package subprocessor

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new subprocessor",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("name", "n", "", "name of the subprocessor")
	createCmd.Flags().StringP("description", "d", "", "description of the subprocessor")
	createCmd.Flags().StringP("logo-remote-url", "l", "", "remote URL of the logo")
	createCmd.Flags().StringP("logo-file", "f", "", "local path to logo file to upload")
	createCmd.Flags().StringSliceP("tags", "t", []string{}, "tags of the subprocessor")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateSubprocessorInput, logoFile *graphql.Upload, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, nil, cmd.NewRequiredFieldMissingError("name")
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	logoRemoteURL := cmd.Config.String("logo-remote-url")
	if logoRemoteURL != "" {
		input.LogoRemoteURL = &logoRemoteURL
	}

	logoFileLoc := cmd.Config.String("logo-file")
	if logoFileLoc != "" {
		file, err := storage.NewUploadFile(logoFileLoc)
		if err != nil {
			return input, nil, err
		}

		logoFile = &graphql.Upload{
			File:        file.RawFile,
			Filename:    file.OriginalName,
			Size:        file.Size,
			ContentType: file.ContentType,
		}
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	return input, logoFile, nil
}

// create a new subprocessor
func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, logoFile, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateSubprocessor(ctx, input, logoFile)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
