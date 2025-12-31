package entitlements

import (
	"sort"

	"github.com/theopenlane/core/pkg/catalog"
	"github.com/theopenlane/core/pkg/catalog/gencatalog"
	"github.com/theopenlane/core/pkg/entitlements"
)

// TrialMonthlyPriceIDs returns Stripe price IDs for monthly prices of modules included with trial subscriptions
func TrialMonthlyPriceIDs(useSandbox bool) []string {
	return monthlyPriceIDs(func(f catalog.Feature) bool {
		return f.IncludeWithTrial
	}, useSandbox)
}

// TrialMonthlyPrices returns prices for monthly modules included with trial subscriptions
func TrialMonthlyPrices(useSandbox bool) []entitlements.Price {
	return monthlyPrices(func(f catalog.Feature) bool {
		return f.IncludeWithTrial
	}, useSandbox)
}

// AllMonthlyPrices returns prices for all monthly modules regardless of trial status
func AllMonthlyPrices(useSandbox bool) []entitlements.Price {
	return monthlyPrices(func(_ catalog.Feature) bool {
		return true
	}, useSandbox)
}

// monthlyPriceIDs returns Stripe price IDs for monthly prices of modules that match the provided filter function
func monthlyPriceIDs(filter func(catalog.Feature) bool, useSandbox bool) []string {
	ids := make([]string, 0)

	for _, f := range gencatalog.GetModules(useSandbox) {
		if filter(f) {
			for _, p := range f.Billing.Prices {
				if p.Interval == "month" {
					ids = append(ids, p.PriceID)
				}
			}
		}
	}

	sort.Strings(ids)

	return ids
}

// monthlyPriceIDs returns the entitlements.Prices for monthly prices of modules that match the provided filter function
func monthlyPrices(filter func(catalog.Feature) bool, useSandbox bool) []entitlements.Price {
	prices := make([]entitlements.Price, 0)

	for module, f := range gencatalog.GetModules(useSandbox) {
		if filter(f) {
			for _, p := range f.Billing.Prices {
				if p.Interval == "month" {
					prices = append(prices, entitlements.Price{
						ID:          p.PriceID,
						Price:       float64(p.UnitAmount),
						Interval:    p.Interval,
						ProductID:   f.ProductID,
						ProductName: module,
					})
				}
			}
		}
	}

	return prices
}
