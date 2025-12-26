//go:build cli

package apitokens

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
	Short: "create a new api token",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "name of the api token token")
	createCmd.Flags().StringP("description", "d", "", "description of the api token")
	createCmd.Flags().DurationP("expiration", "e", 0, "duration of the api token to be valid, leave empty to never expire")
	createCmd.Flags().StringSlice("scopes", []string{"read", "write"}, "scopes to associate with the api token")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateAPITokenInput, err error) {
	name := cmd.Config.String("name")
	if name == "" {
		return input, cmd.NewRequiredFieldMissingError("token name")
	}

	input = graphclient.CreateAPITokenInput{
		Name:   name,
		Scopes: cmd.Config.Strings("scopes"),
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	expiration := cmd.Config.Duration("expiration")
	if expiration != 0 {
		input.ExpiresAt = lo.ToPtr(time.Now().Add(expiration))
	}

	return input, nil
}

// create a new api token
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

	o, err := client.CreateAPIToken(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
