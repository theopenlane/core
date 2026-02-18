//go:build cli

package identityHolder

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base identityHolder command when called without any subcommands
var command = &cobra.Command{
	Use:   "identityHolder",
	Short: "the subcommands for working with IdentityHolders",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the IdentityHolders in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the IdentityHolders and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllIdentityHolders:
		var nodes []*graphclient.GetAllIdentityHolders_IdentityHolders_Edges_Node

		for _, i := range v.IdentityHolders.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetIdentityHolders:
		var nodes []*graphclient.GetIdentityHolders_IdentityHolders_Edges_Node

		for _, i := range v.IdentityHolders.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetIdentityHolderByID:
		e = v.IdentityHolder
	case *graphclient.CreateIdentityHolder:
		e = v.CreateIdentityHolder.IdentityHolder
	case *graphclient.UpdateIdentityHolder:
		e = v.UpdateIdentityHolder.IdentityHolder
	case *graphclient.DeleteIdentityHolder:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.IdentityHolder

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.IdentityHolder
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
func tableOutput(out []graphclient.IdentityHolder) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayID", "FullName", "Email", "Type", "Active", "Campaigns")
	for _, i := range out {
		campaignCount := 0
		if i.Campaigns != nil {
			campaignCount = len(i.Campaigns.Edges)
		}

		writer.AddRow(i.ID, i.DisplayID, i.FullName, i.Email, i.IdentityHolderType, i.IsActive, campaignCount)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteIdentityHolder) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteIdentityHolder.DeletedID)

	writer.Render()
}
