package entmapping

//import (
//	"github.com/theopenlane/core/internal/ent"
//	"github.com/theopenlane/core/pkg/catalog"
//	"github.com/theopenlane/core/pkg/entitlements"
//)
//
//// Example mapping: Entitlements Product -> Catalog Product -> Ent Product
//
//// EntitlementsProductToCatalog maps entitlements.Product to catalog.Product
//func EntitlementsProductToCatalog(ep *entitlements.Product) *catalog.Product {
//	if ep == nil {
//		return nil
//	}
//	return &catalog.Product{
//		ID:          ep.ID,
//		Name:        ep.Name,
//		Description: ep.Description,
//		Metadata:    ep.Metadata,
//	}
//}
//
//// CatalogProductToEnt maps catalog.Product to ent.Product
//func CatalogProductToEnt(cp *catalog.Product) *ent.Product {
//	if cp == nil {
//		return nil
//	}
//	return &ent.Product{
//		ID:          cp.ID,
//		Name:        cp.Name,
//		Description: cp.Description,
//		// Add other fields as needed
//	}
//}
//
//// EntitlementsPriceToCatalog maps entitlements.Price to catalog.Price
//func EntitlementsPriceToCatalog(ep *entitlements.Price) *catalog.Price {
//	if ep == nil {
//		return nil
//	}
//	return &catalog.Price{
//		ID:        ep.ID,
//		Price:     ep.Price,
//		ProductID: ep.ProductID,
//		Interval:  ep.Interval,
//		Currency:  ep.Currency,
//	}
//}
//
//// CatalogPriceToEnt maps catalog.Price to ent.Price
//func CatalogPriceToEnt(cp *catalog.Price) *ent.Price {
//	if cp == nil {
//		return nil
//	}
//	return &ent.Price{
//		ID:        cp.ID,
//		Price:     cp.Price,
//		ProductID: cp.ProductID,
//		Interval:  cp.Interval,
//		Currency:  cp.Currency,
//	}
//}

// Add similar mapping functions for Customer, Subscription, etc. as needed.
