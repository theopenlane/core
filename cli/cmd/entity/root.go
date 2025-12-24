//go:build cli

package entity

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

// cmd represents the base entity command when called without any subcommands
var command = &cobra.Command{
	Use:   "entity",
	Short: "the subcommands for working with entities",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the entities in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the entities and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllEntities:
		var nodes []*graphclient.GetAllEntities_Entities_Edges_Node

		for _, i := range v.Entities.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetEntities:
		var nodes []*graphclient.GetEntities_Entities_Edges_Node

		for _, i := range v.Entities.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetEntityByID:
		e = v.Entity
	case *graphclient.CreateEntity:
		e = v.CreateEntity.Entity
	case *graphclient.UpdateEntity:
		e = v.UpdateEntity.Entity
	case *graphclient.DeleteEntity:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.Entity

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.Entity
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
func tableOutput(out []graphclient.Entity) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "DisplayName", "Description", "EntityType", "Status", "Domains")

	for _, i := range out {
		entityTypeName := ""
		if i.EntityType != nil {
			entityTypeName = i.EntityType.Name
		}

		writer.AddRow(i.ID, *i.Name, *i.DisplayName, *i.Description, entityTypeName, *i.Status, strings.Join(i.Domains, ", "))
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteEntity) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteEntity.DeletedID)

	writer.Render()
}
