package entitlementplan

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// cmd represents the base plan command when called without any subcommands
var command = &cobra.Command{
	Use:   "plan",
	Short: "the subcommands for working with entitlement plans",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the plans in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the plans in JSON format
	if cmd.OutputFormat == cmd.JSONOutput {
		return jsonOutput(e)
	}

	// check the type of the plans and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllEntitlementPlans:
		var nodes []*openlaneclient.GetAllEntitlementPlans_EntitlementPlans_Edges_Node

		for _, i := range v.EntitlementPlans.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetEntitlementPlanByID:
		e = v.EntitlementPlan
	case *openlaneclient.CreateEntitlementPlan:
		e = v.CreateEntitlementPlan.EntitlementPlan
	case *openlaneclient.UpdateEntitlementPlan:
		e = v.UpdateEntitlementPlan.EntitlementPlan
	case *openlaneclient.DeleteEntitlementPlan:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var planList []openlaneclient.EntitlementPlan

	err = json.Unmarshal(s, &planList)
	if err != nil {
		var plan openlaneclient.EntitlementPlan
		err = json.Unmarshal(s, &plan)
		cobra.CheckErr(err)

		planList = append(planList, plan)
	}

	tableOutput(planList)

	return nil
}

// jsonOutput prints the output in a JSON format
func jsonOutput(out any) error {
	s, err := json.Marshal(out)
	cobra.CheckErr(err)

	return cmd.JSONPrint(s)
}

// tableOutput prints the plans in a table format
func tableOutput(plans []openlaneclient.EntitlementPlan) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Display Name", "Description", "Version")

	for _, p := range plans {
		writer.AddRow(p.ID, p.Name, *p.DisplayName, *p.Description, p.Version)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted plan in a table format
func deletedTableOutput(e *openlaneclient.DeleteEntitlementPlan) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteEntitlementPlan.DeletedID)

	writer.Render()
}
