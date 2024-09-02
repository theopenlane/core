package search

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
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

	results, err := client.GlobalSearch(ctx, query)
	cobra.CheckErr(err)

	return consoleOutput(results)
}

func consoleOutput(results *openlaneclient.GlobalSearch) error {
	return jsonOutput(results)
}

// jsonOutput prints the output in a JSON format
func jsonOutput(out any) error {
	s, err := json.Marshal(out)
	cobra.CheckErr(err)

	return cmd.JSONPrint(s)
}
