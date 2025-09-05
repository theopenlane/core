package gencatalog

import (
	"sort"

	catalog "github.com/theopenlane/core/pkg/catalog"
)

// GetDefaultCatalog returns the default catalog; if useSandbox is true, it returns the sandbox catalog
func GetDefaultCatalog(useSandbox bool) catalog.Catalog {
	if useSandbox {
		return DefaultSandboxCatalog
	}

	return DefaultCatalog
}

// GetModules returns a list of all module names in the default catalog
func GetModules(useSandbox bool) catalog.FeatureSet {
	c := GetDefaultCatalog(useSandbox)

	return c.Modules
}

// GetCatalogByAudience returns a catalog filtered by audience; if useSandbox is true, it uses the sandbox catalog
func GetCatalogByAudience(useSandbox bool, audience string) *catalog.Catalog {
	c := GetDefaultCatalog(useSandbox)

	return c.Visible(audience)
}

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
