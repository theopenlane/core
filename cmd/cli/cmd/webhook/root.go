package webhooks

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// cmd represents the base webhook command when called without any subcommands
var command = &cobra.Command{
	Use:   "webhook",
	Short: "the subcommands for working with webhooks",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the output in JSON format
	if cmd.OutputFormat == cmd.JSONOutput {
		return jsonOutput(e)
	}

	// check the type of the output and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllWebhooks:
		var nodes []*openlaneclient.GetAllWebhooks_Webhooks_Edges_Node

		for _, i := range v.Webhooks.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetWebhookByID:
		e = v.Webhook
	case *openlaneclient.CreateWebhook:
		e = v.CreateWebhook.Webhook
	case *openlaneclient.UpdateWebhook:
		e = v.UpdateWebhook.Webhook
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.Webhook

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.Webhook
		err = json.Unmarshal(s, &in)
		cobra.CheckErr(err)

		list = append(list, in)
	}

	tableOutput(list)

	return nil
}

// jsonOutput prints the output in a JSON format
func jsonOutput(out any) error {
	s, err := json.Marshal(out)
	cobra.CheckErr(err)

	return cmd.JSONPrint(s)
}

// tableOutput prints the output in a table format
func tableOutput(out []openlaneclient.Webhook) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Description", "Destination URL", "Enabled")
	for _, i := range out {
		// this doesn't visually show you the json in the table but leaving it in for now
		writer.AddRow(i.ID, i.Name, *i.Description, i.DestinationURL, i.Enabled)
	}

	writer.Render()
}
