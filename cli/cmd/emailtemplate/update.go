//go:build cli

package emailtemplate

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing email template",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "email template id to update")
	updateCmd.Flags().StringP("key", "k", "", "stable identifier for the template")
	updateCmd.Flags().StringP("name", "n", "", "display name for the template")
	updateCmd.Flags().StringP("description", "d", "", "description of the template")
	updateCmd.Flags().String("locale", "", "locale for the template")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateEmailTemplateInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("email template id")
	}

	key := cmd.Config.String("key")
	if key != "" {
		input.Key = &key
	}

	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	return id, input, nil
}

// update an existing email template
func update(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateEmailTemplate(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
