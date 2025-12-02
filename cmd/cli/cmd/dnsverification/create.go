//go:build cli

package dnsverification

import (
	"context"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new DNS verification record",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("cloudflare-hostname-id", "c", "", "Cloudflare hostname ID")
	createCmd.Flags().StringP("dns-txt-record", "d", "", "DNS TXT record name")
	createCmd.Flags().StringP("dns-txt-value", "v", "", "DNS TXT record value")
	createCmd.Flags().StringP("acme-challenge-path", "s", "", "SSL TXT record name")
	createCmd.Flags().StringP("acme-challenge-value", "t", "", "SSL TXT record value")
	createCmd.Flags().StringP("org-id", "o", "", "Organization ID to associate the DNS verification with")
}

func createValidation() error {
	cloudflareHostnameID := cmd.Config.String("cloudflare-hostname-id")
	if cloudflareHostnameID == "" {
		return cmd.NewRequiredFieldMissingError("cloudflare-hostname-id")
	}

	dnsTxtRecord := cmd.Config.String("dns-txt-record")
	if dnsTxtRecord == "" {
		return cmd.NewRequiredFieldMissingError("dns-txt-record")
	}

	dnsTxtValue := cmd.Config.String("dns-txt-value")
	if dnsTxtValue == "" {
		return cmd.NewRequiredFieldMissingError("dns-txt-value")
	}

	sslTxtRecord := cmd.Config.String("acme-challenge-path")
	if sslTxtRecord == "" {
		return cmd.NewRequiredFieldMissingError("acme-challenge-path")
	}

	sslTxtValue := cmd.Config.String("acme-challenge-value")
	if sslTxtValue == "" {
		return cmd.NewRequiredFieldMissingError("acme-challenge-value")
	}

	return nil
}

// create a new DNS verification record
func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	cobra.CheckErr(createValidation())

	input := openlaneclient.CreateDNSVerificationInput{
		CloudflareHostnameID:       cmd.Config.String("cloudflare-hostname-id"),
		DNSTxtRecord:               cmd.Config.String("dns-txt-record"),
		DNSTxtValue:                cmd.Config.String("dns-txt-value"),
		AcmeChallengePath:          lo.ToPtr(cmd.Config.String("acme-challenge-path")),
		ExpectedAcmeChallengeValue: lo.ToPtr(cmd.Config.String("acme-challenge-value")),
		OwnerID:                    lo.ToPtr(cmd.Config.String("org-id")),
	}

	o, err := client.CreateDNSVerification(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
