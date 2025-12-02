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

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing controlImplementation",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "controlImplementation id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("details", "d", "", "details of the control implementation")
	updateCmd.Flags().StringP("status", "s", "", "status of the control implementation")
	updateCmd.Flags().StringP("implementation-date", "m", "", "date the control was implemented")
	updateCmd.Flags().StringP("verification-date", "v", "", "date the control was verified")
	updateCmd.Flags().BoolP("verified", "r", false, "set to true if the control implementation has been verified")
	updateCmd.Flags().StringSlice("add-control-ids", []string{}, "ID of the related control to add")
	updateCmd.Flags().StringSlice("add-subcontrol-ids", []string{}, "ID of the related subcontrol to add")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateControlImplementationInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("controlImplementation id")
	}

	// validation of required fields for the update command
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
		input.AddControlIDs = controlIDs
	}

	subcontrolIDs := cmd.Config.Strings("subcontrol-ids")
	if len(subcontrolIDs) > 0 {
		input.AddSubcontrolIDs = subcontrolIDs
	}

	return id, input, nil
}

// update an existing controlImplementation in the platform
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

	o, err := client.UpdateControlImplementation(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
