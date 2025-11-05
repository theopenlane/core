//go:build clistripe

package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/theopenlane/utils/cli/tables"
)

// listCommand is the CLI action that prints existing Stripe webhook endpoints
func listCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "list all Stripe webhook endpoints",
		Action: func(ctx context.Context, c *cli.Command) error {
			client, err := getStripeClient(c.Root())
			if err != nil {
				return err
			}

			endpoints, err := client.ListWebhookEndpoints(ctx)
			if err != nil {
				return err
			}

			writer := tables.NewTableWriter(outWriter, "ID", "URL", "Status", "API Version", "Events")
			for _, endpoint := range endpoints {
				if err := writer.AddRow(
					endpoint.ID,
					endpoint.URL,
					string(endpoint.Status),
					endpoint.APIVersion,
					fmt.Sprintf("%d events", len(endpoint.EnabledEvents)),
				); err != nil {
					return err
				}
			}
			if err := writer.Render(); err != nil {
				return err
			}

			return nil
		},
	}
}
