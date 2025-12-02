//go:build cli

package dnsverification

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get DNS verification records",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "DNS verification ID")
	getCmd.Flags().StringP("cloudflare-hostname-id", "c", "", "Cloudflare hostname ID")
	getCmd.Flags().StringP("org-id", "o", "", "Organization ID")
}

// get retrieves DNS verification records from the platform
func get(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id := cmd.Config.String("id")

	if id != "" {
		o, err := client.GetDNSVerificationByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	// Build filter criteria
	where := &openlaneclient.DNSVerificationWhereInput{}

	if cloudflareHostnameID := cmd.Config.String("cloudflare-hostname-id"); cloudflareHostnameID != "" {
		where.CloudflareHostnameID = &cloudflareHostnameID
	}

	if orgID := cmd.Config.String("org-id"); orgID != "" {
		where.OwnerID = &orgID
	}

	// If any filters are set, use GetDNSVerifications with filters
	if where.CloudflareHostnameID != nil || where.OwnerID != nil {
		o, err := client.GetDNSVerifications(ctx, cmd.First, cmd.Last, where)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	// Otherwise get all DNS verification records
	o, err := client.GetAllDNSVerifications(ctx)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
