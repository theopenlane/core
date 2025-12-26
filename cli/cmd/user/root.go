//go:build cli

package user

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/go-client/graphclient"
)

// cmd represents the base user command when called without any subcommands
var command = &cobra.Command{
	Use:   "user",
	Short: "the subcommands for working with the user",
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
	case *graphclient.GetAllUsers:
		var nodes []*graphclient.GetAllUsers_Users_Edges_Node

		for _, i := range v.Users.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetUserByID:
		e = v.User
	case *graphclient.GetSelf:
		e = v.Self
	case *graphclient.UpdateUser:
		e = v.UpdateUser.User
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.User

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.User
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
func tableOutput(out []graphclient.User) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Email", "FirstName", "LastName", "DisplayName", "AuthProvider")
	for _, i := range out {
		firstName := ""
		lastName := ""

		if i.FirstName != nil {
			firstName = *i.FirstName
		}

		if i.LastName != nil {
			lastName = *i.LastName
		}

		writer.AddRow(i.ID, i.Email,
			firstName, lastName,
			i.DisplayName, i.AuthProvider)
	}

	writer.Render()
}
