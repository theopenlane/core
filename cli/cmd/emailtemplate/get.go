//go:build cli

package emailtemplate

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get an existing email template",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "email template id to query")
}

// get an existing email template
func get(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id := cmd.Config.String("id")

	if id != "" {
		o, err := client.GetEmailTemplateByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	var order *graphclient.EmailTemplateOrder
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.EmailTemplateOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.EmailTemplateOrderField(*cmd.OrderBy),
		}
	}

	o, err := client.GetAllEmailTemplates(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.EmailTemplateOrder{order})
	cobra.CheckErr(err)

	return consoleOutput(o)
}
