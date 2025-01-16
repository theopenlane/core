package group

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// cmd represents the base group command when called without any subcommands
var command = &cobra.Command{
	Use:   "group",
	Short: "the subcommands for working with groups",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the groups in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the groups and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllGroups:
		var nodes []*openlaneclient.GetAllGroups_Groups_Edges_Node

		for _, i := range v.Groups.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetGroupByID:
		e = v.Group
	case *openlaneclient.CreateGroup:
		e = v.CreateGroup.Group
	case *openlaneclient.UpdateGroup:
		e = v.UpdateGroup.Group
	case *openlaneclient.DeleteGroup:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.Group

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.Group
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
func tableOutput(out []openlaneclient.Group) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Display Name", "Description", "Visibility", "Managed", "Organization", "Members")
	for _, i := range out {
		isManaged := false
		if i.IsManaged != nil {
			isManaged = *i.IsManaged
		}

		writer.AddRow(i.ID, i.Name, i.DisplayName, *i.Description, i.Setting.Visibility, isManaged, i.Owner.DisplayName, len(i.Members))
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteGroup) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteGroup.DeletedID)

	writer.Render()
}
