//go:build cli

package trustcentercontrol

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new trust center control",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("control-id", "c", "", "control id to associate with the trust center")
	createCmd.Flags().StringP("trust-center-id", "t", "", "trust center id to associate with the control")
	createCmd.Flags().StringSliceP("tags", "", []string{}, "tags associated with the trust center control")
}

// createValidation validates the required fields for the command
func createValidation() (openlaneclient.CreateTrustCenterControlInput, error) {
	controlID := cmd.Config.String("control-id")
	if controlID == "" {
		return openlaneclient.CreateTrustCenterControlInput{}, cmd.NewRequiredFieldMissingError("control-id")
	}

	input := openlaneclient.CreateTrustCenterControlInput{
		ControlID: controlID,
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

// create a new trust center control
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

	o, err := client.CreateTrustCenterControl(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
