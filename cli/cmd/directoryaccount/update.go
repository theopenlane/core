//go:build cli

package directoryaccount

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing directory account",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "directory account id to update")
	updateCmd.Flags().String("canonical-email", "", "primary email address of the account")
	updateCmd.Flags().String("display-name", "", "display name of the account")
	updateCmd.Flags().String("given-name", "", "first name reported by the provider")
	updateCmd.Flags().String("family-name", "", "last name reported by the provider")
	updateCmd.Flags().String("job-title", "", "title captured at sync time")
	updateCmd.Flags().String("department", "", "department captured at sync time")
	updateCmd.Flags().String("directory-name", "", "directory source label")
	updateCmd.Flags().StringSlice("tags", []string{}, "tags associated with the account")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateDirectoryAccountInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("directory account id")
	}

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

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	return id, input, nil
}

// update an existing directory account
func update(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateDirectoryAccount(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
