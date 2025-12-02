//go:build cli

package user

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
	Short: "update an existing user",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "user id to update")
	updateCmd.Flags().StringP("first-name", "f", "", "first name of the user")
	updateCmd.Flags().StringP("last-name", "l", "", "last name of the user")
	updateCmd.Flags().StringP("display-name", "d", "", "display name of the user")
	updateCmd.Flags().StringP("email", "e", "", "email of the user")
	updateCmd.Flags().StringP("avatar-file", "a", "", "local of avatar file to upload")
}

// updateValidation validates the input flags provided by the user
func updateValidation() (id string, input openlaneclient.UpdateUserInput, avatarFile *graphql.Upload, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, nil, cmd.NewRequiredFieldMissingError("user id")
	}

	firstName := cmd.Config.String("first-name")
	if firstName != "" {
		input.FirstName = &firstName
	}

	lastName := cmd.Config.String("last-name")
	if lastName != "" {
		input.LastName = &lastName
	}

	displayName := cmd.Config.String("display-name")
	if displayName != "" {
		input.DisplayName = &displayName
	}

	email := cmd.Config.String("email")
	if email != "" {
		input.Email = &email
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

	// TODO: allow updates to user settings
	return id, input, avatarFile, nil
}

// update an existing user
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

	o, err := client.UpdateUser(ctx, id, input, avatarFile)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
