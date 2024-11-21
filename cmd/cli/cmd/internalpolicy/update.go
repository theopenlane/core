package internalpolicy

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing internal policy",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "policy id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("name", "n", "", "name of the procedure")
	updateCmd.Flags().StringP("description", "d", "", "description of the procedure")
	updateCmd.Flags().StringP("status", "s", "", "status of the procedure")
	updateCmd.Flags().StringP("type", "t", "", "type of the procedure")
	updateCmd.Flags().StringP("version", "v", "v0.1", "version of the procedure")
	updateCmd.Flags().StringP("purpose", "p", "", "purpose and scope of the procedure")
	updateCmd.Flags().StringP("background", "b", "", "background information of the procedure")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateInternalPolicyInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("internal policy id")
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

	policyType := cmd.Config.String("type")
	if policyType != "" {
		input.PolicyType = &policyType
	}

	version := cmd.Config.String("version")
	if version != "" {
		input.Version = &version
	}

	purpose := cmd.Config.String("purpose")
	if purpose != "" {
		input.PurposeAndScope = &purpose
	}

	background := cmd.Config.String("background")
	if background != "" {
		input.Background = &background
	}

	return id, input, nil
}

// update an existing internal policy in the platform
func update(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateInternalPolicy(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
