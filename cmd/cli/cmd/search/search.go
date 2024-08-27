package search

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/theopenlane/core/cmd/cli/cmd"

	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/cli/tables"
)

var command = &cobra.Command{
	Use:   "search",
	Short: "search for organizations, groups, users, subscribers, etc in the system",
	Run: func(cmd *cobra.Command, args []string) {
		err := search(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	cmd.RootCmd.AddCommand(command)

	command.Flags().StringP("query", "q", "", "query string to search for")
}

// validate validates the required fields for the command
func validate() (string, error) {
	query := cmd.Config.String("query")
	if query == "" {
		return "", cmd.NewRequiredFieldMissingError("query")
	}

	return query, nil
}

// search searches for organizations, groups, users, subscribers, etc in the system
func search(ctx context.Context) error { // setup http client
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	// filter options
	query, err := validate()
	cobra.CheckErr(err)

	results, err := client.Search(ctx, query)
	cobra.CheckErr(err)

	consoleOutput(results)

	return nil
}

func consoleOutput(results *openlaneclient.Search) {
	// print results
	for _, r := range results.Search.Nodes {
		if len(r.OrganizationSearchResult.Organizations) > 0 {
			fmt.Println("Organization Results")

			writer := tables.NewTableWriter(cmd.RootCmd.OutOrStdout(), "ID", "Name", "DisplayName", "Description")

			for _, o := range r.OrganizationSearchResult.Organizations {
				writer.AddRow(o.ID, o.Name, o.DisplayName, *o.Description)
			}

			writer.Render()
		}

		if len(r.GroupSearchResult.Groups) > 0 {
			fmt.Println("Group Results")

			writer := tables.NewTableWriter(cmd.RootCmd.OutOrStdout(), "ID", "Name", "DisplayName", "Description")

			for _, g := range r.GroupSearchResult.Groups {
				writer.AddRow(g.ID, g.Name, g.DisplayName, *g.Description)
			}

			writer.Render()
		}

		if len(r.UserSearchResult.Users) > 0 {
			fmt.Println("User Results")

			writer := tables.NewTableWriter(cmd.RootCmd.OutOrStdout(), "ID", "FirstName", "LastName", "DisplayName", "Email")

			for _, u := range r.UserSearchResult.Users {
				writer.AddRow(u.ID, *u.FirstName, *u.LastName, u.DisplayName, u.Email)
			}

			writer.Render()
		}

		if len(r.SubscriberSearchResult.Subscribers) > 0 {
			fmt.Println("Subscriber Results")

			writer := tables.NewTableWriter(cmd.RootCmd.OutOrStdout(), "ID", "Email")

			for _, s := range r.SubscriberSearchResult.Subscribers {
				writer.AddRow(s.ID, s.Email)
			}

			writer.Render()
		}
	}
}
