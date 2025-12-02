//go:build cli

package trustcenter

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new trustcenter",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("custom-domain-id", "d", "", "custom domain id for the trustcenter")
	createCmd.Flags().StringSliceP("tags", "t", []string{}, "tags associated with the trustcenter")
	createCmd.Flags().StringP("org-id", "o", "", "Org ID to associate the custom domain with")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateTrustCenterInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags

	customDomainID := cmd.Config.String("custom-domain-id")
	if customDomainID != "" {
		input.CustomDomainID = &customDomainID
	}

	orgID := cmd.Config.String("org-id")
	if orgID != "" {
		input.OwnerID = &orgID
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	return input, nil
}

// create a new trustcenter
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

	o, err := client.CreateTrustCenter(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
