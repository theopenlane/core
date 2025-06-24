package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"maps"
	"os"
	"strings"

	"github.com/stripe/stripe-go/v82"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/pkg/catalog"
	"github.com/theopenlane/core/pkg/entitlements"
)

func main() {
	var catalogFile string
	var apiKey string
	var takeover bool
	var write bool
	flag.StringVar(&catalogFile, "catalog", "./pkg/catalog/catalog.yaml", "catalog file path")
	flag.StringVar(&apiKey, "stripe-key", "", "stripe API key (or set STRIPE_API_KEY)")
	flag.BoolVar(&takeover, "takeover", false, "add managed_by metadata when found")
	flag.BoolVar(&write, "write", false, "write price IDs back to catalog file")
	flag.Parse()

	if apiKey == "" {
		apiKey = os.Getenv("STRIPE_API_KEY")
	}

	cat, err := catalog.LoadCatalog(catalogFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load catalog:", err)
		os.Exit(1)
	}

	sc, err := entitlements.NewStripeClient(entitlements.WithAPIKey(apiKey))
	if err != nil {
		fmt.Fprintln(os.Stderr, "stripe client:", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// map product name -> product
	products, err := sc.ListProducts(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "list products:", err)
		os.Exit(1)
	}
	prodMap := map[string]*stripe.Product{}
	for _, p := range products {
		prodMap[p.Name] = p
	}

	type takeoverInfo struct {
		feature string
		price   catalog.Price
		stripe  *stripe.Price
	}
	var takeovers []takeoverInfo
	missing := false

	check := func(kind string, fs catalog.FeatureSet) {
		for name, feat := range fs {
			prod, ok := prodMap[feat.DisplayName]
			prodExists := ok
			missingPrices := 0
			if ok {
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
						price, err = sc.FindPriceForProduct(ctx, prod.ID, p.UnitAmount, "", p.Interval, p.Nickname, p.LookupKey, md)
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
			} else {
				missingPrices = len(feat.Billing.Prices)
			}

			if !prodExists || missingPrices > 0 {
				missing = true
			}

			fs[name] = feat
			fmt.Printf("%s %-20s product:%v missing_prices:%d\n", kind, name, prodExists, missingPrices)
		}
	}

	check("module", cat.Modules)
	check("addon", cat.Addons)

	if len(takeovers) > 0 {
		writer := tables.NewTableWriter(os.Stdout, "Feature", "LookupKey", "PriceID", "Managed")
		for _, t := range takeovers {
			managed := t.stripe.Metadata[catalog.ManagedByKey]
			if err := writer.AddRow(t.feature, t.price.LookupKey, t.stripe.ID, managed); err != nil {
				fmt.Fprintln(os.Stderr, "add row:", err)
				os.Exit(1)
			}
		}
		if err := writer.Render(); err != nil {
			fmt.Fprintln(os.Stderr, "render table:", err)
			os.Exit(1)
		}

		if !takeover {
			fmt.Print("Take over these prices by adding metadata? (y/N): ")
			r := bufio.NewReader(os.Stdin)
			answer, err := r.ReadString('\n')
			if err != nil {
				fmt.Fprintln(os.Stderr, "read:", err)
				os.Exit(1)
			}
			answer = strings.ToLower(strings.TrimSpace(answer))
			takeover = answer == "y" || answer == "yes"
		}

		if takeover {
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
	}

	if missing {
		fmt.Print("Create missing products and prices? (y/N): ")
		r := bufio.NewReader(os.Stdin)
		answer, err := r.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "read:", err)
			os.Exit(1)
		}
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer == "y" || answer == "yes" {
			if err := cat.EnsurePrices(ctx, sc, "usd"); err != nil {
				fmt.Fprintln(os.Stderr, "create prices:", err)
			}
		}
	}

	if write {
		if err := cat.SaveCatalog(catalogFile); err != nil {
			fmt.Fprintln(os.Stderr, "save catalog:", err)
			os.Exit(1)
		}
	}
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
