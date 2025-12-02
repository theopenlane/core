//go:build cli

package programmembers

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

// cmd represents the base programMembers command when called without any subcommands
var command = &cobra.Command{
	Use:   "program-members",
	Short: "the subcommands for working with the members of a program",
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
	case *openlaneclient.GetProgramMemberships:
		var nodes []*openlaneclient.GetProgramMemberships_ProgramMemberships_Edges_Node

		for _, i := range v.ProgramMemberships.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetAllProgramMemberships:
		var nodes []*openlaneclient.GetAllProgramMemberships_ProgramMemberships_Edges_Node

		for _, i := range v.ProgramMemberships.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.CreateProgramMembership:
		e = v.CreateProgramMembership.ProgramMembership
	case *openlaneclient.UpdateProgramMembership:
		e = v.UpdateProgramMembership.ProgramMembership
	case *openlaneclient.DeleteProgramMembership:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.ProgramMembership

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.ProgramMembership
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
func tableOutput(out []openlaneclient.ProgramMembership) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ProgramID", "UserID", "DisplayName", "FirstName", "LastName", "Email", "Role")
	for _, i := range out {
		firstName := ""

		if i.User == nil {
			continue
		}

		if i.User.FirstName != nil {
			firstName = *i.User.FirstName
		}
		lastName := ""
		if i.User.LastName != nil {
			lastName = *i.User.LastName
		}

		writer.AddRow(
			i.ProgramID,
			i.User.ID,
			i.User.DisplayName,
			firstName,
			lastName,
			i.User.Email,
			i.Role)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteProgramMembership) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteProgramMembership.DeletedID)

	writer.Render()
}
