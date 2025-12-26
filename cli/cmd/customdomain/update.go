//go:build cli

package customdomain

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing custom domain",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "custom domain ID to update")
	updateCmd.Flags().StringP("dns-verification-id", "d", "", "DNS verification ID to associate with the custom domain")
	updateCmd.Flags().Bool("clear-dns-verification", false, "clear the DNS verification association")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateCustomDomainInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("custom domain id")
	}

	// Build the update input with optional fields
	if dnsVerificationID := cmd.Config.String("dns-verification-id"); dnsVerificationID != "" {
		input.DNSVerificationID = &dnsVerificationID
	}

	if cmd.Config.Bool("clear-dns-verification") {
		clearDNS := true
		input.ClearDNSVerification = &clearDNS
	}

	return id, input, nil
}

// update an existing custom domain in the platform
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

	o, err := client.UpdateCustomDomain(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
