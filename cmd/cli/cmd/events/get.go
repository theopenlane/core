package events

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
)

var eventGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get details of existing events",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	cmd.AddCommand(eventGetCmd)

	eventGetCmd.Flags().StringP("id", "i", "", "get a specific event by ID")
}

// get retrieves all events or a specific event by ID
func get(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	id := cmd.Config.String("id")

	if id != "" {
		o, err := client.GetEventByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	o, err := client.GetAllEvents(ctx)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
