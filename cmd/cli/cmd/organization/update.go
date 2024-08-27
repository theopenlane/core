package org

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
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
	cmd.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "org id to update")
	updateCmd.Flags().StringP("name", "n", "", "name of the organization")
	updateCmd.Flags().StringP("display-name", "s", "", "display name of the organization")
	updateCmd.Flags().StringP("description", "d", "", "description of the organization")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateOrganizationInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("organization id")
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

	return id, input, nil
}

// update an existing organization in the platform
func update(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateOrganization(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
