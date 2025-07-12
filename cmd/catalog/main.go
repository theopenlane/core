package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/stripe/stripe-go/v82"
	"github.com/theopenlane/core/pkg/catalog"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/utils/cli/tables"
)

// takeoverInfo stores details about a price we may want to manage via metadata
type takeoverInfo struct {
	feature string
	price   catalog.Price
	stripe  *stripe.Price
}

// featureReport represents the reconciliation status for a feature
type featureReport struct {
	kind          string
	name          string
	product       bool
	missingPrices int
}

// stripeClient defines the subset of entitlements.StripeClient used by this CLI
type stripeClient interface {
	ListProducts(ctx context.Context) ([]*stripe.Product, error)
	GetPrice(ctx context.Context, id string) (*stripe.Price, error)
	FindPriceForProduct(ctx context.Context, productID string, currency string, unitAmount int64, interval, nickname, lookupKey, metadata string, meta map[string]string) (*stripe.Price, error)
	GetPriceByLookupKey(ctx context.Context, lookupKey string) (*stripe.Price, error)
	GetFeatureByLookupKey(ctx context.Context, lookupKey string) (*stripe.EntitlementsFeature, error)
	GetProduct(ctx context.Context, id string) (*stripe.Product, error)
	UpdatePriceMetadata(ctx context.Context, priceID string, metadata map[string]string) (*stripe.Price, error)
}

// newClient is a function that creates a new stripe client. It can be replaced in tests for mocking purposes
var newClient = func(opts ...entitlements.StripeOptions) (stripeClient, error) {
	return entitlements.NewStripeClient(opts...)
}

// outWriter is the output writer for the CLI; it can be replaced in tests for fun and profit
var outWriter io.Writer = os.Stdout

// main is the entry point for the catalog CLI application
func main() {
	if err := catalogApp().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// catalogApp creates a CLI application for reconciling a catalog with Stripe
func catalogApp() *cli.Command {
	app := &cli.Command{
		Name:  "catalog",
		Usage: "reconcile catalog with Stripe",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "catalog",
				Usage: "catalog file path",
				Value: "./pkg/catalog/catalog_sandbox.yaml", // set the value to sandbox catalog by default to avoid disasters
			},
			&cli.StringFlag{
				Name:    "stripe-key",
				Usage:   "stripe API key",
				Sources: cli.EnvVars("STRIPE_API_KEY"),
			},
			&cli.BoolFlag{
				Name:  "takeover",
				Usage: "add managed_by metadata when found",
			},
			&cli.BoolFlag{
				Name:  "write",
				Usage: "write price IDs back to catalog file",
				Value: true, // default to true to ensure up to date price IDs
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
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

			sc, err := newClient(entitlements.WithAPIKey(apiKey))
			if err != nil {
				return fmt.Errorf("stripe client: %w", err)
			}

			conflicts, err := cat.LookupKeyConflicts(ctx, sc)
			if err != nil {
				return fmt.Errorf("lookup keys: %w", err)
			}

			for _, c := range conflicts {
				managed, merr := conflictManaged(ctx, sc, c)
				if merr != nil {
					fmt.Fprintf(os.Stderr, "lookup key %s already exists as %s %s for %s\n", c.LookupKey, c.Resource, c.ID, c.Feature)
					continue
				}

				if !managed {
					fmt.Fprintf(os.Stderr, "lookup key %s already exists as %s %s for %s\n", c.LookupKey, c.Resource, c.ID, c.Feature)
					continue
				}

				parts := strings.SplitN(c.Feature, ":", 2) // nolint:mnd
				if len(parts) != 2 {                       // nolint:mnd
					continue
				}

				name := parts[1]
				var feat catalog.Feature
				var fs catalog.FeatureSet
				if parts[0] == "module" {
					feat = cat.Modules[name]
					fs = cat.Modules
				} else {
					feat = cat.Addons[name]
					fs = cat.Addons
				}

				switch c.Resource {
				case "product":
					feat.ProductID = c.ID
				case "price":
					for i, p := range feat.Billing.Prices {
						if p.LookupKey == c.LookupKey {
							feat.Billing.Prices[i].PriceID = c.ID
						}
					}
				}

				fs[name] = feat
			}

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
				if client, ok := sc.(*entitlements.StripeClient); ok {
					if err := promptAndCreateMissing(ctx, cat, client); err != nil {
						return fmt.Errorf("create prices: %w", err)
					}
				} else {
					return fmt.Errorf("create prices: %w", ErrExpectedClient)
				}
			}

			// Optionally write updated price IDs back to disk
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

var ErrExpectedClient = fmt.Errorf("expected stripeClient type")

// buildProductMap fetches all existing Stripe products and indexes them by ID
// and name so lookups can prefer unique identifiers when available
func buildProductMap(ctx context.Context, sc stripeClient) (map[string]*stripe.Product, error) {
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

// conflictManaged checks whether the conflict resource is managed by module manager.
func conflictManaged(ctx context.Context, sc stripeClient, c catalog.LookupKeyConflict) (bool, error) {
	switch c.Resource {
	case "product":
		prod, err := sc.GetProduct(ctx, c.ID)
		if err != nil {
			return false, err
		}

		if prod != nil && prod.Metadata != nil {
			return prod.Metadata[catalog.ManagedByKey] == catalog.ManagedByValue, nil
		}
	case "feature":
		feat, err := sc.GetFeatureByLookupKey(ctx, c.LookupKey)
		if err != nil {
			return false, err
		}

		if feat != nil && feat.Metadata != nil {
			return feat.Metadata[catalog.ManagedByKey] == catalog.ManagedByValue, nil
		}
	case "price":
		price, err := sc.GetPrice(ctx, c.ID)
		if err != nil {
			return false, err
		}

		if price != nil && price.Metadata != nil {
			return price.Metadata[catalog.ManagedByKey] == catalog.ManagedByValue, nil
		}
	}

	return false, nil
}

// resolveProduct attempts to find the Stripe product for a feature using
// progressively less unique attributes. It tries price IDs first, then price
// lookup keys, and finally falls back to the feature display name
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
func updateFeaturePrices(ctx context.Context, sc stripeClient, prod *stripe.Product, name string, feat catalog.Feature) (catalog.Feature, []takeoverInfo, int) {
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
// It mutates the provided fs map by updating its features with resolved price IDs.
// It returns a slice of unmanaged prices and whether any products or prices are missing.
func processFeatureSet(ctx context.Context, sc stripeClient, prodMap map[string]*stripe.Product, kind string, fs catalog.FeatureSet) ([]takeoverInfo, bool, []featureReport) {
	var takeovers []takeoverInfo

	missing := false

	var reports []featureReport

	names := slices.Collect(maps.Keys(fs))
	slices.Sort(names)

	for _, name := range names {
		feat := fs[name]

		var prod *stripe.Product

		if entSc, ok := sc.(*entitlements.StripeClient); ok {
			prod, _ = resolveProduct(ctx, entSc, prodMap, feat)
		} else if p, ok := prodMap[feat.DisplayName]; ok {
			prod = p
		}

		prodExists := prod != nil
		missingPrices := 0

		if prodExists {
			var t []takeoverInfo

			feat, t, missingPrices = updateFeaturePrices(ctx, sc, prod, name, feat)
			feat.ProductID = prod.ID

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
func handleTakeovers(ctx context.Context, sc stripeClient, takeovers []takeoverInfo, takeover *bool) error {
	if len(takeovers) == 0 {
		return nil
	}

	writer := tables.NewTableWriter(outWriter, "Feature", "LookupKey", "PriceID", "Managed")

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

// promptAndCreateMissing asks the user whether to create any missing products or prices
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

// priceMatchesStripe is what's performing the check on the key required
// non-changing fields to determine if the catalog price has drifted from the Stripe price
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

	writer := tables.NewTableWriter(outWriter, "Type", "Feature", "Product", "MissingPrices")

	for _, r := range reports {
		_ = writer.AddRow(r.kind, string(r.name), r.product, r.missingPrices)
	}

	_ = writer.Render()
}
