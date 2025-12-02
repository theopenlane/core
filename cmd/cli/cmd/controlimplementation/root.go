//go:build cli

package controlimplementation

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base controlImplementation command when called without any subcommands
var command = &cobra.Command{
	Use:     "control-implementation",
	Aliases: []string{"controlimplementation", "ci"},
	Short:   "the subcommands for working with controlImplementations",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the controlImplementations in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the controlImplementations and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllControlImplementations:
		var nodes []*openlaneclient.GetAllControlImplementations_ControlImplementations_Edges_Node

		for _, i := range v.ControlImplementations.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetControlImplementations:
		var nodes []*openlaneclient.GetControlImplementations_ControlImplementations_Edges_Node

		for _, i := range v.ControlImplementations.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetControlImplementationByID:
		e = v.ControlImplementation
	case *openlaneclient.CreateControlImplementation:
		e = v.CreateControlImplementation.ControlImplementation
	case *openlaneclient.UpdateControlImplementation:
		e = v.UpdateControlImplementation.ControlImplementation
	case *openlaneclient.DeleteControlImplementation:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.ControlImplementation

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.ControlImplementation
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
func tableOutput(out []openlaneclient.ControlImplementation) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Details", "Status", "ImplementationDate", "Verified", "VerificationDate", "Controls", "Subcontrols")
	for _, i := range out {
		controlRefCodes := []string{}
		if i.Controls != nil {
			for _, control := range i.Controls.Edges {
				controlRefCodes = append(controlRefCodes, control.Node.RefCode)
			}
		}

		subcontrolRefCodes := []string{}
		if i.Subcontrols != nil {
			for _, subcontrol := range i.Subcontrols.Edges {
				subcontrolRefCodes = append(subcontrolRefCodes, subcontrol.Node.RefCode)
			}
		}

		writer.AddRow(i.ID, *i.Details, i.Status, i.ImplementationDate, *i.Verified, i.VerificationDate, strings.Join(controlRefCodes, ","), strings.Join(subcontrolRefCodes, ","))
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteControlImplementation) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteControlImplementation.DeletedID)

	writer.Render()
}
