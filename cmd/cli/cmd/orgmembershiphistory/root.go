//go:build cli

package orgmembershiphistory

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
)

// cmd represents the base orgMembershipHistory command when called without any subcommands
var command = &cobra.Command{
	Use:   "org-membership-history",
	Short: "the subcommands for working with orgMembershipHistories",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the orgMembershipHistories in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the orgMembershipHistories and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllOrgMembershipHistories:
		var nodes []*openlaneclient.GetAllOrgMembershipHistories_OrgMembershipHistories_Edges_Node

		for _, i := range v.OrgMembershipHistories.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetOrgMembershipHistories:
		var nodes []*openlaneclient.GetOrgMembershipHistories_OrgMembershipHistories_Edges_Node

		for _, i := range v.OrgMembershipHistories.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.OrgMembershipHistory

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.OrgMembershipHistory
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
func tableOutput(out []openlaneclient.OrgMembershipHistory) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Ref", "Operation", "UpdatedAt", "UpdatedBy")
	for _, i := range out {
		writer.AddRow(i.ID, *i.Ref, i.Operation, *i.UpdatedAt, *i.UpdatedBy)
	}

	writer.Render()
}
