//go:build cli

package trustcentercompliance

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new trust center compliance",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("standard-id", "s", "", "standard id for the compliance (required)")
	createCmd.Flags().StringP("trust-center-id", "t", "", "trust center id for the compliance")
	createCmd.Flags().StringSliceP("tags", "", []string{}, "tags associated with the trust center compliance")
}

// createValidation validates the required fields for the command
func createValidation() (*openlaneclient.CreateTrustCenterComplianceInput, error) {
	standardID := cmd.Config.String("standard-id")
	if standardID == "" {
		return nil, cmd.NewRequiredFieldMissingError("standard id")
	}

	input := &openlaneclient.CreateTrustCenterComplianceInput{
		StandardID: standardID,
	}

	trustCenterID := cmd.Config.String("trust-center-id")
	if trustCenterID != "" {
		input.TrustCenterID = &trustCenterID
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	return input, nil
}

// create a new trust center compliance
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

	o, err := client.CreateTrustCenterCompliance(ctx, *input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
