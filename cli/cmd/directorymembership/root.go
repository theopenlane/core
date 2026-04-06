//go:build cli

package directorymembership

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/go-client/graphclient"
)

// command represents the base directory-membership command when called without any subcommands
var command = &cobra.Command{
	Use:   "directory-membership",
	Short: "the subcommands for working with directory memberships",
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
	case *graphclient.GetAllDirectoryMemberships:
		var nodes []*graphclient.GetAllDirectoryMemberships_DirectoryMemberships_Edges_Node

		for _, i := range v.DirectoryMemberships.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetDirectoryMembershipByID:
		e = v.DirectoryMembership
	case *graphclient.CreateDirectoryMembership:
		e = v.CreateDirectoryMembership.DirectoryMembership
	case *graphclient.UpdateDirectoryMembership:
		e = v.UpdateDirectoryMembership.DirectoryMembership
	case *graphclient.DeleteDirectoryMembership:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.DirectoryMembership

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.DirectoryMembership
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
func tableOutput(out []graphclient.DirectoryMembership) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayID", "DirectoryAccountID", "DirectoryGroupID", "Role", "Source", "IntegrationID")
	for _, i := range out {
		writer.AddRow(i.ID, i.DisplayID, i.DirectoryAccountID, i.DirectoryGroupID, derefRole(i.Role), derefStr(i.Source), i.IntegrationID)
	}

	writer.Render()
}

// deletedTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteDirectoryMembership) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteDirectoryMembership.DeletedID)

	writer.Render()
}

// derefStr safely dereferences a string pointer
func derefStr(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

// derefRole safely dereferences a DirectoryMembershipRole pointer
func derefRole(r *enums.DirectoryMembershipRole) string {
	if r == nil {
		return ""
	}

	return string(*r)
}
