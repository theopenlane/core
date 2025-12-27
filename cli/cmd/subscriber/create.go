//go:build cli

package subscribers

import (
	"context"
	"os"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "add subscribers to a organization",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringSliceP("emails", "e", []string{}, "email address of the subscriber()")
	createCmd.Flags().StringSlice("tags", []string{}, "tags associated with the subscriber")
}

// createValidation validates the required fields for the command
func createValidation() (input []*graphclient.CreateSubscriberInput, err error) {
	email := cmd.Config.Strings("emails")
	if len(email) == 0 {
		return input, cmd.NewRequiredFieldMissingError("emails")
	}

	for _, e := range email {
		i := &graphclient.CreateSubscriberInput{
			Email: e,
		}

		tags := cmd.Config.Strings("tags")
		if len(tags) > 0 {
			i.Tags = tags
		}

		input = append(input, i)
	}

	return input, nil
}

func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	if cmd.InputFile != "" {
		input, err := os.OpenFile(cmd.InputFile, os.O_RDWR|os.O_CREATE, os.ModePerm)
		cobra.CheckErr(err)

		defer input.Close()

		in := graphql.Upload{
			File: input,
		}

		o, err := client.CreateBulkCSVSubscriber(ctx, in)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateBulkSubscriber(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
