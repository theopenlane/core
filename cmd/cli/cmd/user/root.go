package user

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/cli/tables"
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
	if cmd.OutputFormat == cmd.JSONOutput {
		return jsonOutput(e)
	}

	// check the type of the output and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllUsers:
		var nodes []*openlaneclient.GetAllUsers_Users_Edges_Node

		for _, i := range v.Users.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetUserByID:
		e = v.User
	case *openlaneclient.CreateUser:
		e = v.CreateUser.User
	case *openlaneclient.UpdateUser:
		e = v.UpdateUser.User
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.User

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.User
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
func tableOutput(out []openlaneclient.User) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Email", "FirstName", "LastName", "DisplayName", "AuthProvider")
	for _, i := range out {
		// this doesn't visually show you the json in the table but leaving it in for now
		writer.AddRow(i.ID, i.Email, *i.FirstName, *i.LastName, i.DisplayName, i.AuthProvider)
	}

	writer.Render()
}
