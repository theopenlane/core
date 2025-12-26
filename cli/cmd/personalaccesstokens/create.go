//go:build cli

package tokens

import (
	"context"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new personal access token",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "name of the personal access token")
	createCmd.Flags().StringP("description", "d", "", "description of the pat")
	createCmd.Flags().StringSliceP("organizations", "o", []string{}, "organization(s) id to associate the pat with")
	createCmd.Flags().DurationP("expiration", "e", 0, "duration of the pat to be valid, leave empty to never expire")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreatePersonalAccessTokenInput, err error) {
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("token name")
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	organizations := cmd.Config.Strings("organizations")
	if organizations != nil {
		input.OrganizationIDs = organizations
	}

	expiration := cmd.Config.Duration("expiration")
	if expiration != 0 {
		input.ExpiresAt = lo.ToPtr(time.Now().Add(expiration))
	}

	return input, nil
}

// create a new personal access token
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

	o, err := client.CreatePersonalAccessToken(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
