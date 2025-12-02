//go:build cli

package entitytypehistory

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
)

// cmd represents the base entityTypeHistory command when called without any subcommands
var command = &cobra.Command{
	Use:   "entity-type-history",
	Short: "the subcommands for working with entityTypeHistories",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the entityTypeHistories in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the entityTypeHistories and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllEntityTypeHistories:
		var nodes []*openlaneclient.GetAllEntityTypeHistories_EntityTypeHistories_Edges_Node

		for _, i := range v.EntityTypeHistories.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetEntityTypeHistories:
		var nodes []*openlaneclient.GetEntityTypeHistories_EntityTypeHistories_Edges_Node

		for _, i := range v.EntityTypeHistories.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.EntityTypeHistory

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.EntityTypeHistory
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
func tableOutput(out []openlaneclient.EntityTypeHistory) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Ref", "Operation", "UpdatedAt", "UpdatedBy")
	for _, i := range out {
		writer.AddRow(i.ID, *i.Ref, i.Operation, *i.UpdatedAt, *i.UpdatedBy)
	}

	writer.Render()
}
