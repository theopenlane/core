//go:build cli

package directorymembership

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new directory membership",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().String("integration-id", "", "integration that owns this membership")
	createCmd.Flags().String("directory-sync-run-id", "", "sync run that produced this snapshot")
	createCmd.Flags().String("directory-account-id", "", "directory account participating in this membership")
	createCmd.Flags().String("directory-group-id", "", "directory group associated with this membership")
	createCmd.Flags().String("role", "", "membership role reported by the provider")
	createCmd.Flags().String("source", "", "mechanism used to populate the membership (api, scim, csv, etc)")
	createCmd.Flags().String("directory-instance-id", "", "stable external workspace or tenant identifier")
	createCmd.Flags().String("platform-id", "", "platform associated with this membership")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateDirectoryMembershipInput, err error) {
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

	directoryAccountID := cmd.Config.String("directory-account-id")
	if directoryAccountID == "" {
		return input, cmd.NewRequiredFieldMissingError("directory account id")
	}

	input.DirectoryAccountID = directoryAccountID

	directoryGroupID := cmd.Config.String("directory-group-id")
	if directoryGroupID == "" {
		return input, cmd.NewRequiredFieldMissingError("directory group id")
	}

	input.DirectoryGroupID = directoryGroupID

	role := cmd.Config.String("role")
	if role != "" {
		r := enums.DirectoryMembershipRole(role)
		input.Role = &r
	}

	source := cmd.Config.String("source")
	if source != "" {
		input.Source = &source
	}

	directoryInstanceID := cmd.Config.String("directory-instance-id")
	if directoryInstanceID != "" {
		input.DirectoryInstanceID = &directoryInstanceID
	}

	platformID := cmd.Config.String("platform-id")
	if platformID != "" {
		input.PlatformID = &platformID
	}

	return input, nil
}

// create a new directory membership
func create(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateDirectoryMembership(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
