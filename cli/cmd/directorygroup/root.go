//go:build cli

package directorygroup

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/go-client/graphclient"
)

// command represents the base directory-group command when called without any subcommands
var command = &cobra.Command{
	Use:   "directory-group",
	Short: "the subcommands for working with directory groups",
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
	case *graphclient.GetAllDirectoryGroups:
		var nodes []*graphclient.GetAllDirectoryGroups_DirectoryGroups_Edges_Node

		for _, i := range v.DirectoryGroups.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetDirectoryGroupByID:
		e = v.DirectoryGroup
	case *graphclient.CreateDirectoryGroup:
		e = v.CreateDirectoryGroup.DirectoryGroup
	case *graphclient.UpdateDirectoryGroup:
		e = v.UpdateDirectoryGroup.DirectoryGroup
	case *graphclient.DeleteDirectoryGroup:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.DirectoryGroup

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.DirectoryGroup
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
func tableOutput(out []graphclient.DirectoryGroup) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayID", "ExternalID", "DisplayName", "Email", "Classification", "Status", "MemberCount", "Description")
	for _, i := range out {
		writer.AddRow(i.ID, i.DisplayID, i.ExternalID, derefStr(i.DisplayName), derefStr(i.Email), i.Classification, i.Status, derefInt64(i.MemberCount), derefStr(i.Description))
	}

	writer.Render()
}

// deletedTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteDirectoryGroup) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteDirectoryGroup.DeletedID)

	writer.Render()
}

// derefStr safely dereferences a string pointer
func derefStr(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

// derefInt64 safely dereferences an int64 pointer
func derefInt64(i *int64) int64 {
	if i == nil {
		return 0
	}

	return *i
}
