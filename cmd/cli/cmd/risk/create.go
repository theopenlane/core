package risk

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new risk",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("name", "n", "", "name of the risk")
	createCmd.Flags().StringP("description", "d", "", "description of the risk")
	createCmd.Flags().StringP("status", "s", "", "status of the risk")
	createCmd.Flags().StringSliceP("programs", "p", []string{}, "program ID(s) associated with the risk")
	createCmd.Flags().StringP("type", "t", "", "type of the risk")
	createCmd.Flags().StringP("business-costs", "b", "", "business costs associated with the risk")
	createCmd.Flags().StringP("impact", "m", "", "impact of the risk")
	createCmd.Flags().StringP("likelihood", "l", "", "likelihood of the risk")
	createCmd.Flags().StringP("mitigation", "g", "", "mitigation for the risk")
	createCmd.Flags().StringP("satisfies", "a", "", "which controls are satisfied by the risk")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateRiskInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	input.ProgramIDs = cmd.Config.Strings("programs")

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

	return input, nil
}

// create a new risk
func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateRisk(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
