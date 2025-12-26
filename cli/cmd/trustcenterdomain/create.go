//go:build cli

package trustcenterdomain

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new trust center domain",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("cname-record", "c", "", "CNAME record name for the custom domain")
	createCmd.Flags().StringP("trust-center-id", "t", "", "Trust center ID to associate the domain with")
}

func createValidation() error {
	cnameRecord := cmd.Config.String("cname-record")
	if cnameRecord == "" {
		return cmd.NewRequiredFieldMissingError("cname-record")
	}

	trustCenterID := cmd.Config.String("trust-center-id")
	if trustCenterID == "" {
		return cmd.NewRequiredFieldMissingError("trust-center-id")
	}

	return nil
}

// create a new trust center domain
func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	cnameRecord := cmd.Config.String("cname-record")
	if cnameRecord == "" {
		cobra.CheckErr(cmd.NewRequiredFieldMissingError("cname-record"))
	}

	trustCenterID := cmd.Config.String("trust-center-id")
	if trustCenterID == "" {
		cobra.CheckErr(cmd.NewRequiredFieldMissingError("trust-center-id"))
	}

	cobra.CheckErr(createValidation())

	input := graphclient.CreateTrustCenterDomainInput{
		CnameRecord:   cnameRecord,
		TrustCenterID: trustCenterID,
	}

	o, err := client.CreateTrustCenterDomain(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
