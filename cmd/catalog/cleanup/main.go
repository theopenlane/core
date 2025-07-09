package main

import (
	"context"
	"fmt"
	"os"

	"github.com/stripe/stripe-go/v82"
	"github.com/urfave/cli/v3"

	"github.com/theopenlane/core/pkg/entitlements"
)

func priceInUse(ctx context.Context, sc *entitlements.StripeClient, priceID string) (bool, error) {
	params := &stripe.SubscriptionListParams{Price: stripe.String(priceID)}
	seq := sc.Client.V1Subscriptions.List(ctx, params)

	var firstErr error
	inUse := false

	seq(func(_ *stripe.Subscription, err error) bool {
		if err != nil {
			firstErr = err
			return false
		}

		inUse = true
		return false
	})

	return inUse, firstErr
}

func detachFeatures(ctx context.Context, sc *entitlements.StripeClient, prodID string, used map[string]struct{}, keep bool) error {
	features, err := sc.ListProductFeatures(ctx, prodID)
	if err != nil {
		return fmt.Errorf("list features for %s: %w", prodID, err)
	}

	for _, f := range features {
		if f.EntitlementFeature == nil {
			continue
		}

		if keep {
			used[f.EntitlementFeature.ID] = struct{}{}
			continue
		}

		if _, err := sc.Client.V1ProductFeatures.Delete(ctx, f.ID, &stripe.ProductFeatureDeleteParams{Product: stripe.String(prodID)}); err != nil {
			return fmt.Errorf("detach feature %s from product %s: %w", f.ID, prodID, err)
		}

		fmt.Printf("detached feature %s from product %s\n", f.ID, prodID)
	}

	return nil
}

func runCleanup(ctx context.Context, sc *entitlements.StripeClient) error {
	usedFeatures := make(map[string]struct{})

	products, err := sc.ListProducts(ctx)
	if err != nil {
		return fmt.Errorf("list products: %w", err)
	}

	for _, prod := range products {
		prodUsed := false

		prices, err := sc.ListPricesForProduct(ctx, prod.ID)
		if err != nil {
			return fmt.Errorf("list prices for %s: %w", prod.ID, err)
		}

		for _, price := range prices {
			inUse, err := priceInUse(ctx, sc, price.ID)
			if err != nil {
				return fmt.Errorf("check subscriptions for price %s: %w", price.ID, err)
			}

			if inUse {
				prodUsed = true
				continue
			}

			if _, err := sc.Client.V1Prices.Update(ctx, price.ID, &stripe.PriceUpdateParams{Active: stripe.Bool(false)}); err != nil {
				return fmt.Errorf("archive price %s: %w", price.ID, err)
			}
			fmt.Printf("archived price %s\n", price.ID)
		}

		if err := detachFeatures(ctx, sc, prod.ID, usedFeatures, prodUsed); err != nil {
			return err
		}

		if !prodUsed {
			if _, err := sc.UpdateProductWithParams(ctx, prod.ID, &stripe.ProductUpdateParams{Active: stripe.Bool(false)}); err != nil {
				return fmt.Errorf("archive product %s: %w", prod.ID, err)
			}
			fmt.Printf("archived product %s\n", prod.ID)
		}
	}

	it := sc.Client.V1EntitlementsFeatures.List(ctx, &stripe.EntitlementsFeatureListParams{})
	for f, err := range it {
		if err != nil {
			return fmt.Errorf("list features: %w", err)
		}

		if _, ok := usedFeatures[f.ID]; !ok {
			if _, err := sc.Client.V1EntitlementsFeatures.Update(ctx, f.ID, &stripe.EntitlementsFeatureUpdateParams{Active: stripe.Bool(false)}); err != nil {
				return fmt.Errorf("archive feature %s: %w", f.ID, err)
			}

			fmt.Printf("archived feature %s\n", f.ID)
		}
	}

	return nil
}

func cleanupApp() *cli.Command {
	return &cli.Command{
		Name:  "cleanup",
		Usage: "remove unused Stripe features, prices and products",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "stripe-key", Usage: "stripe API key", Sources: cli.EnvVars("STRIPE_API_KEY")},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			key := c.String("stripe-key")
			client, err := entitlements.NewStripeClient(entitlements.WithAPIKey(key))
			if err != nil {
				return fmt.Errorf("stripe client: %w", err)
			}
			return runCleanup(ctx, client)
		},
	}
}

func main() {
	if err := cleanupApp().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
