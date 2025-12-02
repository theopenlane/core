//go:build cli

package customdomain

import (
	"context"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new custom domain",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("cname-record", "c", "", "CNAME record name")
	createCmd.Flags().StringP("mappable-domain", "m", "", "Domain CNAME will be mapped to")
	createCmd.Flags().StringP("org-id", "o", "", "Org ID to associate the custom domain with")
}

func createValidation() error {

	cnameRecord := cmd.Config.String("cname-record")
	if cnameRecord == "" {
		return cmd.NewRequiredFieldMissingError("cname-record")
	}

	mappableDomainName := cmd.Config.String("mappable-domain")
	if mappableDomainName == "" {
		return cmd.NewRequiredFieldMissingError("mappable-domain")
	}

	return nil
}

// create a new custom domain
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

	mappableDomainName := cmd.Config.String("mappable-domain")
	if mappableDomainName == "" {
		cobra.CheckErr(cmd.NewRequiredFieldMissingError("mappable-domain"))
	}

	cobra.CheckErr(createValidation())

	mappableDomains, err := client.GetMappableDomains(ctx, cmd.First, cmd.Last, &openlaneclient.MappableDomainWhereInput{
		Name: &mappableDomainName,
	})
	cobra.CheckErr(err)

	if len(mappableDomains.MappableDomains.Edges) == 0 {
		cobra.CheckErr(cmd.ErrNotFound)
	}
	input := openlaneclient.CreateCustomDomainInput{
		CnameRecord:      cnameRecord,
		MappableDomainID: mappableDomains.MappableDomains.Edges[0].Node.ID,
		OwnerID:          lo.ToPtr(cmd.Config.String("org-id")),
	}

	o, err := client.CreateCustomDomain(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
