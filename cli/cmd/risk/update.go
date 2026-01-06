//go:build cli

package risk

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/go-client/graphclient"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
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
	updateCmd.Flags().StringP("details", "d", "", "details of the risk")
	updateCmd.Flags().StringP("status", "s", "", "status of the risk")
	updateCmd.Flags().StringP("business-costs", "b", "", "business costs associated with the risk")
	updateCmd.Flags().StringP("impact", "m", "", "impact of the risk")
	updateCmd.Flags().StringP("likelihood", "l", "", "likelihood of the risk")
	updateCmd.Flags().StringP("mitigation", "g", "", "mitigation for the risk")
	updateCmd.Flags().Int64P("score", "o", 0, "score of the risk")

	updateCmd.Flags().StringSlice("add-programs", []string{}, "add program(s) to the risk")
	updateCmd.Flags().StringSlice("remove-programs", []string{}, "remove program(s) from the risk")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateRiskInput, err error) {
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

	details := cmd.Config.String("details")
	if details != "" {
		input.Details = &details
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = enums.ToRiskStatus(status)
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

	score := cmd.Config.Int64("score")
	if score != 0 {
		input.Score = &score
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
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateRisk(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
