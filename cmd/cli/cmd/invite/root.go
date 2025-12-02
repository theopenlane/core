//go:build cli

package invite

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

// cmd represents the base invite command when called without any subcommands
var command = &cobra.Command{
	Use:   "invite",
	Short: "the subcommands for working with the invitations of a organization",
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
	case *openlaneclient.GetAllInvites:
		var nodes []*openlaneclient.GetAllInvites_Invites_Edges_Node

		for _, i := range v.Invites.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetInviteByID:
		e = v.Invite
	case *openlaneclient.CreateInvite:
		e = v.CreateInvite.Invite
	case *openlaneclient.DeleteInvite:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.Invite

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.Invite
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
func tableOutput(out []openlaneclient.Invite) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Recipient", "Role", "Status")
	for _, i := range out {
		writer.AddRow(i.ID, i.Recipient, i.Role, i.Status)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteInvite) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteInvite.DeletedID)

	writer.Render()
}
