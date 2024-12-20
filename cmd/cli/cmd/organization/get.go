package org

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get details of existing organization(s)",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "get a specific organization by ID")
	getCmd.Flags().BoolP("include-personal-orgs", "p", false, "include personal organizations in the output")
}

// get an organization in the platform
func get(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	// filter options
	id := cmd.Config.String("id")

	// if an org ID is provided, filter on that organization, otherwise get all
	if id != "" {
		o, err := client.GetOrganizationByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	includePersonalOrgs := cmd.Config.Bool("include-personal-orgs")

	if includePersonalOrgs {
		o, err := client.GetAllOrganizations(ctx)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	// don't include personal orgs
	where := &openlaneclient.OrganizationWhereInput{
		PersonalOrg: &includePersonalOrgs,
	}

	o, err := client.GetOrganizations(ctx, where)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
