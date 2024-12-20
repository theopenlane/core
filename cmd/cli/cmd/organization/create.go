package org

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new organization",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "name of the organization")
	createCmd.Flags().StringP("display-name", "s", "", "display name of the organization")
	createCmd.Flags().StringP("description", "d", "", "description of the organization")
	createCmd.Flags().StringP("parent-org-id", "p", "", "parent organization id, leave empty to create a root org")
	createCmd.Flags().StringSlice("tags", []string{}, "tags associated with the organization")

	// TODO: https://github.com/theopenlane/core/issues/734
	// remove flag once the feature is implemented
	createCmd.Flags().BoolP("dedicated-db", "D", false, "create a dedicated database for the organization")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateOrganizationInput, err error) {
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("organization name")
	}

	displayName := cmd.Config.String("display-name")
	if displayName != "" {
		input.DisplayName = &displayName
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	parentOrgID := cmd.Config.String("parent-org-id")
	if parentOrgID != "" {
		input.ParentID = &parentOrgID
	}

	dedicatedDB := cmd.Config.Bool("dedicated-db")
	if dedicatedDB {
		input.DedicatedDb = &dedicatedDB
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	return input, nil
}

// create an organization in the platform
func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateOrganization(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
