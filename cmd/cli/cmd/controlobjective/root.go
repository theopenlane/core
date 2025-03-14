package controlobjective

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// command represents the base controlObjective command when called without any subcommands
var command = &cobra.Command{
	Use:   "control-objective",
	Short: "the subcommands for working with controlObjectives",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the controlObjectives in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the controlObjectives and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllControlObjectives:
		var nodes []*openlaneclient.GetAllControlObjectives_ControlObjectives_Edges_Node

		for _, i := range v.ControlObjectives.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetControlObjectives:
		var nodes []*openlaneclient.GetControlObjectives_ControlObjectives_Edges_Node

		for _, i := range v.ControlObjectives.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetControlObjectiveByID:
		e = v.ControlObjective
	case *openlaneclient.CreateControlObjective:
		e = v.CreateControlObjective.ControlObjective
	case *openlaneclient.UpdateControlObjective:
		e = v.UpdateControlObjective.ControlObjective
	case *openlaneclient.DeleteControlObjective:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.ControlObjective

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.ControlObjective
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
func tableOutput(out []openlaneclient.ControlObjective) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "DesiredOutcome", "Source", "Revision", "Status", "ControlObjectiveType", "Controls", "Programs")

	for _, i := range out {
		programs := []string{}

		for _, p := range i.Programs {
			programs = append(programs, p.Name)
		}

		controls := []string{}
		for _, c := range i.Controls {
			controls = append(controls, c.RefCode)
		}

		writer.AddRow(i.ID, i.Name, *i.DesiredOutcome, *i.Source, *i.Revision, *i.Status, *i.ControlObjectiveType, strings.Join(controls, ", "), strings.Join(programs, ", "))
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteControlObjective) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteControlObjective.DeletedID)

	writer.Render()
}
