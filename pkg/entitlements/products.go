package entitlements

import "github.com/stripe/stripe-go/v81"

// GetProductByID gets a product by ID
func (sc *StripeClient) GetProductByID(id string) (*stripe.Product, error) {
	product, err := sc.Client.Products.Get(id, &stripe.ProductParams{})
	if err != nil {
		return nil, err
	}

	return product, nil
}

// CreateProduct creates a new product in Stripe
func (sc *StripeClient) CreateProduct(name, description string) (*stripe.Product, error) {
	params := &stripe.ProductParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
	}

	return sc.Client.Products.New(params)
}

// ListProducts lists all products in Stripe
func (sc *StripeClient) ListProducts() ([]*stripe.Product, error) {
	var products []*stripe.Product

	params := &stripe.ProductListParams{}
	i := sc.Client.Products.List(params)

	for i.Next() {
		p := i.Product()
		products = append(products, p)
	}

	if err := i.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// GetProduct retrieves a product by its ID
func (sc *StripeClient) GetProduct(productID string) (*stripe.Product, error) {
	return sc.Client.Products.Get(productID, nil)
}

// UpdateProduct updates a product in Stripe
func (sc *StripeClient) UpdateProduct(productID, name, description string) (*stripe.Product, error) {
	params := &stripe.ProductParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
	}

	return sc.Client.Products.Update(productID, params)
}

// DeleteProduct deletes a product in Stripe
func (sc *StripeClient) DeleteProduct(productID string) (*stripe.Product, error) {
	return sc.Client.Products.Del(productID, nil)
}

// GetAllProductPricesMapped retrieves all products and prices from stripe which are active
func (sc *StripeClient) GetAllProductPricesMapped() []Product {
	productParams := &stripe.ProductListParams{}
	productParams.Filters.AddFilter("active", "", "true")

	iter := sc.Client.Products.List(productParams)
	products := []Product{}

	for iter.Next() {
		productData := iter.Product()
		if productData.DefaultPrice == nil {
			continue
		}

		priceData := sc.GetPricesMapped()
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

		featureData := sc.GetFeaturesByProductID(productData.ID)
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
			ID:          productData.ID,
			Name:        productData.Name,
			Description: productData.Description,
			Prices:      prices,
			Features:    features,
		})
	}

	return products
}

// GetFeaturesByProductID retrieves all product features from stripe which are active and maps them into a []ProductFeature struct
func (sc *StripeClient) GetFeaturesByProductID(productID string) []ProductFeature {
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
				FeatureID: list.ProductFeature().EntitlementFeature.ID,
				Name:      list.ProductFeature().EntitlementFeature.Name,
				Lookupkey: list.ProductFeature().EntitlementFeature.LookupKey,
			})
		}
	}

	return productfeatures
}
