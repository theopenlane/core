//go:build cli

package trustcenternda

import (
	"context"
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "submit a new trust center NDA",
	Run: func(cmd *cobra.Command, args []string) {
		err := submit(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(submitCmd)

	// command line flags for the submit command
	submitCmd.Flags().StringP("template-id", "t", "", "template id for the NDA")
	submitCmd.Flags().StringP("response", "r", "", "response to the NDA")
}

// submitValidation validates the required fields for the command
func submitValidation() (input graphclient.SubmitTrustCenterNDAResponseInput, err error) {
	templateID := cmd.Config.String("template-id")
	if templateID == "" {
		return input, cmd.NewRequiredFieldMissingError("template id")
	}

	responseFile := cmd.Config.String("response")
	if responseFile == "" {
		return input, cmd.NewRequiredFieldMissingError("response")
	}

	responseStr, err := os.ReadFile(responseFile)
	if err != nil {
		return input, err
	}

	var responseJSON map[string]any
	if err = json.Unmarshal(responseStr, &responseJSON); err != nil {
		return input, err
	}

	input.TemplateID = templateID
	input.Response = responseJSON

	return input, nil
}

// submit a new trust center NDA
func submit(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := submitValidation()
	cobra.CheckErr(err)

	o, err := client.SubmitTrustCenterNDAResponse(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
