//go:build cli

package directorygroup

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new directory group",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("external-id", "e", "", "stable identifier from the directory system")
	createCmd.Flags().String("integration-id", "", "integration that owns this directory group")
	createCmd.Flags().String("directory-sync-run-id", "", "sync run that produced this snapshot")
	createCmd.Flags().String("display-name", "", "display name of the group")
	createCmd.Flags().String("email", "", "primary group email address")
	createCmd.Flags().StringP("description", "d", "", "free-form description")
	createCmd.Flags().String("directory-instance-id", "", "stable external workspace or tenant identifier")
	createCmd.Flags().String("platform-id", "", "platform associated with this directory group")
	createCmd.Flags().StringSlice("tags", []string{}, "tags associated with the group")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateDirectoryGroupInput, err error) {
	externalID := cmd.Config.String("external-id")
	if externalID == "" {
		return input, cmd.NewRequiredFieldMissingError("external id")
	}

	input.ExternalID = externalID

	integrationID := cmd.Config.String("integration-id")
	if integrationID == "" {
		return input, cmd.NewRequiredFieldMissingError("integration id")
	}

	input.IntegrationID = integrationID

	directorySyncRunID := cmd.Config.String("directory-sync-run-id")
	if directorySyncRunID == "" {
		return input, cmd.NewRequiredFieldMissingError("directory sync run id")
	}

	input.DirectorySyncRunID = directorySyncRunID

	displayName := cmd.Config.String("display-name")
	if displayName != "" {
		input.DisplayName = &displayName
	}

	email := cmd.Config.String("email")
	if email != "" {
		input.Email = &email
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	directoryInstanceID := cmd.Config.String("directory-instance-id")
	if directoryInstanceID != "" {
		input.DirectoryInstanceID = &directoryInstanceID
	}

	platformID := cmd.Config.String("platform-id")
	if platformID != "" {
		input.PlatformID = &platformID
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	return input, nil
}

// create a new directory group
func create(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateDirectoryGroup(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
