//go:build cli

package customdomain

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get custom domains",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "custom domain ID")
	getCmd.Flags().StringP("cname", "c", "", "cname record")
	getCmd.Flags().StringP("org-id", "o", "", "organization ID")
}

// get retrieves custom domains from the platform
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
		o, err := client.GetCustomDomainByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	order := &graphclient.CustomDomainOrder{}
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.CustomDomainOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.CustomDomainOrderField(*cmd.OrderBy),
		}
	}

	// Build filter criteria
	where := &graphclient.CustomDomainWhereInput{}

	if domain := cmd.Config.String("cname"); domain != "" {
		where.CnameRecord = &domain
	}

	if orgID := cmd.Config.String("org-id"); orgID != "" {
		where.OwnerID = &orgID
	}

	// If any filters are set, use GetCustomDomains with filters
	if where.CnameRecord != nil || where.OwnerID != nil {
		o, err := client.GetCustomDomains(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, where, []*graphclient.CustomDomainOrder{order})
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	// Otherwise get all custom domains
	o, err := client.GetAllCustomDomains(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.CustomDomainOrder{order})
	cobra.CheckErr(err)

	return consoleOutput(o)
}
