//go:build cli

package dnsverification

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	openlaneclient "github.com/theopenlane/go-client"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing DNS verification record",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "DNS verification ID to update")
	updateCmd.Flags().StringP("dns-verification-status", "d", "", "DNS verification status (PENDING, ACTIVE, FAILED)")
	updateCmd.Flags().StringP("dns-verification-status-reason", "r", "", "DNS verification status reason")
	updateCmd.Flags().StringP("acme-challenge-status", "s", "", "SSL certificate status (PENDING, ACTIVE, FAILED)")
	updateCmd.Flags().StringP("acme-challenge-status-reason", "c", "", "SSL certificate status reason")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateDNSVerificationInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("DNS verification id")
	}

	// Build the update input with optional fields
	if dnsStatus := cmd.Config.String("dns-verification-status"); dnsStatus != "" {
		status := enums.DNSVerificationStatus(dnsStatus)
		input.DNSVerificationStatus = &status
	}

	if dnsStatusReason := cmd.Config.String("dns-verification-status-reason"); dnsStatusReason != "" {
		input.DNSVerificationStatusReason = &dnsStatusReason
	}

	if sslStatus := cmd.Config.String("acme-challenge-status"); sslStatus != "" {
		status := enums.SSLVerificationStatus(sslStatus)
		input.AcmeChallengeStatus = &status
	}

	if sslStatusReason := cmd.Config.String("acme-challenge-status-reason"); sslStatusReason != "" {
		input.AcmeChallengeStatusReason = &sslStatusReason
	}

	return id, input, nil
}

// update an existing DNS verification record in the platform
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

	o, err := client.UpdateDNSVerification(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
