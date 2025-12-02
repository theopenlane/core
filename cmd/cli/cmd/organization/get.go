//go:build cli

package org

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
	"github.com/theopenlane/iam/tokens"
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
	getCmd.Flags().BoolP("current-only", "c", false, "get the current authorized organization only")
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

	// get the current organization based on the access token in the client
	if cmd.Config.Bool("current-only") {
		// if the current only flag is set, we only want the current organization
		token, err := client.Config().Credentials.AccessToken()
		cobra.CheckErr(err)

		jwt, err := tokens.ParseUnverifiedTokenClaims(token)
		cobra.CheckErr(err)

		if jwt.OrgID == "" {
			log.Error().Err(err).Msg("no organization ID found in the token claims, cannot get current organization")
		}

		o, err := client.GetOrganizationByID(ctx, jwt.OrgID)
		cobra.CheckErr(err)

		return consoleOutput(o)
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

	o, err := client.GetOrganizations(ctx, cmd.First, cmd.Last, where)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
