package orgmembers

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
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
	case *openlaneclient.GetOrgMembersByOrgID:
		var nodes []*openlaneclient.GetOrgMembersByOrgID_OrgMemberships_Edges_Node

		for _, i := range v.OrgMemberships.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.AddUserToOrgWithRole:
		e = v.CreateOrgMembership.OrgMembership
	case *openlaneclient.UpdateUserRoleInOrg:
		e = v.UpdateOrgMembership.OrgMembership
	case *openlaneclient.RemoveUserFromOrg:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.OrgMembership

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.OrgMembership
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
func tableOutput(out []openlaneclient.OrgMembership) {
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
func deletedTableOutput(e *openlaneclient.RemoveUserFromOrg) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteOrgMembership.DeletedID)

	writer.Render()
}
