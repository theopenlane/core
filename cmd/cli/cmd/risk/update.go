package risk

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing risk",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "risk id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("name", "n", "", "name of the risk")
	updateCmd.Flags().StringP("description", "d", "", "description of the risk")
	updateCmd.Flags().StringP("status", "s", "", "status of the risk")
	updateCmd.Flags().StringP("type", "t", "", "type of the risk")
	updateCmd.Flags().StringP("business-costs", "b", "", "business costs associated with the risk")
	updateCmd.Flags().StringP("impact", "m", "", "impact of the risk")
	updateCmd.Flags().StringP("likelihood", "l", "", "likelihood of the risk")
	updateCmd.Flags().StringP("mitigation", "g", "", "mitigation for the risk")
	updateCmd.Flags().StringP("satisfies", "a", "", "which controls are satisfied by the risk")
	updateCmd.Flags().StringSlice("add-programs", []string{}, "add program(s) to the risk")
	updateCmd.Flags().StringSlice("remove-programs", []string{}, "remove program(s) from the risk")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateRiskInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("risk id")
	}

	// validation of required fields for the update command
	// output the input struct with the required fields and optional fields based on the command line flags
	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = &status
	}

	riskType := cmd.Config.String("type")
	if riskType != "" {
		input.RiskType = &riskType
	}

	businessCosts := cmd.Config.String("business-costs")
	if businessCosts != "" {
		input.BusinessCosts = &businessCosts
	}

	impact := cmd.Config.String("impact")
	if impact != "" {
		input.Impact = enums.ToRiskImpact(impact)
	}

	likelihood := cmd.Config.String("likelihood")
	if likelihood != "" {
		input.Likelihood = enums.ToRiskLikelihood(likelihood)
	}

	mitigation := cmd.Config.String("mitigation")
	if mitigation != "" {
		input.Mitigation = &mitigation
	}

	satisfies := cmd.Config.String("satisfies")
	if satisfies != "" {
		input.Satisfies = &satisfies
	}

	addPrograms := cmd.Config.Strings("add-programs")
	if len(addPrograms) > 0 {
		input.AddProgramIDs = addPrograms
	}

	removePrograms := cmd.Config.Strings("remove-programs")
	if len(removePrograms) > 0 {
		input.RemoveProgramIDs = removePrograms
	}

	return id, input, nil
}

// update an existing risk in the platform
func update(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateRisk(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
