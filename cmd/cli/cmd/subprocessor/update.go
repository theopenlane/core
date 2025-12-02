//go:build cli

package subprocessor

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/objects/storage"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing subprocessor",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "subprocessor id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("name", "n", "", "name of the subprocessor")
	updateCmd.Flags().StringP("description", "d", "", "description of the subprocessor")
	updateCmd.Flags().StringP("logo-remote-url", "l", "", "remote URL of the logo")
	updateCmd.Flags().StringP("logo-file", "f", "", "local path to logo file to upload")
	updateCmd.Flags().StringSliceP("tags", "t", []string{}, "tags of the subprocessor")
	updateCmd.Flags().BoolP("clear-description", "", false, "clear the description field")
	updateCmd.Flags().BoolP("clear-logo-remote-url", "", false, "clear the logo remote URL field")
	updateCmd.Flags().BoolP("clear-tags", "", false, "clear all tags")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateSubprocessorInput, logoFile *graphql.Upload, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, nil, cmd.NewRequiredFieldMissingError("subprocessor id")
	}

	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
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
			return id, input, nil, err
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

	// Handle clear flags
	if cmd.Config.Bool("clear-description") {
		clearDescription := true
		input.ClearDescription = &clearDescription
	}

	if cmd.Config.Bool("clear-logo-remote-url") {
		clearLogoRemoteURL := true
		input.ClearLogoRemoteURL = &clearLogoRemoteURL
	}

	if cmd.Config.Bool("clear-tags") {
		clearTags := true
		input.ClearTags = &clearTags
	}

	return id, input, logoFile, nil
}

// update an existing subprocessor in the platform
func update(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, logoFile, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateSubprocessor(ctx, id, input, logoFile)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
