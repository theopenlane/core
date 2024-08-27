package events

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// cmd represents the base event command when called without any subcommands
var command = &cobra.Command{
	Use:   "event",
	Short: "the subcommands for working with events",
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
	case *openlaneclient.GetAllEvents:
		var nodes []*openlaneclient.GetAllEvents_Events_Edges_Node

		for _, i := range v.Events.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetEventByID:
		e = v.Event
	case *openlaneclient.CreateEvent:
		e = v.CreateEvent.Event
	case *openlaneclient.UpdateEvent:
		e = v.UpdateEvent.Event
	case *openlaneclient.DeleteEvent:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.Event

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.Event
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
func tableOutput(out []openlaneclient.Event) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "EventType", "EventMetadata", "CorrelationID")

	for _, i := range out {
		writer.AddRow(i.ID, i.EventType, i.Metadata, i.CorrelationID)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted plan in a table format
func deletedTableOutput(e *openlaneclient.DeleteEvent) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteEvent.DeletedID)

	writer.Render()
}
