package entmapping

import (
	"github.com/theopenlane/core/pkg/catalog"
	"github.com/theopenlane/core/pkg/entitlements"
)

// EntitlementsProductToCatalogV2 maps entitlements.Product to catalog.Feature (module/addon)
func EntitlementsProductToCatalogV2(ep *entitlements.Product) *catalog.Feature {
	if ep == nil {
		return nil
	}
	return &catalog.Feature{
		DisplayName: ep.Name,
		Description: ep.Description,
		// Billing and Audience mapping can be extended as needed
	}
}

// CatalogPriceToEntitlementsV2 maps catalog.Price to entitlements.Price
func CatalogPriceToEntitlementsV2(cp *catalog.Price) *entitlements.Price {
	if cp == nil {
		return nil
	}
	return &entitlements.Price{
		ID:       cp.PriceID,
		Price:    float64(cp.UnitAmount) / 100.0,
		Interval: cp.Interval,
		// Currency, ProductID, etc. can be mapped as needed
	}
}

// EntitlementsPriceToCatalogV2 maps entitlements.Price to catalog.Price
func EntitlementsPriceToCatalogV2(ep *entitlements.Price) *catalog.Price {
	if ep == nil {
		return nil
	}
	return &catalog.Price{
		PriceID:    ep.ID,
		UnitAmount: int64(ep.Price * 100),
		Interval:   ep.Interval,
		// Currency, Nickname, LookupKey, Metadata can be mapped as needed
	}
}

// Extend with more mapping functions as needed for ent/generated types.
