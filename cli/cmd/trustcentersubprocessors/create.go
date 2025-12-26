//go:build cli

package trustcentersubprocessors

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new trust center subprocessor",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("subprocessor-id", "s", "", "ID of the subprocessor")
	createCmd.Flags().StringP("trust-center-id", "t", "", "ID of the trust center")
	createCmd.Flags().StringP("category", "c", "", "category of the subprocessor (e.g. 'Data Warehouse' or 'Infrastructure Hosting')")
	createCmd.Flags().StringSliceP("countries", "", []string{}, "country codes or countries where the subprocessor is located")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateTrustCenterSubprocessorInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.SubprocessorID = cmd.Config.String("subprocessor-id")
	if input.SubprocessorID == "" {
		return input, cmd.NewRequiredFieldMissingError("subprocessor-id")
	}

	input.Category = cmd.Config.String("category")
	if input.Category == "" {
		return input, cmd.NewRequiredFieldMissingError("category")
	}

	trustCenterID := cmd.Config.String("trust-center-id")
	if trustCenterID != "" {
		input.TrustCenterID = &trustCenterID
	}

	countries := cmd.Config.Strings("countries")
	if len(countries) > 0 {
		input.Countries = countries
	}

	return input, nil
}

// create a new trust center subprocessor
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

	o, err := client.CreateTrustCenterSubprocessor(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
