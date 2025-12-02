//go:build cli

package control

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get an existing control",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)
	getCmd.Flags().StringP("id", "i", "", "control id to query")
	getCmd.Flags().StringP("ref-code", "r", "", "the unique reference code of the control")
}

// get an existing control in the platform
func get(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}
	// filter options
	id := cmd.Config.String("id")
	refCode := cmd.Config.String("ref-code")

	// if an control ID is provided, filter on that control
	if id != "" {
		o, err := client.GetControlByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	// if a ref code is provided, filter on that control
	if refCode != "" {
		o, err := client.GetControls(ctx, cmd.First, cmd.Last, &openlaneclient.ControlWhereInput{
			RefCode: &refCode,
		})
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	// get all will be filtered for the authorized organization(s)
	o, err := client.GetAllControls(ctx)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
