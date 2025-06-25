package main

import (
	"bufio"
	"context"
	"fmt"
	"maps"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/stripe/stripe-go/v82"
	"github.com/theopenlane/core/pkg/catalog"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/utils/cli/tables"
)

// takeoverInfo stores details about a price we may want to manage via metadata.
type takeoverInfo struct {
	feature string
	price   catalog.Price
	stripe  *stripe.Price
}

// featureReport represents the reconciliation status for a feature.
type featureReport struct {
	kind          string
	name          string
	product       bool
	missingPrices int
}

func main() {
	if err := catalogApp().Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func catalogApp() *cli.App {
	app := &cli.App{
		Name:  "catalog",
		Usage: "reconcile catalog with Stripe",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "catalog",
				Usage: "catalog file path",
				Value: "./pkg/catalog/catalog.yaml",
			},
			&cli.StringFlag{
				Name:    "stripe-key",
				Usage:   "stripe API key",
				EnvVars: []string{"STRIPE_API_KEY"},
			},
			&cli.BoolFlag{
				Name:  "takeover",
				Usage: "add managed_by metadata when found",
			},
			&cli.BoolFlag{
				Name:  "write",
				Usage: "write price IDs back to catalog file",
			},
		},
		Action: func(c *cli.Context) error {
			catalogFile := c.String("catalog")
			apiKey := c.String("stripe-key")
			takeover := c.Bool("takeover")
			write := c.Bool("write")

			cat, err := catalog.LoadCatalog(catalogFile)
			if err != nil {
				return fmt.Errorf("load catalog: %w", err)
			}

			if cat.IsCurrent() {
				fmt.Printf("Catalog version %s already processed, skipping reconciliation\n", cat.Version)
				return nil
			}

			sc, err := entitlements.NewStripeClient(entitlements.WithAPIKey(apiKey))
			if err != nil {
				return fmt.Errorf("stripe client: %w", err)
			}

			ctx := c.Context

			// Pull all existing products from Stripe to build a lookup by name.
			prodMap, err := buildProductMap(ctx, sc)
			if err != nil {
				return fmt.Errorf("list products: %w", err)
			}

			mods, missMods, modReports := processFeatureSet(ctx, sc, prodMap, "module", cat.Modules)
			adds, missAdds, addReports := processFeatureSet(ctx, sc, prodMap, "addon", cat.Addons)

			featuresreports := append([]featureReport{}, modReports...)
			featuresreports = append(featuresreports, addReports...)
			printFeatureReports(featuresreports)

			takeovers := append([]takeoverInfo{}, mods...)
			takeovers = append(takeovers, adds...)
			missing := missMods || missAdds

			// Offer to take over unmanaged prices if any were found.
			if err := handleTakeovers(ctx, sc, takeovers, &takeover); err != nil {
				return fmt.Errorf("takeover: %w", err)
			}

			// Prompt to create missing products or prices if needed.
			if missing {
				if err := promptAndCreateMissing(ctx, cat, sc); err != nil {
					return fmt.Errorf("create prices: %w", err)
				}
			}

			// Optionally write updated price IDs back to disk.
			if write {
				diff, err := cat.SaveCatalog(catalogFile)
				if err != nil {
					return fmt.Errorf("save catalog: %w", err)
				}
				fmt.Printf("Catalog successfully written to %s\n", catalogFile)
				if diff != "" {
					fmt.Println("Catalog changes:\n" + diff)
				}
			}

			return nil
		},
	}

	return app
}

// buildProductMap fetches all existing Stripe products and indexes them by ID
// and name so lookups can prefer unique identifiers when available.
func buildProductMap(ctx context.Context, sc *entitlements.StripeClient) (map[string]*stripe.Product, error) {
	products, err := sc.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	prodMap := map[string]*stripe.Product{}
	for _, p := range products {
		if p.ID != "" {
			prodMap[p.ID] = p
		}

		if p.Name != "" {
			prodMap[p.Name] = p
		}
	}

	return prodMap, nil
}

// resolveProduct attempts to find the Stripe product for a feature using
// progressively less unique attributes. It tries price IDs first, then price
// lookup keys, and finally falls back to the feature display name.
func resolveProduct(ctx context.Context, sc *entitlements.StripeClient, prodMap map[string]*stripe.Product, feat catalog.Feature) (*stripe.Product, error) {
	// try to discover product via price IDs
	for _, p := range feat.Billing.Prices {
		if p.PriceID == "" {
			continue
		}

		pr, err := sc.GetPrice(ctx, p.PriceID)
		if err == nil && pr != nil && pr.Product != nil {
			return sc.GetProductByID(ctx, pr.Product.ID)
		}
	}

	// next try lookup keys
	for _, p := range feat.Billing.Prices {
		if p.LookupKey == "" {
			continue
		}

		pr, err := sc.GetPriceByLookupKey(ctx, p.LookupKey)
		if err == nil && pr != nil && pr.Product != nil {
			return sc.GetProductByID(ctx, pr.Product.ID)
		}
	}

	// finally fall back to display name lookup in the provided product map
	if prod, ok := prodMap[feat.DisplayName]; ok {
		return prod, nil
	}

	return nil, nil
}

// updateFeaturePrices ensures each price in a feature has its price ID set and
// returns any prices that are missing management metadata.
func updateFeaturePrices(ctx context.Context, sc *entitlements.StripeClient, prod *stripe.Product, name string, feat catalog.Feature) (catalog.Feature, []takeoverInfo, int) {
	missingPrices := 0
	var takeovers []takeoverInfo

	for i, p := range feat.Billing.Prices {
		md := map[string]string{catalog.ManagedByKey: catalog.ManagedByValue}
		maps.Copy(md, p.Metadata)

		var price *stripe.Price
		var err error

		if p.PriceID != "" {
			price, err = sc.GetPrice(ctx, p.PriceID)
			if err != nil || price == nil {
				missingPrices++
				continue
			}

			if !priceMatchesStripe(price, p, prod.ID) {
				fmt.Fprintf(os.Stderr, "[WARN] price %s for feature %s does not match catalog; to modify an existing price create a new one and update subscriptions\n", p.PriceID, name)
			}
		} else {
			price, err = sc.FindPriceForProduct(ctx, prod.ID, "", p.UnitAmount, "", p.Interval, p.Nickname, p.LookupKey, md)
			if err != nil || price == nil {
				missingPrices++
				continue
			}
		}

		feat.Billing.Prices[i].PriceID = price.ID
		if price.Metadata[catalog.ManagedByKey] != catalog.ManagedByValue {
			takeovers = append(takeovers, takeoverInfo{feature: name, price: p, stripe: price})
		}
	}

	return feat, takeovers, missingPrices
}

// processFeatureSet reconciles a set of features with Stripe products and prices.
// It returns a slice of unmanaged prices and whether any products or prices are missing.
func processFeatureSet(ctx context.Context, sc *entitlements.StripeClient, prodMap map[string]*stripe.Product, kind string, fs catalog.FeatureSet) ([]takeoverInfo, bool, []featureReport) {
	var takeovers []takeoverInfo
	missing := false
	var reports []featureReport

	for name, feat := range fs {
		prod, _ := resolveProduct(ctx, sc, prodMap, feat)
		prodExists := prod != nil
		missingPrices := 0

		if prodExists {
			var t []takeoverInfo
			feat, t, missingPrices = updateFeaturePrices(ctx, sc, prod, name, feat)
			takeovers = append(takeovers, t...)
		} else {
			missingPrices = len(feat.Billing.Prices)
		}

		if !prodExists || missingPrices > 0 {
			missing = true
		}

		fs[name] = feat
		reports = append(reports, featureReport{kind: kind, name: name, product: prodExists, missingPrices: missingPrices})
	}

	return takeovers, missing, reports
}

// handleTakeovers optionally updates Stripe prices with management metadata.
func handleTakeovers(ctx context.Context, sc *entitlements.StripeClient, takeovers []takeoverInfo, takeover *bool) error {
	if len(takeovers) == 0 {
		return nil
	}

	writer := tables.NewTableWriter(os.Stdout, "Feature", "LookupKey", "PriceID", "Managed")
	for _, t := range takeovers {
		managed := t.stripe.Metadata[catalog.ManagedByKey]
		if err := writer.AddRow(t.feature, t.price.LookupKey, t.stripe.ID, managed); err != nil {
			return err
		}
	}
	if err := writer.Render(); err != nil {
		return err
	}

	if !*takeover {
		fmt.Print("Take over these prices by adding metadata? (y/N): ")
		r := bufio.NewReader(os.Stdin)
		answer, err := r.ReadString('\n')
		if err != nil {
			return err
		}
		answer = strings.ToLower(strings.TrimSpace(answer))
		*takeover = answer == "y" || answer == "yes"
	}

	if *takeover {
		for _, t := range takeovers {
			md := t.stripe.Metadata
			if md == nil {
				md = map[string]string{}
			}
			md[catalog.ManagedByKey] = catalog.ManagedByValue
			if _, err := sc.UpdatePriceMetadata(ctx, t.stripe.ID, md); err != nil {
				fmt.Fprintln(os.Stderr, "update price:", err)
			}
		}
	}

	return nil
}

// promptAndCreateMissing asks the user whether to create any missing products or prices.
func promptAndCreateMissing(ctx context.Context, cat *catalog.Catalog, sc *entitlements.StripeClient) error {
	fmt.Print("Create missing products and prices? (y/N): ")
	r := bufio.NewReader(os.Stdin)
	answer, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer == "y" || answer == "yes" {
		return cat.EnsurePrices(ctx, sc, "usd")
	}
	return nil
}

func priceMatchesStripe(p *stripe.Price, cp catalog.Price, prodID string) bool {
	if p == nil {
		return false
	}

	if p.Product == nil || p.Product.ID != prodID {
		return false
	}

	if cp.UnitAmount != 0 && p.UnitAmount != cp.UnitAmount {
		return false
	}

	if cp.Interval != "" {
		if p.Recurring == nil || string(p.Recurring.Interval) != cp.Interval {
			return false
		}
	}

	if cp.Nickname != "" && p.Nickname != cp.Nickname {
		return false
	}

	if cp.LookupKey != "" && p.LookupKey != cp.LookupKey {
		return false
	}

	for k, v := range cp.Metadata {
		if p.Metadata == nil || p.Metadata[k] != v {
			return false
		}
	}

	return true
}

// printFeatureReports renders a table summarizing missing products and prices.
func printFeatureReports(reports []featureReport) {
	if len(reports) == 0 {
		return
	}

	writer := tables.NewTableWriter(os.Stdout, "Type", "Feature", "Product", "MissingPrices")
	for _, r := range reports {
		_ = writer.AddRow(r.kind, r.name, r.product, r.missingPrices)
	}

	_ = writer.Render()
}
