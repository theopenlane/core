//go:build ignore

package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/stripe/stripe-go/v84"
	"github.com/theopenlane/utils/cli/tables"
	"github.com/urfave/cli/v3"

	"github.com/theopenlane/core/pkg/entitlements"
)

// stripeClient defines the methods used by this CLI. It matches entitlements.StripeClient so tests can substitute a fake implementation
type stripeClient interface {
	TagPriceMigration(ctx context.Context, fromPriceID, toPriceID string) error
	ListSubscriptions(ctx context.Context, customerID string) ([]*stripe.Subscription, error)
	MigrateSubscriptionPrice(ctx context.Context, sub *stripe.Subscription, oldPriceID, newPriceID string) (*stripe.Subscription, error)
}

// newClient is a function that creates a new stripe client. It can be replaced in tests for mocking purposes
var newClient = func(opts ...entitlements.StripeOptions) (stripeClient, error) {
	return entitlements.NewStripeClient(opts...)
}

// outWriter is the output writer for the CLI. It can be replaced in tests for fun and profit
var outWriter io.Writer = os.Stdout

// migrationApp creates a new CLI application for migrating prices in Stripe
// It allows tagging a price migration and optionally migrating subscriptions from an old price to a new one
// It supports dry-run mode to list what would be migrated without making changes
// It requires the old and new price IDs, and optionally customer IDs to migrate
// The Stripe API key can be provided via a flag or environment variable
func migrationApp() *cli.Command {
	app := &cli.Command{
		Name:  "pricemigrate",
		Usage: "tag and optionally migrate subscriptions to a new price",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "old-price", Usage: "price ID to migrate from", Required: true},
			&cli.StringFlag{Name: "new-price", Usage: "price ID to migrate to", Required: true},
			&cli.StringSliceFlag{Name: "customers", Usage: "comma separated customer IDs to migrate"},
			&cli.StringFlag{Name: "stripe-key", Usage: "stripe API key", Sources: cli.EnvVars("STRIPE_API_KEY")},
			&cli.BoolFlag{Name: "no-migrate", Usage: "only tag the price and do not update subscriptions"},
			&cli.BoolFlag{Name: "dry-run", Usage: "list customers and subscriptions that would be migrated", Value: true}, // default to true to avoid accidental migrations
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			oldPrice := c.String("old-price")
			newPrice := c.String("new-price")
			apiKey := c.String("stripe-key")
			customers := c.StringSlice("customers")
			skip := c.Bool("no-migrate")
			dryRun := c.Bool("dry-run")

			sc, err := newClient(entitlements.WithAPIKey(apiKey))
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

			writer := tables.NewTableWriter(outWriter, "Customer", "Subscription", "From", "To")

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

	return app
}

func main() {
	if err := migrationApp().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
