//go:build cli

package trustcenternda

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var sendEmailCmd = &cobra.Command{
	Use:   "send-email",
	Short: "Sends email for trust center NDA request",
	Run: func(cmd *cobra.Command, args []string) {
		err := sendNDAEmail(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(sendEmailCmd)

	// command line flags for the send email command
	sendEmailCmd.Flags().StringP("trust-center-id", "t", "", "trust center id for the NDA")
	sendEmailCmd.Flags().StringP("email", "e", "", "email to send the NDA to")
}

func sendEmailValidation() (graphclient.SendTrustCenterNDAInput, error) {
	input := graphclient.SendTrustCenterNDAInput{}

	trustCenterID := cmd.Config.String("trust-center-id")
	if trustCenterID != "" {
		input.TrustCenterID = trustCenterID
	} else {
		return input, cmd.NewRequiredFieldMissingError("trust center id")
	}

	email := cmd.Config.String("email")
	if email != "" {
		input.Email = email
	} else {
		return input, cmd.NewRequiredFieldMissingError("email")
	}

	return input, nil
}

// create a new trust center NDA
func sendNDAEmail(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := sendEmailValidation()
	cobra.CheckErr(err)

	o, err := client.SendTrustCenterNDAEmail(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
