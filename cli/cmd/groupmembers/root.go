//go:build cli

package groupmembers

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cli/cmd"
	openlane "github.com/theopenlane/go-client"
)

// cmd represents the base groupMembers command when called without any subcommands
var command = &cobra.Command{
	Use:   "group-members",
	Short: "the subcommands for working with the members of a group",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the output in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the output and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllGroupMemberships:
		var nodes []*graphclient.GetAllGroupMemberships_GroupMemberships_Edges_Node

		for _, i := range v.GroupMemberships.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GeGroupMemberships:
		var nodes []*graphclient.GeGroupMemberships_GroupMemberships_Edges_Node

		for _, i := range v.GroupMemberships.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.CreateGroupMembership:
		e = v.CreateGroupMembership.GroupMembership
	case *graphclient.UpdateGroupMembership:
		e = v.UpdateGroupMembership.GroupMembership
	case *graphclient.DeleteGroupMembership:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlane.GroupMembership

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlane.GroupMembership
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
func tableOutput(out []openlane.GroupMembership) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "GroupID", "UserID", "DisplayName", "FirstName", "LastName", "Email", "Role")
	for _, i := range out {
		writer.AddRow(i.GroupID, i.User.ID, i.User.DisplayName, *i.User.FirstName, *i.User.LastName, i.User.Email, i.Role)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteGroupMembership) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteGroupMembership.DeletedID)

	writer.Render()
}
