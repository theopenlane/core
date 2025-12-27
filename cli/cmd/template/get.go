//go:build cli

package templates

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get details of existing template(s)",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "get a specific template by ID")
}

// get retrieves templates from the platform
func get(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id := cmd.Config.String("id")

	if id != "" {
		o, err := client.GetTemplateByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	var order *graphclient.TemplateOrder
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.TemplateOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.TemplateOrderField(*cmd.OrderBy),
		}
	}

	o, err := client.GetAllTemplates(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.TemplateOrder{order})
	cobra.CheckErr(err)

	return consoleOutput(o)
}
