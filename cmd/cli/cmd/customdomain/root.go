//go:build cli

package customdomain

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
)

// command represents the base customdomain command when called without any subcommands
var command = &cobra.Command{
	Use:     "custom-domain",
	Aliases: []string{"customdomain"},
	Short:   "the subcommands for working with custom domains",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the custom domains in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the custom domains and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.UpdateCustomDomain:
		nodes := []*openlaneclient.UpdateCustomDomain_UpdateCustomDomain_CustomDomain{
			&v.UpdateCustomDomain.CustomDomain,
		}

		e = nodes

	case *openlaneclient.CreateCustomDomain:
		nodes := []*openlaneclient.CreateCustomDomain_CreateCustomDomain_CustomDomain{
			&v.CreateCustomDomain.CustomDomain,
		}

		e = nodes
	case *openlaneclient.GetCustomDomainByID:
		nodes := []*openlaneclient.GetCustomDomainByID_CustomDomain{
			&v.CustomDomain,
		}

		e = nodes
	case *openlaneclient.GetAllCustomDomains:
		var nodes []*openlaneclient.GetAllCustomDomains_CustomDomains_Edges_Node

		for _, i := range v.CustomDomains.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetCustomDomains:
		var nodes []*openlaneclient.GetCustomDomains_CustomDomains_Edges_Node
		for _, i := range v.CustomDomains.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.DeleteCustomDomain:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.CustomDomain

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.CustomDomain
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
func tableOutput(out []openlaneclient.CustomDomain) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "CNAME", "Verification ID", "Created At")
	for _, i := range out {
		verificationID := ""
		if i.DNSVerificationID != nil {
			verificationID = *i.DNSVerificationID
		}

		writer.AddRow(i.ID, i.CnameRecord, verificationID, i.CreatedAt)
	}

	writer.Render()
}

// deletedTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteCustomDomain) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteCustomDomain.DeletedID)

	writer.Render()
}
