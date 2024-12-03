package entitlements

import (
	"fmt"
	"os"

	"github.com/stripe/stripe-go/v81"
	"gopkg.in/yaml.v3"
)

// GetProducts retrieves all products from stripe which are active
func (sc *StripeClient) GetProducts() []Product {
	productParams := &stripe.ProductListParams{}
	productParams.Filters.AddFilter("active", "", "true")

	iter := sc.Client.Products.List(productParams)
	products := []Product{}

	for iter.Next() {
		productData := iter.Product()
		if productData.DefaultPrice == nil {
			continue
		}

		priceData := sc.GetPrices()
		prices := []Price{}

		for _, price := range priceData {
			if price.ProductID == productData.ID {
				prices = append(prices, Price{
					ID:        price.ID,
					Price:     price.Price,
					ProductID: price.ProductID,
					Interval:  price.Interval,
				})
			}
		}

		featureData := sc.GetProductFeatures(productData.ID)
		features := []Feature{}

		for _, feature := range featureData {
			if feature.FeatureID == "" {
				continue
			}

			features = append(features, Feature{
				ID:               feature.FeatureID,
				ProductFeatureID: feature.ProductFeatureID,
				Name:             feature.Name,
				Lookupkey:        feature.Lookupkey,
			})
		}

		products = append(products, Product{
			ID:          productData.ID,
			Name:        productData.Name,
			Description: productData.Description,
			Prices:      prices,
			Features:    features,
		})
	}

	return products
}

// GetProductFeatures retrieves all product features from stripe which are active and maps them into a []ProductFeature struct
func (sc *StripeClient) GetProductFeatures(productID string) []ProductFeature {
	productfeatures := []ProductFeature{
		{
			ProductID: productID,
		},
	}

	list := sc.Client.ProductFeatures.List(&stripe.ProductFeatureListParams{
		Product: stripe.String(productID),
	})

	for list.Next() {
		if list.ProductFeature().ID != "" {
			productfeatures = append(productfeatures, ProductFeature{
				ProductFeatureID: list.ProductFeature().ID,
				FeatureID:        list.ProductFeature().EntitlementFeature.ID,
				Name:             list.ProductFeature().EntitlementFeature.Name,
				Lookupkey:        list.ProductFeature().EntitlementFeature.LookupKey,
			})
		}
	}

	return productfeatures
}

// GetPrices retrieves all prices from stripe which are active and maps them into a []Price struct
func (sc *StripeClient) GetPrices() []Price {
	priceParams := &stripe.PriceListParams{}

	iter := sc.Client.Prices.List(priceParams)
	prices := []Price{}

	for iter.Next() {
		priceData := iter.Price()
		if priceData.Product == nil {
			continue
		}

		prices = append(prices, Price{
			ID:        priceData.ID,
			Price:     float64(priceData.UnitAmount) / 100, // nolint:mnd
			ProductID: priceData.Product.ID,
			Interval:  string(priceData.Recurring.Interval),
		})
	}

	return prices
}

// WritePlansToYAML writes the []Product information into a YAML file.
func WritePlansToYAML(product []Product, filename string) error {
	// Marshal the []Product information into YAML
	data, err := yaml.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal plans to YAML: %w", err)
	}

	// Write the YAML data to a file
	err = os.WriteFile(filename, data, 0600) // nolint:mnd
	if err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}

	return nil
}
