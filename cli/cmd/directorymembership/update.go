//go:build cli

package directorymembership

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing directory membership",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "directory membership id to update")
	updateCmd.Flags().String("role", "", "membership role reported by the provider")
	updateCmd.Flags().String("source", "", "mechanism used to populate the membership")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateDirectoryMembershipInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("directory membership id")
	}

	role := cmd.Config.String("role")
	if role != "" {
		r := enums.DirectoryMembershipRole(role)
		input.Role = &r
	}

	source := cmd.Config.String("source")
	if source != "" {
		input.Source = &source
	}

	return id, input, nil
}

// update an existing directory membership
func update(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateDirectoryMembership(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
