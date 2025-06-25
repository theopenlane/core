package entitlements

import (
	"context"

	"github.com/stripe/stripe-go/v82"
)

// GetProductByID gets a product by ID
func (sc *StripeClient) GetProductByID(ctx context.Context, id string) (*stripe.Product, error) {
	return sc.Client.V1Products.Retrieve(ctx, id, &stripe.ProductRetrieveParams{})
}

// CreateProduct creates a new product in Stripe
func (sc *StripeClient) CreateProduct(ctx context.Context, name, description string) (*stripe.Product, error) {
	params := &stripe.ProductCreateParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
	}

	return sc.Client.V1Products.Create(ctx, params)
}

// ListProducts lists all products in Stripe
func (sc *StripeClient) ListProducts(ctx context.Context) (products []*stripe.Product, err error) {
	params := &stripe.ProductListParams{}
	result := sc.Client.V1Products.List(ctx, params)

	if result == nil {
		return nil, ErrProductListFailed
	}

	for product, err := range result {
		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	return
}

// GetProduct retrieves a product by its ID
func (sc *StripeClient) GetProduct(ctx context.Context, productID string) (*stripe.Product, error) {
	return sc.Client.V1Products.Retrieve(ctx, productID, nil)
}

// UpdateProduct updates a product in Stripe
func (sc *StripeClient) UpdateProduct(ctx context.Context, productID, name, description string) (*stripe.Product, error) {
	params := &stripe.ProductUpdateParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
	}

	return sc.Client.V1Products.Update(ctx, productID, params)
}

// DeleteProduct deletes a product in Stripe
func (sc *StripeClient) DeleteProduct(ctx context.Context, productID string) (*stripe.Product, error) {
	return sc.Client.V1Products.Delete(ctx, productID, nil)
}

// GetAllProductPricesMapped retrieves all products and prices from stripe which are active
func (sc *StripeClient) GetAllProductPricesMapped(ctx context.Context) (products []Product) {
	productParams := &stripe.ProductListParams{}
	productParams.Filters.AddFilter("active", "", "true")

	result := sc.Client.V1Products.List(ctx, productParams)

	if result == nil {
		return
	}

	for product, err := range result {
		if err != nil {
			continue
		}

		if product.DefaultPrice == nil {
			continue
		}

		priceData := sc.GetPricesMapped(ctx)
		prices := []Price{}

		for _, price := range priceData {
			if price.ProductID == product.ID {
				prices = append(prices, Price{
					ID:        price.ID,
					Price:     price.Price,
					ProductID: price.ProductID,
					Interval:  price.Interval,
				})
			}
		}

		featureData := sc.GetFeaturesByProductID(ctx, product.ID)
		features := []Feature{}

		for _, feature := range featureData {
			if feature.FeatureID == "" {
				continue
			}

			features = append(features, Feature{
				ID:        feature.FeatureID,
				Name:      feature.Name,
				Lookupkey: feature.Lookupkey,
			})
		}

		products = append(products, Product{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Prices:      prices,
			Features:    features,
		})
	}

	return products
}

// GetFeaturesByProductID retrieves all product features from stripe which are active and maps them into a []ProductFeature struct
func (sc *StripeClient) GetFeaturesByProductID(ctx context.Context, productID string) []ProductFeature {
	productfeatures := []ProductFeature{
		{
			ProductID: productID,
		},
	}

	result := sc.Client.V1ProductFeatures.List(ctx, &stripe.ProductFeatureListParams{
		Product: stripe.String(productID),
	})

	for feature, err := range result {
		if err != nil {
			continue
		}

		if feature.ID == "" {
			continue
		}

		productfeatures = append(productfeatures, ProductFeature{
			FeatureID: feature.EntitlementFeature.ID,
			Name:      feature.EntitlementFeature.Name,
			Lookupkey: feature.EntitlementFeature.LookupKey,
		})
	}

	return productfeatures
}
