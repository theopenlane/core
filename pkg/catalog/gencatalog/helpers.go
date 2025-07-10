package gencatalog

import (
	"sort"

	catalog "github.com/theopenlane/core/pkg/catalog"
)

// PersonalOrgMonthlyPriceIDs returns Stripe price IDs for monthly prices of modules marked as PersonalOrg enabled
func PersonalOrgMonthlyPriceIDs() []string {
	return monthlyPriceIDs(func(f catalog.Feature) bool {
		return f.PersonalOrg
	})
}

// TrialMonthlyPriceIDs returns Stripe price IDs for monthly prices of modules included with trial subscriptions
func TrialMonthlyPriceIDs() []string {
	return monthlyPriceIDs(func(f catalog.Feature) bool {
		return f.IncludeWithTrial
	})
}

func monthlyPriceIDs(filter func(catalog.Feature) bool) []string {
	ids := make([]string, 0)
	for _, f := range DefaultCatalog.Modules {
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
