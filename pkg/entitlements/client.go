package entitlements

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/client"
	"github.com/stripe/stripe-go/v81/customer"
	subs "github.com/stripe/stripe-go/v81/subscription"
	"gopkg.in/yaml.v2"
)

// StripeClientInterface defines the interface for the Stripe client
type StripeClientInterface interface {
	CreateCustomer(params *stripe.CustomerParams) (*stripe.Customer, error)
	GetCustomer(id string, params *stripe.CustomerParams) (*stripe.Customer, error)
	UpdateCustomer(id string, params *stripe.CustomerParams) (*stripe.Customer, error)
	ListCustomers(params *stripe.CustomerListParams) *customer.Iter
	CreateSubscription(params *stripe.SubscriptionParams) (*stripe.Subscription, error)
	GetSubscription(id string, params *stripe.SubscriptionParams) (*stripe.Subscription, error)
	UpdateSubscription(id string, params *stripe.SubscriptionParams) (*stripe.Subscription, error)
	ListSubscriptions(params *stripe.SubscriptionListParams) *subs.Iter
}

// StripeClient is a client for the Stripe API
type StripeClient struct {
	// Client is the Stripe client and is used for accessing all subsequent stripe objects, e.g. products, prices, etc.
	Client *client.API
	// apikey is the Stripe API key
	apikey string
	// Customer is a ref to a generic Customer struct used to wrap Stripe customer and Openlane Organizations (typically)
	Cust Customer
	// Plans are a nomenclature for the recurring context that holds the payment information and is synonymous with Stripe subscriptions
	Plan Plan
	// Product is a stripe product; also know as a "tier"
	Product Product
	// Price holds the interval and amount to be billed
	Price Price
}

// Customer holds the customer information
type Customer struct {
	ID             string `json:"customer_id" yaml:"customer_id"`
	Email          string `json:"email" yaml:"email"`
	Phone          string `json:"phone" yaml:"phone"`
	Address        string `json:"address" yaml:"address"`
	Plans          []Plan `json:"plans" yaml:"plans"`
	StripeParams   *stripe.CustomerParams
	StripeCustomer []stripe.Customer
}

// Plan is the recurring context that holds the payment information
type Plan struct {
	ID                 string `json:"plan_id" yaml:"plan_id"`
	Product            string `json:"product_id" yaml:"product_id"`
	Price              string `json:"price_id" yaml:"price_id"`
	StartDate          int64  `json:"start_date" yaml:"start_date"`
	EndDate            int64  `json:"end_date" yaml:"end_date"`
	StripeParams       *stripe.SubscriptionParams
	StripeSubscription []stripe.Subscription
	Products           []Product
	Features           []Feature
	TrialEnd           int64
	Status             string
}

// Plan is the recurring context that holds the payment information
type Subscription struct {
	ID                 string `json:"plan_id" yaml:"plan_id"`
	Product            string `json:"product_id" yaml:"product_id"`
	Price              string `json:"price_id" yaml:"price_id"`
	StartDate          int64  `json:"start_date" yaml:"start_date"`
	EndDate            int64  `json:"end_date" yaml:"end_date"`
	StripeParams       *stripe.SubscriptionParams
	StripeSubscription []stripe.Subscription
	StripeProduct      []stripe.Product
	StripeFeature      []stripe.ProductFeature
	Products           []Product
	Features           []Feature
	Prices             []Price
	TrialEnd           int64
	Status             string
	StripeCustomerID   string
	OrganizationID     string
	DaysUntilDue       int64
}

// Product holds what we'd more commply call a "tier"
type Product struct {
	ID            string                `json:"product_id" yaml:"product_id"`
	Name          string                `json:"name" yaml:"name"`
	Description   string                `json:"description" yaml:"description"`
	Features      []Feature             `json:"features" yaml:"features"`
	Prices        []Price               `json:"prices" yaml:"prices"`
	StripeParams  *stripe.ProductParams `json:"stripe_params,omitempty" yaml:"stripe_params,omitempty"`
	StripeProduct []stripe.Product      `json:"stripe_product,omitempty" yaml:"stripe_product,omitempty"`
}

// Price holds stripe price params and the associated Product
type Price struct {
	ID           string              `json:"price_id" yaml:"price_id"`
	Price        float64             `json:"price" yaml:"price"`
	ProductID    string              `json:"product_id" yaml:"product_id"`
	Interval     string              `json:"interval" yaml:"interval"`
	StripeParams *stripe.PriceParams `json:"stripe_params,omitempty" yaml:"stripe_params,omitempty"`
	StripePrice  []stripe.Price      `json:"stripe_price,omitempty" yaml:"stripe_price,omitempty"`
}

// Checkout holds the checkout information
type Checkout struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// Feature are part of a product
type Feature struct {
	ID               string `json:"id" yaml:"id"`
	Name             string `json:"name" yaml:"name"`
	Lookupkey        string `json:"lookupkey" yaml:"lookupkey"`
	ProductFeatureID string `json:"product_feature_id" yaml:"product_feature_id"`
}

type ProductFeature struct {
	ProductFeatureID string `json:"product_feature_id" yaml:"product_feature_id"`
	FeatureID        string `json:"feature_id" yaml:"feature_id"`
	ProductID        string `json:"product_id" yaml:"product_id"`
	Name             string `json:"name" yaml:"name"`
	Lookupkey        string `json:"lookupkey" yaml:"lookupkey"`
}

// NewStripeClient creates a new Stripe client
func NewStripeClient(opts ...StripeOptions) *StripeClient {
	sc := &StripeClient{}
	for _, opt := range opts {
		opt(sc)
	}

	sc.Client = client.New(sc.apikey, nil)

	return sc
}

// StripeOptions is a type for setting options on the Stripe client
type StripeOptions func(*StripeClient)

// WithAPIKey sets the API key for the Stripe client
func WithAPIKey(apiKey string) StripeOptions {
	return func(sc *StripeClient) {
		sc.apikey = apiKey
	}
}

// CreateCustomer creates a new customer
func (sc *StripeClient) CreateCustomer(email string) (*stripe.Customer, error) {
	customer, err := sc.Client.Customers.New(&stripe.CustomerParams{Email: &email})
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// GetCustomerByID gets a customer by ID
func (sc *StripeClient) GetCustomerByID(id string) (*stripe.Customer, error) {
	customer, err := sc.Client.Customers.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// UpdateCustomer updates a customer
func (sc *StripeClient) UpdateCustomer(id string, params *stripe.CustomerParams) (*stripe.Customer, error) {
	customer, err := sc.Client.Customers.Update(id, params)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// DeleteCustomer deletes a customer
func (sc *StripeClient) DeleteCustomer(id string) error {
	_, err := sc.Client.Customers.Del(id, nil)
	if err != nil {
		return err
	}

	return nil
}

// CreateSubscription creates a new subscription
func (sc *StripeClient) CreateSubscription(params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	subscription, err := sc.Client.Subscriptions.New(params)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// ListStripeSubscriptions lists stripe subscriptions by customer
func (sc *StripeClient) ListOrCreateStripeSubscriptions(customerID string) (*Subscription, error) {
	i := sc.Client.Subscriptions.List(&stripe.SubscriptionListParams{
		Customer: stripe.String(customerID),
	})

	subs := &Subscription{}

	if !i.Next() {
		sub, err := sc.CreateTrialSubscription(customerID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create trial subscription")
			return nil, err
		}

		return sub, nil
	}

	// assumes customer can only have 1 subscription if there are any
	subs = sc.mapStripeSubscription(i.Subscription())

	return subs, nil
}

// GetSubscriptionByID gets a subscription by ID
func (sc *StripeClient) GetSubscriptionByID(id string) (*stripe.Subscription, error) {
	subscription, err := sc.Client.Subscriptions.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// UpdateSubscription updates a subscription
func (sc *StripeClient) UpdateSubscription(id string, params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	subscription, err := sc.Client.Subscriptions.Update(id, params)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// CancelSubscription cancels a subscription
func (sc *StripeClient) CancelSubscription(id string, params *stripe.SubscriptionCancelParams) (*stripe.Subscription, error) {
	subscription, err := sc.Client.Subscriptions.Cancel(id, params)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

var trialdays int64 = 30

type Subs struct {
	SubsID string
	Prices []Price
}

// CreateTrialSubscription creates a trial subscription
func (sc *StripeClient) CreateTrialSubscription(customerID string) (*Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String("price_1QKLyeBvxky1R7SvaZYGWyQb"),
			},
		},
		TrialPeriodDays: stripe.Int64(trialdays),
		PaymentSettings: &stripe.SubscriptionPaymentSettingsParams{
			SaveDefaultPaymentMethod: stripe.String(string(stripe.SubscriptionPaymentSettingsSaveDefaultPaymentMethodOnSubscription)),
		},
		CollectionMethod: stripe.String(string(stripe.SubscriptionCollectionMethodChargeAutomatically)),
		TrialSettings: &stripe.SubscriptionTrialSettingsParams{
			EndBehavior: &stripe.SubscriptionTrialSettingsEndBehaviorParams{
				MissingPaymentMethod: stripe.String(string(stripe.SubscriptionTrialSettingsEndBehaviorMissingPaymentMethodPause)),
			},
		},
	}

	subs, err := sc.CreateSubscription(params)
	if err != nil {
		log.Err(err).Msg("Failed to create trial subscription")
	}

	log.Info().Msgf("Created trial subscription with ID: %s", subs.ID)

	mappedsubscription := sc.mapStripeSubscription(subs)

	return mappedsubscription, nil
}

func (sc *StripeClient) mapStripeSubscription(subs *stripe.Subscription) *Subscription {
	subscript := Subscription{}

	var prices []Price

	for _, item := range subs.Items.Data {
		prices = append(prices, Price{
			ID:        item.Price.ID,
			Price:     float64(item.Price.UnitAmount),
			ProductID: item.Price.Product.ID,
			Interval:  string(item.Price.Recurring.Interval),
		})
	}

	for _, product := range prices {
		prodFeat := sc.GetProductFeatures(product.ProductID)
		for _, feature := range prodFeat {
			featureList := []Feature{
				{
					ID:               feature.FeatureID,
					ProductFeatureID: feature.ProductFeatureID,
					Name:             feature.Name,
					Lookupkey:        feature.Lookupkey,
				},
			}

			subscript.Features = append(subscript.Features, featureList...)
		}
	}

	subscription := &Subscription{
		ID:               subs.ID,
		Prices:           prices,
		StartDate:        subs.CurrentPeriodStart,
		EndDate:          subs.CurrentPeriodEnd,
		TrialEnd:         subs.TrialEnd,
		Status:           string(subs.Status),
		StripeCustomerID: subs.Customer.ID,
		OrganizationID:   subs.Metadata["organization_id"],
		DaysUntilDue:     subs.DaysUntilDue,
		Features:         subscript.Features,
	}

	return subscription
}

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

// CreateCheckoutSession creates a new checkout session for the customer portal and given product and price
func (sc *StripeClient) CreateBillingPortalUpdateSession(subsID, custID string) (Checkout, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  &custID,
		ReturnURL: stripe.String("http://localhost:3001/organization-settings/billing-usage/pricing"),
		FlowData: &stripe.BillingPortalSessionFlowDataParams{
			Type: stripe.String("subscription_update"),
			SubscriptionUpdate: &stripe.BillingPortalSessionFlowDataSubscriptionUpdateParams{
				Subscription: &subsID,
			},
		},
	}

	billingPortalSession, err := sc.Client.BillingPortalSessions.New(params)
	if err != nil {
		return Checkout{}, err
	}

	return Checkout{
		ID:  billingPortalSession.ID,
		URL: billingPortalSession.URL,
	}, nil
}

// WritePlansToYAML writes the []Product information into a YAML file.
func WritePlansToYAML(product []Product, filename string) error {
	// Marshal the []Product information into YAML
	data, err := yaml.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal plans to YAML: %w", err)
	}

	// Write the YAML data to a file
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}

	return nil
}
