package internalpolicy

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new internal policy",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("name", "n", "", "name of the policy")
	createCmd.Flags().StringP("details", "d", "", "details of the policy")
	createCmd.Flags().StringP("status", "s", "", "status of the policy e.g. draft, published, archived, etc.")
	createCmd.Flags().StringP("type", "t", "", "type of the policy")
	createCmd.Flags().StringP("revision", "v", models.DefaultRevision, "revision of the policy")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateInternalPolicyInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = enums.ToDocumentStatus(status)
	}

	policyType := cmd.Config.String("type")
	if policyType != "" {
		input.PolicyType = &policyType
	}

	revision := cmd.Config.String("revision")
	if revision != "" {
		input.Revision = &revision
	}

	details := cmd.Config.String("details")
	if details != "" {
		input.Details = &details
	}

	return input, nil
}

// create a new internal policy
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

	o, err := client.CreateInternalPolicy(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
