package entitlements

import "github.com/stripe/stripe-go/v81"

// NOTE: these are incomplete and are only meant to be used as a reference currently - they will be updated as the project progresses
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
