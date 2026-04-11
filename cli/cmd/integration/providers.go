//go:build cli

package integrations

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
)

var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "list available integration providers",
	Run: func(cmd *cobra.Command, args []string) {
		err := listProviders(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(providersCmd)
}

// listProviders retrieves all available integration provider definitions
func listProviders(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	o, err := client.ListIntegrationProviders(ctx)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
