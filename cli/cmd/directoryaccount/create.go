//go:build cli

package directoryaccount

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new directory account",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("external-id", "e", "", "stable identifier from the directory system")
	createCmd.Flags().String("canonical-email", "", "primary email address of the account")
	createCmd.Flags().String("display-name", "", "display name of the account")
	createCmd.Flags().String("given-name", "", "first name reported by the provider")
	createCmd.Flags().String("family-name", "", "last name reported by the provider")
	createCmd.Flags().String("job-title", "", "title captured at sync time")
	createCmd.Flags().String("department", "", "department captured at sync time")
	createCmd.Flags().String("directory-name", "", "directory source label (e.g. googleworkspace, github, slack)")
	createCmd.Flags().String("directory-instance-id", "", "stable external workspace or tenant identifier")
	createCmd.Flags().String("integration-id", "", "integration that owns this directory account")
	createCmd.Flags().String("directory-sync-run-id", "", "sync run that produced this snapshot")
	createCmd.Flags().String("platform-id", "", "platform associated with this directory account")
	createCmd.Flags().StringSlice("tags", []string{}, "tags associated with the account")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateDirectoryAccountInput, err error) {
	externalID := cmd.Config.String("external-id")
	if externalID == "" {
		return input, cmd.NewRequiredFieldMissingError("external id")
	}

	input.ExternalID = externalID

	canonicalEmail := cmd.Config.String("canonical-email")
	if canonicalEmail != "" {
		input.CanonicalEmail = &canonicalEmail
	}

	displayName := cmd.Config.String("display-name")
	if displayName != "" {
		input.DisplayName = &displayName
	}

	givenName := cmd.Config.String("given-name")
	if givenName != "" {
		input.GivenName = &givenName
	}

	familyName := cmd.Config.String("family-name")
	if familyName != "" {
		input.FamilyName = &familyName
	}

	jobTitle := cmd.Config.String("job-title")
	if jobTitle != "" {
		input.JobTitle = &jobTitle
	}

	department := cmd.Config.String("department")
	if department != "" {
		input.Department = &department
	}

	directoryName := cmd.Config.String("directory-name")
	if directoryName != "" {
		input.DirectoryName = &directoryName
	}

	directoryInstanceID := cmd.Config.String("directory-instance-id")
	if directoryInstanceID != "" {
		input.DirectoryInstanceID = &directoryInstanceID
	}

	integrationID := cmd.Config.String("integration-id")
	if integrationID != "" {
		input.IntegrationID = &integrationID
	}

	directorySyncRunID := cmd.Config.String("directory-sync-run-id")
	if directorySyncRunID != "" {
		input.DirectorySyncRunID = &directorySyncRunID
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

// create a new directory account
func create(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateDirectoryAccount(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
