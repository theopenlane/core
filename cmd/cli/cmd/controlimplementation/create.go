//go:build cli

package controlimplementation

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	openlaneclient "github.com/theopenlane/go-client"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new controlImplementation",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("details", "d", "", "details of the control implementation")
	createCmd.Flags().StringP("status", "s", "", "status of the control implementation")
	createCmd.Flags().StringP("implementation-date", "m", "", "date the control was implemented")
	createCmd.Flags().StringP("verification-date", "v", "", "date the control was verified")
	createCmd.Flags().BoolP("verified", "r", false, "set to true if the control implementation has been verified")
	createCmd.Flags().StringSlice("control-ids", []string{}, "ID of the related control")
	createCmd.Flags().StringSlice("subcontrol-ids", []string{}, "ID of the related subcontrol")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateControlImplementationInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	details := cmd.Config.String("details")
	if details != "" {
		input.Details = &details
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = enums.ToDocumentStatus(status)
	}

	implementationDate := cmd.Config.String("implementation-date")
	if implementationDate != "" {
		date, err := time.Parse(time.RFC3339, implementationDate)
		cobra.CheckErr(err)

		input.ImplementationDate = &date
	}

	verificationDate := cmd.Config.String("verification-date")
	if verificationDate != "" {
		date, err := time.Parse(time.RFC3339, verificationDate)
		cobra.CheckErr(err)

		input.VerificationDate = &date
	}

	verified := cmd.Config.Bool("verified")
	if verified {
		input.Verified = &verified
	}

	controlIDs := cmd.Config.Strings("control-ids")
	if len(controlIDs) > 0 {
		input.ControlIDs = controlIDs
	}

	subcontrolIDs := cmd.Config.Strings("subcontrol-ids")
	if len(subcontrolIDs) > 0 {
		input.SubcontrolIDs = subcontrolIDs
	}

	return input, nil
}

// create a new controlImplementation
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

	o, err := client.CreateControlImplementation(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
