//go:build cli

package directoryaccount

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/go-client/graphclient"
)

// command represents the base directory-account command when called without any subcommands
var command = &cobra.Command{
	Use:   "directory-account",
	Short: "the subcommands for working with directory accounts",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	switch v := e.(type) {
	case *graphclient.GetAllDirectoryAccounts:
		var nodes []*graphclient.GetAllDirectoryAccounts_DirectoryAccounts_Edges_Node

		for _, i := range v.DirectoryAccounts.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetDirectoryAccountByID:
		e = v.DirectoryAccount
	case *graphclient.CreateDirectoryAccount:
		e = v.CreateDirectoryAccount.DirectoryAccount
	case *graphclient.UpdateDirectoryAccount:
		e = v.UpdateDirectoryAccount.DirectoryAccount
	case *graphclient.DeleteDirectoryAccount:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.DirectoryAccount

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.DirectoryAccount
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
func tableOutput(out []graphclient.DirectoryAccount) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayID", "ExternalID", "DisplayName", "CanonicalEmail", "GivenName", "FamilyName", "Status", "Department")
	for _, i := range out {
		writer.AddRow(i.ID, i.DisplayID, i.ExternalID, derefStr(i.DisplayName), derefStr(i.CanonicalEmail), derefStr(i.GivenName), derefStr(i.FamilyName), i.Status, derefStr(i.Department))
	}

	writer.Render()
}

// deletedTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteDirectoryAccount) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteDirectoryAccount.DeletedID)

	writer.Render()
}

// derefStr safely dereferences a string pointer
func derefStr(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}
