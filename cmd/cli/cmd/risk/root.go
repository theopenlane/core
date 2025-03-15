package risk

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// command represents the base risk command when called without any subcommands
var command = &cobra.Command{
	Use:   "risk",
	Short: "the subcommands for working with risks",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the risks in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the risks and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllRisks:
		var nodes []*openlaneclient.GetAllRisks_Risks_Edges_Node

		for _, i := range v.Risks.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetRisks:
		var nodes []*openlaneclient.GetRisks_Risks_Edges_Node

		for _, i := range v.Risks.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetRiskByID:
		e = v.Risk
	case *openlaneclient.CreateRisk:
		e = v.CreateRisk.Risk
	case *openlaneclient.UpdateRisk:
		e = v.UpdateRisk.Risk
	case *openlaneclient.DeleteRisk:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.Risk

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.Risk
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
func tableOutput(out []openlaneclient.Risk) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayID", "Name", "Details", "Status", "Type", "BusinessCosts", "Impact", "Likelihood", "Score", "Mitigation", "Program(s)")

	for _, i := range out {
		programs := []string{}

		for _, p := range i.Programs.Edges {
			programs = append(programs, p.Node.Name)
		}

		writer.AddRow(i.ID, i.DisplayID, i.Name, *i.Details, *i.Status, *i.RiskType, *i.BusinessCosts, *i.Impact, *i.Likelihood, *i.Score, *i.Mitigation, strings.Join(programs, ","))
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteRisk) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteRisk.DeletedID)

	writer.Render()
}
