package entitlements

import (
	"sort"

	"github.com/theopenlane/core/pkg/catalog"
	"github.com/theopenlane/core/pkg/catalog/gencatalog"
)

// PersonalOrgMonthlyPriceIDs returns Stripe price IDs for monthly prices of modules marked as PersonalOrg enabled
func PersonalOrgMonthlyPriceIDs(useSandbox bool) []string {
	return monthlyPriceIDs(func(f catalog.Feature) bool {
		return f.PersonalOrg
	}, useSandbox)
}

// TrialMonthlyPriceIDs returns Stripe price IDs for monthly prices of modules included with trial subscriptions
func TrialMonthlyPriceIDs(useSandbox bool) []string {
	return monthlyPriceIDs(func(f catalog.Feature) bool {
		return f.IncludeWithTrial
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
