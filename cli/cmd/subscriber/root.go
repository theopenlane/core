//go:build cli

package subscribers

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/go-client/graphclient"
)

// cmd represents the base subscribers command when called without any subcommands
var command = &cobra.Command{
	Use:   "subscriber",
	Short: "the subcommands for working with the subscribers of a organization",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the out in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the out in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the out and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetSubscribers:
		var nodes []*graphclient.GetSubscribers_Subscribers_Edges_Node

		for _, i := range v.Subscribers.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetAllSubscribers:
		var nodes []*graphclient.GetAllSubscribers_Subscribers_Edges_Node

		for _, i := range v.Subscribers.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetSubscriberByEmail:
		e = v.Subscriber
	case *graphclient.CreateBulkSubscriber:
		e = v.CreateBulkSubscriber.Subscribers
	case *graphclient.CreateBulkCSVSubscriber:
		e = v.CreateBulkCSVSubscriber.Subscribers
	case *graphclient.CreateSubscriber:
		e = v.CreateSubscriber.Subscriber
	case *graphclient.DeleteSubscriber:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.Subscriber

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.Subscriber
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
func tableOutput(out []graphclient.Subscriber) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Email", "Verified", "Active")
	for _, i := range out {
		writer.AddRow(i.ID, i.Email, i.VerifiedEmail, i.Active)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted email in a table format
func deletedTableOutput(e *graphclient.DeleteSubscriber) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteSubscriber.Email)

	writer.Render()
}
