package main

import (
	"fmt"
	"os"

	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/utils/cli/tables"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "pricemigrate",
		Usage: "tag and optionally migrate subscriptions to a new price",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "old-price", Usage: "price ID to migrate from", Required: true},
			&cli.StringFlag{Name: "new-price", Usage: "price ID to migrate to", Required: true},
			&cli.StringSliceFlag{Name: "customers", Usage: "comma separated customer IDs to migrate"},
			&cli.StringFlag{Name: "stripe-key", Usage: "stripe API key", EnvVars: []string{"STRIPE_API_KEY"}},
			&cli.BoolFlag{Name: "no-migrate", Usage: "only tag the price and do not update subscriptions"},
			&cli.BoolFlag{Name: "dry-run", Usage: "list customers and subscriptions that would be migrated"},
		},
		Action: func(c *cli.Context) error {
			ctx := c.Context

			oldPrice := c.String("old-price")
			newPrice := c.String("new-price")
			apiKey := c.String("stripe-key")
			customers := c.StringSlice("customers")
			skip := c.Bool("no-migrate")
			dryRun := c.Bool("dry-run")

			sc, err := entitlements.NewStripeClient(entitlements.WithAPIKey(apiKey))
			if err != nil {
				return fmt.Errorf("stripe client: %w", err)
			}

			if !dryRun {
				if err := sc.TagPriceMigration(ctx, oldPrice, newPrice); err != nil {
					return fmt.Errorf("tag price: %w", err)
				}
			}

			if skip || len(customers) == 0 {
				return nil
			}

			writer := tables.NewTableWriter(os.Stdout, "Customer", "Subscription", "From", "To")

			for _, cid := range customers {
				subs, err := sc.ListSubscriptions(ctx, cid)
				if err != nil {
					return fmt.Errorf("list subscriptions for %s: %w", cid, err)
				}

				for _, sub := range subs {
					hasOld := false
					for _, item := range sub.Items.Data {
						if item.Price != nil && item.Price.ID == oldPrice {
							hasOld = true
							break
						}
					}

					if !hasOld {
						continue
					}

					if dryRun {
						if err := writer.AddRow(cid, sub.ID, oldPrice, newPrice); err != nil {
							return err
						}
						continue
					}

					if _, err := sc.MigrateSubscriptionPrice(ctx, sub, oldPrice, newPrice); err != nil {
						return fmt.Errorf("update subscription %s: %w", sub.ID, err)
					}
				}
			}

			if dryRun {
				if err := writer.Render(); err != nil {
					return err
				}
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
