//go:build cli

package trustcentersubprocessors

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing trust center subprocessor",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "trust center subprocessor id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("kind-id", "k", "", "ID of the trust center subprocessor kind")
	updateCmd.Flags().StringP("kind-name", "", "", "name of the trust center subprocessor kind (e.g. 'Data Warehouse' or 'Infrastructure Hosting')")
	updateCmd.Flags().StringSliceP("countries", "", []string{}, "country codes or countries where the subprocessor is located")
	updateCmd.Flags().StringSliceP("append-countries", "", []string{}, "append country codes or countries to the existing list")
	updateCmd.Flags().BoolP("clear-countries", "", false, "clear all countries")
	updateCmd.Flags().BoolP("clear-kind", "", false, "clear the subprocessor kind field")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateTrustCenterSubprocessorInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("trust center subprocessor id")
	}

	kindID := cmd.Config.String("kind-id")
	if kindID != "" {
		input.TrustCenterSubprocessorKindID = &kindID
	}

	kindName := cmd.Config.String("kind-name")
	if kindName != "" {
		input.TrustCenterSubprocessorKindName = &kindName
	}

	countries := cmd.Config.Strings("countries")
	if len(countries) > 0 {
		input.Countries = countries
	}

	appendCountries := cmd.Config.Strings("append-countries")
	if len(appendCountries) > 0 {
		input.AppendCountries = appendCountries
	}

	// Handle clear flags
	if cmd.Config.Bool("clear-countries") {
		clearCountries := true
		input.ClearCountries = &clearCountries
	}

	if cmd.Config.Bool("clear-kind") {
		clearKind := true
		input.ClearTrustCenterSubprocessorKind = &clearKind
	}

	return id, input, nil
}

// update an existing trust center subprocessor in the platform
func update(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateTrustCenterSubprocessor(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
