package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/stripe/stripe-go/v82"

	"github.com/theopenlane/core/pkg/catalog"
	"github.com/theopenlane/core/pkg/entitlements"
)

const apiKey = ""

func main() {
	var catalogFile string
	var apiKey string
	flag.StringVar(&catalogFile, "catalog", "./config/catalog.yaml", "catalog file path")
	flag.StringVar(&apiKey, "stripe-key", "", "stripe API key (or set STRIPE_API_KEY)")
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

	check := func(kind string, fs catalog.FeatureSet) {
		for name, feat := range fs {
			prod, ok := prodMap[feat.DisplayName]
			prodExists := ok
			missingPrices := 0
			if ok {
				for _, p := range feat.Billing.Prices {
					md := map[string]string{catalog.ManagedByKey: catalog.ManagedByValue}
					for k, v := range p.Metadata {
						md[k] = v
					}
					price, err := sc.FindPriceForProduct(ctx, prod.ID, p.UnitAmount, "", p.Interval, p.Nickname, p.LookupKey, md)
					if err != nil || price == nil {
						missingPrices++
					}
				}
			} else {
				missingPrices = len(feat.Billing.Prices)
			}
			fmt.Printf("%s %-20s product:%v missing_prices:%d\n", kind, name, prodExists, missingPrices)
		}
	}

	check("module", cat.Modules)
	check("addon", cat.Addons)
}
