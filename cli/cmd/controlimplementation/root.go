//go:build cli

package controlimplementation

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
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
	case *graphclient.GetAllControlImplementations:
		var nodes []*graphclient.GetAllControlImplementations_ControlImplementations_Edges_Node

		for _, i := range v.ControlImplementations.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetControlImplementations:
		var nodes []*graphclient.GetControlImplementations_ControlImplementations_Edges_Node

		for _, i := range v.ControlImplementations.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetControlImplementationByID:
		e = v.ControlImplementation
	case *graphclient.CreateControlImplementation:
		e = v.CreateControlImplementation.ControlImplementation
	case *graphclient.UpdateControlImplementation:
		e = v.UpdateControlImplementation.ControlImplementation
	case *graphclient.DeleteControlImplementation:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.ControlImplementation

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.ControlImplementation
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
func tableOutput(out []graphclient.ControlImplementation) {
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
func deletedTableOutput(e *graphclient.DeleteControlImplementation) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteControlImplementation.DeletedID)

	writer.Render()
}
