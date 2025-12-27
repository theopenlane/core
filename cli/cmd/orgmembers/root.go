//go:build cli

package orgmembers

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/go-client/graphclient"
)

// cmd represents the base orgMembers command when called without any subcommands
var command = &cobra.Command{
	Use:   "org-members",
	Short: "the subcommands for working with the members of a organization",
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
	case *graphclient.GetOrgMemberships:
		var nodes []*graphclient.GetOrgMemberships_OrgMemberships_Edges_Node

		for _, i := range v.OrgMemberships.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.CreateOrgMembership:
		e = v.CreateOrgMembership.OrgMembership
	case *graphclient.UpdateOrgMembership:
		e = v.UpdateOrgMembership.OrgMembership
	case *graphclient.DeleteOrgMembership:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.OrgMembership

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.OrgMembership
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
func tableOutput(out []graphclient.OrgMembership) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "UserID", "DisplayName", "FirstName", "LastName", "Email", "Role")

	for _, i := range out {
		userID := i.UserID
		if userID == "" && i.User != nil {
			userID = i.User.ID
		}

		var (
			displayName string
			firstName   string
			lastName    string
			email       string
		)

		if i.User != nil {
			displayName = i.User.DisplayName
			firstName = *i.User.FirstName
			lastName = *i.User.LastName
			email = i.User.Email
		}

		writer.AddRow(userID, displayName, firstName, lastName, email, i.Role)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteOrgMembership) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteOrgMembership.DeletedID)

	writer.Render()
}
