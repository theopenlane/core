//go:build cli

package entitytype

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/go-client/graphclient"
)

// cmd represents the base entity type command when called without any subcommands
var command = &cobra.Command{
	Use:   "entity-type",
	Short: "the subcommands for working with entity types",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the entity types in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the entity types and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllEntityTypes:
		var nodes []*graphclient.GetAllEntityTypes_EntityTypes_Edges_Node

		for _, i := range v.EntityTypes.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetEntityTypes:
		var nodes []*graphclient.GetEntityTypes_EntityTypes_Edges_Node

		for _, i := range v.EntityTypes.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetEntityTypeByID:
		e = v.EntityType
	case *graphclient.CreateEntityType:
		e = v.CreateEntityType.EntityType
	case *graphclient.UpdateEntityType:
		e = v.UpdateEntityType.EntityType
	case *graphclient.DeleteEntityType:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.EntityType

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.EntityType
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
func tableOutput(out []graphclient.EntityType) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name")
	for _, i := range out {
		writer.AddRow(i.ID, i.Name)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteEntityType) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteEntityType.DeletedID)

	writer.Render()
}
