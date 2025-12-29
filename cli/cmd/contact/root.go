//go:build cli

package contact

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/go-client/graphclient"
)

// cmd represents the base contact command when called without any subcommands
var command = &cobra.Command{
	Use:   "contact",
	Short: "the subcommands for working with contacts",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the contacts in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the contacts and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllContacts:
		var nodes []*graphclient.GetAllContacts_Contacts_Edges_Node

		for _, i := range v.Contacts.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetContacts:
		var nodes []*graphclient.GetContacts_Contacts_Edges_Node

		for _, i := range v.Contacts.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetContactByID:
		e = v.Contact
	case *graphclient.CreateContact:
		e = v.CreateContact.Contact
	case *graphclient.UpdateContact:
		e = v.UpdateContact.Contact
	case *graphclient.DeleteContact:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.Contact

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.Contact
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
func tableOutput(out []graphclient.Contact) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Status", "Company", "Title", "Email", "PhoneNumber", "Address")
	for _, i := range out {
		writer.AddRow(i.ID, i.FullName, i.Status.String(), *i.Company, *i.Title, *i.Email, *i.PhoneNumber, *i.Address)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteContact) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteContact.DeletedID)

	writer.Render()
}
