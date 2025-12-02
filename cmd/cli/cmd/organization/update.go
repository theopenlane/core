//go:build cli

package org

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/objects/storage"
	openlaneclient "github.com/theopenlane/go-client"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing organization",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "org id to update")
	updateCmd.Flags().StringP("name", "n", "", "name of the organization")
	updateCmd.Flags().StringP("display-name", "s", "", "display name of the organization")
	updateCmd.Flags().StringP("description", "d", "", "description of the organization")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateOrganizationInput, avatarFile *graphql.Upload, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, nil, cmd.NewRequiredFieldMissingError("organization id")
	}

	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
	}

	displayName := cmd.Config.String("display-name")
	if displayName != "" {
		input.DisplayName = &displayName
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	avatarFileLoc := cmd.Config.String("avatar-file")
	if avatarFileLoc != "" {
		file, err := storage.NewUploadFile(avatarFileLoc)
		if err != nil {
			return id, input, nil, err
		}

		avatarFile = &graphql.Upload{
			File:        file.RawFile,
			Filename:    file.OriginalName,
			Size:        file.Size,
			ContentType: file.ContentType,
		}
	}

	return id, input, avatarFile, nil
}

// update an existing organization in the platform
func update(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, avatarFile, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateOrganization(ctx, id, input, avatarFile)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
